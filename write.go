package iqlog

import (
	"bytes"
	"io"
	"sync"
)

// Global buffer pool for all logging calls
var asyncBufferPool = sync.Pool{
	New: func() any {
		// Preallocate a bytes.Buffer with capacity
		buf := bytes.NewBuffer(make([]byte, 0, maxLineLen))
		return buf
	},
}

// asyncItem is what travels down the async writer channel: either a log
// buffer, or a flush marker (buf == nil) whose done channel is closed once
// every previously queued buffer has been written.
type asyncItem struct {
	buf  *bytes.Buffer
	done chan struct{}
}

type asyncWriter struct {
	ch  chan asyncItem
	wg  sync.WaitGroup
	out io.Writer

	// mu guards closed so Write/Flush never send on a closed channel
	mu     sync.RWMutex
	closed bool
}

func Flush() {
	GlobalLogger.Flush()
}

func (l *logger) Flush() {
	l.mu.Lock()
	out := l.out
	l.mu.Unlock()

	if asyncWr, ok := out.(*asyncWriter); ok {
		asyncWr.Flush()
	}
}

// spawn 1 writer goroutine
func newAsyncWriter(out io.Writer, bufferCount int) *asyncWriter {
	aw := &asyncWriter{
		ch:  make(chan asyncItem, bufferCount),
		out: out,
	}
	aw.wg.Add(1)
	go aw.loop()
	return aw
}

// single goroutine that reads from the channel and writes
func (aw *asyncWriter) loop() {
	defer aw.wg.Done()
	for item := range aw.ch {
		if item.buf == nil {
			// flush marker: everything queued before it has been written
			close(item.done)
			continue
		}
		// Finally, do the I/O
		aw.out.Write(item.buf.Bytes())
		// Put buffer back for reuse
		asyncBufferPool.Put(item.buf)
	}
}

// Write passes data straight to our channel
// fallback in case we want the logger to implement io.Writer as well
func (aw *asyncWriter) Write(p []byte) (n int, err error) {
	buf := asyncBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Write(p)

	aw.mu.RLock()
	if aw.closed {
		aw.mu.RUnlock()
		asyncBufferPool.Put(buf)
		// writer has been closed: fall back to a synchronous write so the
		// line is not lost
		return aw.out.Write(p)
	}
	aw.ch <- asyncItem{buf: buf}
	aw.mu.RUnlock()
	return len(p), nil
}

// Flush blocks until every buffer queued before the call has been written.
// Unlike Close, the writer remains usable afterwards.
func (aw *asyncWriter) Flush() {
	done := make(chan struct{})

	aw.mu.RLock()
	if aw.closed {
		aw.mu.RUnlock()
		return
	}
	aw.ch <- asyncItem{done: done}
	aw.mu.RUnlock()

	<-done
}

// Close drains all pending buffers, stops the writer goroutine, and makes
// subsequent Writes fall back to synchronous writes. Safe to call twice.
func (aw *asyncWriter) Close() error {
	aw.mu.Lock()
	if aw.closed {
		aw.mu.Unlock()
		return nil
	}
	aw.closed = true
	close(aw.ch)
	aw.mu.Unlock()
	aw.wg.Wait()
	return nil
}

// Submits buffer to the channel for async writing
func (aw *asyncWriter) WriteBuffer(buf *bytes.Buffer) {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	if aw.closed {
		aw.out.Write(buf.Bytes())
		return
	}
	aw.ch <- asyncItem{buf: buf}
}
