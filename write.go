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
	native, _ := out.(nativeLogWriter)
	aw := &asyncWriter{ch: make(chan asyncItem, bufferCount), out: out, native: native, onErr: onErr, policy: policy, writeMu: writeMu}
	aw.wg.Add(1)
	go aw.loop()
	return aw
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

func (l *Logger) Flush() error {
	if l.writer.closed.Load() {
		return ErrClosed
	}
	if aw, ok := l.writer.active.Load().out.(*asyncWriter); ok {
		aw.Flush()
	}
	return l.takeWriteError()
}

func (l *Logger) Close() error {
	l.writer.mu.Lock()
	if l.writer.closed.Swap(true) {
		l.writer.mu.Unlock()
		return nil
	}
	old := l.writer.active.Swap(&outputState{out: io.Discard, configured: io.Discard, concurrent: true})
	l.writer.mu.Unlock()
	if err := retireWriter(old.out, nil); err != nil {
		return err
	}
	return l.takeWriteError()
}

// newOutputState builds the output state a writer-replacement method installs.
// It carries the configured ConcurrentWriter setting forward so SetWriter does
// not silently re-serialize a writer the caller declared concurrent, and it
// treats our own async wrapper and io.Discard as concurrent.
func (l *Logger) newOutputState(w io.Writer) *outputState {
	_, async := w.(*asyncWriter)
	return &outputState{
		out:        w,
		configured: w,
		concurrent: async || w == io.Discard || l.config.Load().concurrentWriter,
	}
}
