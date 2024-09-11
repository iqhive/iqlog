package iqlog

import (
	"bytes"
	"io"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
var jsonBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
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
		// Close the asyncWriter to flush all buffers
		asyncWr.Close()
		// Reopen the asyncWriter to continue logging
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

func (aw *asyncWriter) loop() {
	defer aw.wg.Done()
	for buf := range aw.ch {
		aw.out.Write(buf.Bytes())
		// Return the buffer to global buffer pool
		bufferPool.Put(buf)
	}
}

// Write passes data straight to our channel. This is a fallback
// in case we want the logger to implement io.Writer as well.
func (aw *asyncWriter) Write(p []byte) (n int, err error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Write(p)
	aw.ch <- buf
	return len(p), nil
}

// Close closes the channel and waits for the loop to finish
func (aw *asyncWriter) Close() error {
	close(aw.ch)
	aw.wg.Wait()
	return nil
}

// reusable buffers
var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func (l *logger) WriteRecordNoAlloc(record LogRecord) error {
	// Prepare a stack-allocated array big enough for time prefix, level, msg, newline, etc.
	var out [512]byte
	used := 0

	// (A) Write optional time prefix (disable or keep if you want it).
	//     If you want no allocations, consider a fixed format:
	if l.IncludeTimePrefix {
		t := record.Time
		// Example: we can just do t.Format(...) into a small buffer
		// But note t.Format in the stdlib can sometimes allocate.
		// A truly no-alloc way is manual: appendInt, etc.
		// For brevity, let's assume we do a small fixed time format manually.
		used += copy(out[used:], t.Format(l.TimePrefixFormat))
		out[used] = ' '
		used++
	}

	// (B) Write out the level text (INFO, etc.) with no color
	used += copy(out[used:], "INFO ")

	// (C) [Optional] write applicationName
	if l.applicationName != "" {
		out[used] = '['
		used++
		used += copy(out[used:], l.applicationName)
		out[used] = ']'
		used++
		out[used] = ' '
		used++
	}

	// (D) Write the actual message
	copy(out[used:], record.Message[:record.MsgLen])
	used += record.MsgLen

	// (E) Add newline
	if l.newLine {
		out[used] = '\n'
		used++
	}

	// Finally, write it:
	_, err := l.out.Write(out[:used])
	return err
}

// func (l *IQLogger) log(level Level, msg string, kvs ...KeyVal) {
// 	// TODO: Skip if the level isn't enabled, etc

// 	// Grab a buffer from the pool
// 	buf := bufPool.Get().(*bytes.Buffer)
// 	defer bufPool.Put(buf)

// 	// Reset so it's empty
// 	buf.Reset()

// 	// TODO: Build the log line directly, e.g. JSON or minimal text

// 	// Actually write to io.Writer
// 	l.writer.Write(buf.Bytes())
// }
