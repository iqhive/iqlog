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

type asyncWriter struct {
	ch  chan *bytes.Buffer
	wg  sync.WaitGroup
	out io.Writer
}

func Flush() {
	GlobalLogger.Flush()
}

func (l *logger) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if asyncWr, ok := l.out.(*asyncWriter); ok {
		// close asyncWriter to flush all buffers
		asyncWr.Close()
		// reopen to continue logging
		l.out = newAsyncWriter(l.out, cap(asyncWr.ch))
	}
}

// spawn 1 writer goroutine
func newAsyncWriter(out io.Writer, bufferCount int) *asyncWriter {
	aw := &asyncWriter{
		ch:  make(chan *bytes.Buffer, bufferCount),
		out: out,
	}
	aw.wg.Add(1)
	go aw.loop()
	return aw
}

// single goroutine that reads from the channel and writes
func (aw *asyncWriter) loop() {
	defer aw.wg.Done()
	for buf := range aw.ch {
		// Finally, do the I/O
		aw.out.Write(buf.Bytes())
		// Put buffer back for reuse
		asyncBufferPool.Put(buf)
	}
}

// Write passes data straight to our channel
// fallback in case we want the logger to implement io.Writer as well
func (aw *asyncWriter) Write(p []byte) (n int, err error) {
	buf := asyncBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Write(p)
	aw.ch <- buf
	return len(p), nil
}

// closes the channel and waits for the loop to finish
func (aw *asyncWriter) Close() error {
	close(aw.ch)
	aw.wg.Wait()
	return nil
}

// Submits buffer to the channel for async writing
func (aw *asyncWriter) WriteBuffer(buf *bytes.Buffer) {
	aw.ch <- buf
}
