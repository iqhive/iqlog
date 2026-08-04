//go:build windows

package iqlog

import (
	"io"
	"os"
	"sync"

	"github.com/iqhive/iqlog/ringbuffer"
)

type ringWriter struct {
	ringBuffer *ringbuffer.RingBuffer[[]byte]
	writer     io.Writer
	// writeMu serializes writes to writer between the consumer goroutine
	// and the synchronous fallback path in Write
	writeMu sync.Mutex
	// closedMu guards closed so Write never enqueues after Close
	closedMu sync.RWMutex
	closed   bool
	wg       sync.WaitGroup
}

func (rw *ringWriter) Start() {
	rw.wg.Add(1)
	go func() {
		defer rw.wg.Done()
		for {
			val, ok := rw.ringBuffer.DequeueBlocking()
			if !ok {
				// ring closed and fully drained
				return
			}
			rw.writeMu.Lock()
			rw.writer.Write(val)
			rw.writeMu.Unlock()
		}
	}()
}

func (rw *ringWriter) Write(p []byte) (n int, err error) {
	rw.closedMu.RLock()
	if rw.closed {
		rw.closedMu.RUnlock()
		// writer has been closed: write synchronously so the line is not lost
		rw.writeMu.Lock()
		defer rw.writeMu.Unlock()
		return rw.writer.Write(p)
	}
	// Copy p because callers (e.g. pooled line buffers) may reuse the
	// underlying array before the consumer goroutine writes it out.
	c := make([]byte, len(p))
	copy(c, p)
	ok := rw.ringBuffer.Enqueue(c)
	rw.closedMu.RUnlock()
	if !ok {
		// ring is full: write synchronously rather than silently dropping
		rw.writeMu.Lock()
		defer rw.writeMu.Unlock()
		return rw.writer.Write(p)
	}
	return len(p), nil
}

// Close drains all queued lines, stops the consumer goroutine, and makes
// subsequent Writes synchronous. Safe to call twice.
func (rw *ringWriter) Close() error {
	rw.closedMu.Lock()
	if rw.closed {
		rw.closedMu.Unlock()
		return nil
	}
	rw.closed = true
	rw.ringBuffer.Close()
	rw.closedMu.Unlock()
	rw.wg.Wait()
	return nil
}

// replaceWriter swaps the logger output under the logger mutex and closes
// the previous writer when it is one of our wrapper types, so its goroutine
// exits and queued lines are drained instead of leaking.
func (l *logger) replaceWriter(w io.Writer) {
	l.mu.Lock()
	old := l.out
	l.out = w
	l.mu.Unlock()
	if old == w {
		return
	}
	switch ow := old.(type) {
	case *asyncWriter:
		_ = ow.Close()
	case *ringWriter:
		_ = ow.Close()
	}
}

func (l *logger) SetWriter(w io.Writer) {
	// option 1 - plain old writer
	l.replaceWriter(w)
}

func (l *logger) SetAsyncWriter(w io.Writer) {
	// option 2 - async writer with a background flusher goroutine
	l.replaceWriter(newAsyncWriter(w, 1000))
}

func (l *logger) SetRingbufferWriter(w io.Writer) {
	// option 3 - ring writer, which should be better for non-stop loggings
	rw := &ringWriter{
		ringBuffer: ringbuffer.NewRingBuffer[[]byte](10000),
		writer:     w,
	}

	rw.Start()

	l.replaceWriter(rw)
}

func (l *logger) GetWriter() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}

func SetWriter(w io.Writer) {
	GlobalLogger.SetWriter(w)
}

func GetWriter() io.Writer {
	return GlobalLogger.GetWriter()
}

// SetApplicationName sets the application name on the GlobalLogger
func SetApplicationName(applicationName string) {
	GlobalLogger.SetApplicationName(applicationName)
}
func (l *logger) SetApplicationName(name string) {
	l.mu.Lock()
	l.applicationName = name
	l.mu.Unlock()
}

// SetSyslogHost sets the syslog host on the GlobalLogger
func SetSyslogHost(host string) {
	GlobalLogger.SetSyslogHost(host)
}

// SetDebugMode sets the syslog host on the GlobalLogger
func SetDebugMode(debugMode bool) {
	GlobalLogger.SetDebugMode(debugMode)
}
func (l *logger) SetDebugMode(d bool) {
	l.mu.Lock()
	if d {
		l.SetLevel(LevelDebug)
	} else {
		l.SetLevel(LevelInfo)
	}
	l.mu.Unlock()
}

// SetNewLine sets the new line on the GlobalLogger
func SetNewLine(newLine bool) {
	GlobalLogger.SetNewLine(newLine)
}
func (l *logger) SetNewLine(d bool) {
	l.mu.Lock()
	l.newLine.Store(d)
	l.mu.Unlock()
}

// SetCallerDepth sets the capture callers on the GlobalLogger
func SetCallerDepth(captureCaller int) {
	GlobalLogger.SetCallerDepth(captureCaller)
}
func (l *logger) SetCallerDepth(d int) {
	l.mu.Lock()
	l.CallerDepth.Store(int32(d))
	l.mu.Unlock()
}

func SetUseColour(enabled bool) {
	useColour.Store(enabled)
}
func (l *logger) SetUseColour(d bool) { useColour.Store(d) }

// SetJSONMode sets the JSON mode on the GlobalLogger
func SetJSONMode(jsonMode bool) {
	GlobalLogger.SetJSONMode(jsonMode)
}
func (l *logger) SetJSONMode(isJSONmode bool) {
	l.mu.Lock()
	l.jsonMode.Store(isJSONmode)
	l.newLine.Store(true)
	l.IncludeTime.Store(isJSONmode)
	l.mu.Unlock()
}

// SetSyslogHost is a no-op on Windows since syslog is not available
func (l *logger) SetSyslogHost(newhost string) {
	l.mu.Lock()
	l.syslogHost = newhost
	l.mu.Unlock()
	if newhost != "" {
		l.Warnf("Syslog is not supported on Windows. Host %s will be ignored. Output remains on stderr.", newhost)
	} else {
		l.Info("Log output set to StdErr")
	}
	l.replaceWriter(os.Stderr)
}
