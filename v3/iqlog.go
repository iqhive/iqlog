package iqlog

import (
	"context"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"

	terminal "golang.org/x/term"
)

// GlobalLogger is used by the Global Logging functions
var GlobalLogger *logger = NewGlobalIQLogger()
var useColour = terminal.IsTerminal(int(os.Stderr.Fd())) && (runtime.GOOS != "windows")

const maxLineLen = 256

type logger struct {
	ctx context.Context
	err error

	out io.Writer

	Level           Level
	jsonMode        bool
	applicationName string
	syslogHost      string

	logFields LogFields

	IncludeTime     bool
	TimestampFormat TimestampFormat
	captureCallers  bool
	newLine         bool
	mu              sync.Mutex

	// A sync.Pool to handle re-usable buffers to reduce allocations.
	// bufferPool     sync.Pool
	// fixedSlicePool sync.Pool
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
	l.SetCaptureCallers(true)

	return l
}

func NewIQLogger(jsonMode bool) *logger {
	logger := &logger{
		logFields:       NewLogFields(),
		captureCallers:  false,
		IncludeTime:     true,
		TimestampFormat: TimestampFormatRFC3339Milli,
		newLine:         true,
		out:             io.Discard,
		Level:           LevelInfo,
	}
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
	SetApplicationName(applicationName)
	SetDebugMode(debugMode)
	SetCaptureCallers(true)
	SetUseColour(true)
	SetNewLine(true)
	SetSyslogHost(syslogHost)
	if GlobalLogger == nil {
		GlobalLogger = NewGlobalIQLogger()
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// Implementation is a no-op here:
func (l *logger) WithGroup(name string) *logger {
	return l
}

func (l *logger) copy() *logger {
	nl := *l
	// nl.baseRecord = l.baseRecord
	return &nl
}

// Add a context to the log entry.
func WithContext(ctx context.Context) *logger {
	if GlobalLogger == nil {
		return NewGlobalIQLogger()
	} else {
		nl := GlobalLogger
		nl.ctx = ctx
		return nl
	}
}

// func (l *logger) HandleMsg(level Level, msg string, args ...interface{}) {
// 	lb := &preallocLine{
// 		logger: l,
// 	}
// 	if l.jsonMode {
// 		lb.output[0] = '{'
// 		lb.bytesUsed++
// 	}
// 	lb.AddTime()

// 	if lb.logger.jsonMode {
// 		lb.writeFinalJSON(msg, args...)
// 	} else {
// 		lb.writeFinalConsole(msg, args...)
// 	}
// }

// func (l *logger) HandleMsgf(level Level, format string, args ...interface{}) {
// 	lb := &preallocLine{
// 		logger: l,
// 	}
// 	if l.jsonMode {
// 		lb.output[0] = '{'
// 		lb.bytesUsed++
// 	}
// 	lb.AddTime()

// 	// TODO: use fmt.Appendf ?
// 	str := fmt.Sprintf(format, args...)

// 	if lb.logger.jsonMode {
// 		lb.writeFinalJSON(str)
// 	} else {
// 		lb.writeFinalConsole(str)
// 	}
// }
