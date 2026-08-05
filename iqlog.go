package iqlog

import (
	"context"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"

	terminal "golang.org/x/term"
)

// GlobalLogger is used by the Global Logging functions
var GlobalLogger *logger = NewGlobalIQLogger()
var useColour = func() *atomic.Bool {
	b := &atomic.Bool{}
	b.Store(terminal.IsTerminal(int(os.Stderr.Fd())) && (runtime.GOOS != "windows"))
	return b
}()

const maxLineLen = 1024
const maxStringLen = 256

type logger struct {
	ctx context.Context
	err error

	out io.Writer

	// level, jsonMode, IncludeTime, newLine, and CallerDepth are accessed
	// atomically so they can be reconfigured while other goroutines log
	level           atomic.Int32
	jsonMode        atomic.Bool
	applicationName string
	syslogHost      string

	// logFields LogFields

	IncludeTime     atomic.Bool
	TimestampFormat TimestampFormat
	newLine         atomic.Bool
	// mu serializes writes to out; it is a pointer so copies of the
	// logger (e.g. from WithContext) share the same lock for a shared writer
	mu *sync.Mutex
	// writeErr records the most recent error returned by the output
	// writer; shared between logger copies and background writers
	writeErr *writeErrBox

	// A sync.Pool to handle re-usable buffers to reduce allocations.
	// bufferPool     sync.Pool
	// fixedSlicePool sync.Pool
	CallerDepth atomic.Int32
}

type TimestampFormat int

const (
	TimestampFormatRFC3339 TimestampFormat = iota
	TimestampFormatRFC3339Milli
	TimestampFormatRFC3339Micro
	TimestampFormatRFC3339Nano
)

// NewIQLogger creates and return a new Logger
func NewGlobalIQLogger() *logger {

	l := NewIQLogger(false)
	l.SetUseColour(terminal.IsTerminal(int(os.Stderr.Fd())) && (runtime.GOOS != "windows"))
	l.SetCallerDepth(1)

	return l
}

func NewIQLogger(jsonMode bool) *logger {
	logger := &logger{
		// logFields:       NewLogFields(),
		TimestampFormat: TimestampFormatRFC3339Milli,
		out:             io.Discard,
		mu:              &sync.Mutex{},
		writeErr:        &writeErrBox{},
	}
	logger.SetCallerDepth(0) // >0 = Capture callers
	logger.IncludeTime.Store(true)
	logger.newLine.Store(true)
	logger.SetLevel(LevelInfo)
	logger.SetWriter(os.Stderr)
	debugStr := os.Getenv("IQLOG_DEBUG")
	if b, _ := strconv.ParseBool(debugStr); b {
		logger.SetDebugMode(true)
	}

	if jsonMode {
		logger.SetJSONMode(true)
	} else {
		logger.SetJSONMode(false)
	}

	// // Global buffer pool for all logging calls
	// logger.bufferPool = sync.Pool{
	// 	New: func() any {
	// 		return bytes.NewBuffer(make([]byte, 0, maxLineLen))
	// 	},
	// }
	// oneBuffer := logger.bufferPool.Get().(*bytes.Buffer)
	// twoBuffer := logger.bufferPool.Get().(*bytes.Buffer)
	// logger.bufferPool.Put(oneBuffer)
	// logger.bufferPool.Put(twoBuffer)

	// // Global buffer pool for all logging calls
	// logger.fixedSlicePool = sync.Pool{
	// 	New: func() any {
	// 		// Preallocate a slice with capacity
	// 		return &[maxLineLen]byte{}
	// 		// s := [maxLineLen]byte{}
	// 		// return &s
	// 	},
	// }
	// oneSlice := logger.fixedSlicePool.Get().(*[maxLineLen]byte)
	// twoSlice := logger.fixedSlicePool.Get().(*[maxLineLen]byte)
	// logger.fixedSlicePool.Put(oneSlice)
	// logger.fixedSlicePool.Put(twoSlice)

	return logger
}

// Init performs all the base configuration of the GlobalLogger
// This function is typically called when the application is starting up
// It sets log/slog and log's default output to use the iqlog format.
func Init(applicationName string, syslogHost string, debugMode bool) {
	if GlobalLogger == nil {
		GlobalLogger = NewGlobalIQLogger()
	}
	SetApplicationName(applicationName)
	SetDebugMode(debugMode)
	SetCallerDepth(1)
	SetUseColour(terminal.IsTerminal(int(os.Stderr.Fd())) && (runtime.GOOS != "windows"))
	SetNewLine(true)
	SetSyslogHost(syslogHost)
}

// writeErrBox holds the most recent writer error behind its own lock so
// background writer goroutines can record errors without touching the
// logger mutex.
type writeErrBox struct {
	mu  sync.Mutex
	err error
}

func (l *logger) writeLocked(line []byte) {
	l.mu.Lock()
	_, err := l.out.Write(line)
	l.mu.Unlock()
	l.recordWriteErr(err)
}

// recordWriteErr stores a writer error for later inspection via
// LastWriteError. Safe to call from background writer goroutines.
func (l *logger) recordWriteErr(err error) {
	if err == nil || l.writeErr == nil {
		return
	}
	l.writeErr.mu.Lock()
	l.writeErr.err = err
	l.writeErr.mu.Unlock()
}

// LastWriteError returns the most recent error returned by the logger's
// output writer, or nil if all writes have succeeded. Errors from async
// and ring-buffer background writes are also recorded here.
func (l *logger) LastWriteError() error {
	if l.writeErr == nil {
		return nil
	}
	l.writeErr.mu.Lock()
	defer l.writeErr.mu.Unlock()
	return l.writeErr.err
}

// LastWriteError returns the most recent write error of the GlobalLogger.
func LastWriteError() error {
	return GlobalLogger.LastWriteError()
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// Implementation is a no-op here:
func (l *logger) WithGroup(name string) *logger {
	return l
}

func (l *logger) copy() *logger {
	// copy fields individually, sharing the mutex pointer so copies
	// serialize writes against the original logger
	nl := &logger{
		ctx:             l.ctx,
		err:             l.err,
		out:             l.out,
		writeErr:        l.writeErr,
		applicationName: l.applicationName,
		syslogHost:      l.syslogHost,
		TimestampFormat: l.TimestampFormat,
		mu:              l.mu,
	}
	nl.jsonMode.Store(l.jsonMode.Load())
	nl.IncludeTime.Store(l.IncludeTime.Load())
	nl.newLine.Store(l.newLine.Load())
	nl.CallerDepth.Store(l.CallerDepth.Load())
	nl.SetLevel(l.Level())
	return nl
}

// Add a context to the log entry.
func WithContext(ctx context.Context) *logger {
	if GlobalLogger == nil {
		return NewGlobalIQLogger()
	}
	nl := GlobalLogger.copy()
	nl.ctx = ctx
	return nl
}

// func (l *logger) HandleMsg(level Level, msg string, args ...interface{}) {
// 	lb := &preallocLine{
// 		logger: l,
// 	}
// 	if l.jsonMode.Load() {
// 		lb.output[0] = '{'
// 		lb.bytesUsed++
// 	}
// 	lb.AddTime()

// 	if lb.logger.jsonMode.Load() {
// 		lb.writeFinalJSON(msg, args...)
// 	} else {
// 		lb.writeFinalConsole(msg, args...)
// 	}
// }

// func (l *logger) HandleMsgf(level Level, format string, args ...interface{}) {
// 	lb := &preallocLine{
// 		logger: l,
// 	}
// 	if l.jsonMode.Load() {
// 		lb.output[0] = '{'
// 		lb.bytesUsed++
// 	}
// 	lb.AddTime()

// 	// TODO: use fmt.Appendf ?
// 	str := fmt.Sprintf(format, args...)

// 	if lb.logger.jsonMode.Load() {
// 		lb.writeFinalJSON(str)
// 	} else {
// 		lb.writeFinalConsole(str)
// 	}
// }
