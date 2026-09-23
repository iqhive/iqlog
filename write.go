package iqlog

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

type asyncItem struct {
	buf   []byte
	done  chan struct{}
	level Level
}

type asyncWriter struct {
	ch     chan asyncItem
	out    io.Writer
	native nativeLogWriter
	onErr  func(error)
	policy OverflowPolicy
	closed atomic.Bool
	sendMu sync.RWMutex
	// writeMu is the logger's shared write mutex, not one of our own: a
	// retiring async writer drains to the same destination the replacement is
	// already writing to, so both -- and any synchronous write -- have to
	// serialize on the same lock.
	writeMu *sync.Mutex
	wg      sync.WaitGroup
}

func newAsyncWriter(out io.Writer, bufferCount int, policy OverflowPolicy, onErr func(error), writeMu *sync.Mutex) *asyncWriter {
	// Never queue in front of another queue. A wrapper handed back in through
	// GetWriter is about to be retired by the caller, and a closed wrapper
	// writes synchronously under the shared write mutex that this writer's
	// loop already holds, which deadlocks the loop. Writing to the real
	// destination also lets retireWriter recognise it as still in use.
	out = unwrapAsync(out)
	native, _ := out.(nativeLogWriter)
	aw := &asyncWriter{ch: make(chan asyncItem, bufferCount), out: out, native: native, onErr: onErr, policy: policy, writeMu: writeMu}
	aw.wg.Add(1)
	go aw.loop()
	return aw
}

// unwrapAsync returns the destination underneath any of this package's own
// queue wrappers.
func unwrapAsync(w io.Writer) io.Writer {
	for {
		aw, ok := w.(*asyncWriter)
		if !ok {
			return w
		}
		w = aw.out
	}
}

func (aw *asyncWriter) loop() {
	defer aw.wg.Done()
	for item := range aw.ch {
		if item.done != nil {
			close(item.done)
			continue
		}
		aw.writeRecovered(item.buf, item.level)
		releaseEventBuffer(item.buf)
	}
}

// writeRecovered turns a panic from the destination into a reported error.
// A synchronous write propagates such a panic to the call site that chose to
// log, but a background write has no call site to propagate to, and an
// unrecovered panic on this goroutine would take the host process down
// because a log sink misbehaved.
func (aw *asyncWriter) writeRecovered(buf []byte, level Level) {
	defer func() {
		if r := recover(); r != nil && aw.onErr != nil {
			aw.onErr(fmt.Errorf("iqlog: log writer panicked: %v", r))
		}
	}()
	aw.write(buf, level)
}

func (aw *asyncWriter) write(buf []byte, level Level) {
	n, err := aw.writeSerialized(buf, level)
	if err == nil && n != len(buf) {
		err = io.ErrShortWrite
	}
	if err != nil && aw.onErr != nil {
		aw.onErr(err)
	}
}

// writeSerialized writes through the shared write mutex. The unlock is
// deferred because the destination is caller-supplied and may panic.
func (aw *asyncWriter) writeSerialized(buf []byte, level Level) (int, error) {
	aw.writeMu.Lock()
	defer aw.writeMu.Unlock()
	if aw.native != nil {
		return aw.native.writeLevel(level, buf)
	}
	return aw.out.Write(buf)
}

// Write satisfies io.Writer. The logging hot path uses WriteOwned to avoid a copy.
func (aw *asyncWriter) Write(p []byte) (int, error) {
	buf := acquireEventBuffer()
	buf = append(buf, p...)
	aw.WriteOwned(buf, LevelUnknown)
	return len(p), nil
}

// WriteOwned transfers ownership of buf to the writer. The buffer must not be
// accessed after this call.
func (aw *asyncWriter) WriteOwned(buf []byte, level Level) {
	// The unlock is deferred because two paths below write to the
	// caller-supplied destination while this lock is held, and a panic there
	// would otherwise leave it read-locked forever, wedging Close. Holding it
	// across those writes is harmless: Close publishes closed under the write
	// lock, so anything that observes it has already let Close through.
	aw.sendMu.RLock()
	defer aw.sendMu.RUnlock()
	if aw.closed.Load() {
		aw.write(buf, level)
		releaseEventBuffer(buf)
		return
	}
	item := asyncItem{buf: buf, level: level}
	switch aw.policy {
	case OverflowDrop:
		select {
		case aw.ch <- item:
		default:
			// Terminal records are never dropped: preserve order by draining
			// the already-accepted records, then write directly (the same
			// fallback OverflowSync uses).
			if level.terminal() {
				done := make(chan struct{})
				aw.ch <- asyncItem{done: done}
				<-done
				aw.write(buf, level)
				releaseEventBuffer(buf)
				return
			}
			releaseEventBuffer(buf)
			if aw.onErr != nil {
				aw.onErr(ErrWriteDropped)
			}
		}
	case OverflowSync:
		select {
		case aw.ch <- item:
		default:
			// Preserve record order: wait for all already accepted records before
			// falling back to a synchronous write.
			done := make(chan struct{})
			aw.ch <- asyncItem{done: done}
			<-done
			aw.write(buf, level)
			releaseEventBuffer(buf)
		}
	default:
		aw.ch <- item
	}
}

func (aw *asyncWriter) Flush() {
	done := make(chan struct{})
	aw.sendMu.RLock()
	if aw.closed.Load() {
		aw.sendMu.RUnlock()
		return
	}
	aw.ch <- asyncItem{done: done}
	aw.sendMu.RUnlock()
	<-done
}

func (aw *asyncWriter) Close() error {
	aw.sendMu.Lock()
	if aw.closed.Swap(true) {
		aw.sendMu.Unlock()
		return nil
	}
	close(aw.ch)
	aw.sendMu.Unlock()
	aw.wg.Wait()
	return nil
}

// Flush waits for every record the writer had accepted when Flush was called.
// The state is read under writer.mu: a reconfiguration drains the writer it
// retires while holding that lock, so records queued on the previous writer
// have been written by the time the new one is observed here.
func (l *Logger) Flush() error {
	l.writer.mu.Lock()
	closed := l.writer.closed.Load()
	state := l.writer.active.Load()
	l.writer.mu.Unlock()
	if closed {
		return ErrClosed
	}
	if aw, ok := state.out.(*asyncWriter); ok {
		aw.Flush()
	}
	return l.takeWriteError()
}

// Close drains and stops the writer. The drain happens under writer.mu so a
// concurrent Flush cannot return before the accepted records are written.
func (l *Logger) Close() error {
	l.writer.mu.Lock()
	if l.writer.closed.Swap(true) {
		l.writer.mu.Unlock()
		return nil
	}
	// Writes go to io.Discard from here on, but Config() keeps reporting the
	// caller-owned destination, which Close does not touch, so a logger can be
	// rebuilt from a closed one's configuration.
	configured := l.writer.active.Load().configured
	old := l.writer.active.Swap(&outputState{out: io.Discard, configured: configured, concurrent: true})
	err := retireWriter(old.out, nil)
	l.writer.mu.Unlock()
	if err != nil {
		return err
	}
	return l.takeWriteError()
}

// newOutputState builds the output state a writer-replacement method installs.
// It carries the configured ConcurrentWriter setting forward so SetWriter does
// not silently re-serialize a writer the caller declared concurrent, and it
// treats our own async wrapper and io.Discard as concurrent. The configured
// writer is the destination underneath our own queue wrapper, as it is for a
// Config-built async logger, so Config().Writer never hands the live wrapper
// back to a SetConfig that would retire it. It intentionally omits native:
// replacing the writer stops native logging, and replaceWriter clears the
// NativeLog flag so the config and active routing never disagree.
func (l *Logger) newOutputState(w io.Writer) *outputState {
	_, async := w.(*asyncWriter)
	return &outputState{
		out:        w,
		configured: unwrapAsync(w),
		concurrent: async || w == io.Discard || l.config.Load().concurrentWriter,
	}
}
