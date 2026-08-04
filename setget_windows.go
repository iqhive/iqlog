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
}

func (rw *ringWriter) Start() {
	go func() {
		for {
			val, ok := rw.ringBuffer.DequeueBlocking()
			if !ok {
				// Nothing available, maybe sleep or continue
				// fmt.Println("Nothing available on ring!")
				// os.Exit(1)
				continue
			}
			rw.writeMu.Lock()
			rw.writer.Write(val)
			rw.writeMu.Unlock()
		}
	}()
}

func (rw *ringWriter) Write(p []byte) (n int, err error) {
	// Copy p because callers (e.g. pooled line buffers) may reuse the
	// underlying array before the consumer goroutine writes it out.
	c := make([]byte, len(p))
	copy(c, p)
	if !rw.ringBuffer.Enqueue(c) {
		// ring is full: write synchronously rather than silently dropping
		rw.writeMu.Lock()
		defer rw.writeMu.Unlock()
		return rw.writer.Write(p)
	}
	return len(p), nil
}

func (l *logger) SetWriter(w io.Writer) {
	// option 1 - plain old writer
	l.out = w
}

func (l *logger) SetAsyncWriter(w io.Writer) {
	// // option 1 - plain old writer
	// l.out = w

	// option 2 - async writer, which should be ok, but its really not
	l.out = newAsyncWriter(w, 1000)
}

func (l *logger) SetRingbufferWriter(w io.Writer) {
	// option 3 - ring writer, which should be better for non-stop loggings, lets see
	rw := &ringWriter{
		ringBuffer: ringbuffer.NewRingBuffer[[]byte](10000),
		writer:     w,
	}

	rw.Start()

	l.out = rw
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
func (l *logger) SetApplicationName(name string) { l.applicationName = name }

// SetSyslogHost sets the syslog host on the GlobalLogger
func SetSyslogHost(host string) {
	GlobalLogger.SetSyslogHost(host)
}

// SetDebugMode sets the syslog host on the GlobalLogger
func SetDebugMode(debugMode bool) {
	GlobalLogger.SetDebugMode(debugMode)
}
func (l *logger) SetDebugMode(d bool) {
	if d {
		l.Level = LevelDebug
	} else {
		l.Level = LevelInfo
	}
}

// SetNewLine sets the new line on the GlobalLogger
func SetNewLine(newLine bool) {
	GlobalLogger.SetNewLine(newLine)
}
func (l *logger) SetNewLine(d bool) {
	l.newLine = d
}

// SetCallerDepth sets the capture callers on the GlobalLogger
func SetCallerDepth(captureCaller int) {
	GlobalLogger.SetCallerDepth(captureCaller)
}
func (l *logger) SetCallerDepth(d int) { l.CallerDepth = d }

func SetUseColour(enabled bool) {
	useColour = enabled
}
func (l *logger) SetUseColour(d bool) { useColour = d }

// SetJSONMode sets the JSON mode on the GlobalLogger
func SetJSONMode(jsonMode bool) {
	GlobalLogger.SetJSONMode(jsonMode)
}
func (l *logger) SetJSONMode(isJSONmode bool) {
	if isJSONmode {
		l.jsonMode = isJSONmode
		l.newLine = true
		l.IncludeTime = true
	} else {
		l.jsonMode = isJSONmode
		l.newLine = true
		l.IncludeTime = false
	}
}

// SetSyslogHost is a no-op on Windows since syslog is not available
func (l *logger) SetSyslogHost(newhost string) {
	l.syslogHost = newhost
	if newhost != "" {
		l.Warnf("Syslog is not supported on Windows. Host %s will be ignored. Output remains on stderr.", newhost)
		l.out = os.Stderr
		return
	}
	if newhost == "" {
		l.Info("Log output set to StdErr")
		l.out = os.Stderr
		return
	}
}
