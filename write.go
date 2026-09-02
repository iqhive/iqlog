package iqlog

import (
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
	ch      chan asyncItem
	out     io.Writer
	native  nativeLogWriter
	onErr   func(error)
	policy  OverflowPolicy
	closed  atomic.Bool
	sendMu  sync.RWMutex
	writeMu sync.Mutex
	wg      sync.WaitGroup
}

func newAsyncWriter(out io.Writer, bufferCount int, policy OverflowPolicy, onErr func(error)) *asyncWriter {
	native, _ := out.(nativeLogWriter)
	aw := &asyncWriter{ch: make(chan asyncItem, bufferCount), out: out, native: native, onErr: onErr, policy: policy}
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
		aw.write(item.buf, item.level)
		releaseEventBuffer(item.buf)
	}
}

func (aw *asyncWriter) write(buf []byte, level Level) {
	aw.writeMu.Lock()
	var n int
	var err error
	if aw.native != nil {
		n, err = aw.native.writeLevel(level, buf)
	} else {
		n, err = aw.out.Write(buf)
	}
	aw.writeMu.Unlock()
	if err == nil && n != len(buf) {
		err = io.ErrShortWrite
	}
	if err != nil && aw.onErr != nil {
		aw.onErr(err)
	}
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
	aw.sendMu.RLock()
	if aw.closed.Load() {
		aw.sendMu.RUnlock()
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
	aw.sendMu.RUnlock()
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
