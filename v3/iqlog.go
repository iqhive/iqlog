package iqlog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	terminal "golang.org/x/term"
)

// GlobalLogger is used by the Global Logging functions
var GlobalLogger *logger = NewGlobalIQLogger()

type logger struct {
	ctx context.Context
	err error

	// slog *slog.Logger
	slog *slogEmu

	out io.Writer

	debug           bool
	jsonMode        bool
	applicationName string
	syslogHost      string

	logFields LogFields

	IncludeTimePrefix bool
	TimePrefixFormat  string
	useColour         bool
	captureCallers    bool
	newLine           bool
	mu                sync.Mutex

	// A sync.Pool to handle re-usable buffers to reduce allocations.
	bufferPool sync.Pool

	// TODO: consider adding a scratch buffer for type-to-string conversions here
	// TODO: benchmark the performance of this vs the current method

	// Instead of storing LogFields, store a LogRecord template
	// for "WithField" calls:
	baseRecord LogRecord
}

// NewIQLogger creates and return a new Logger
func NewGlobalIQLogger() *logger {

	l := NewIQLogger(false)
	l.SetUseColour(terminal.IsTerminal(int(os.Stderr.Fd())) && (runtime.GOOS != "windows"))
	l.SetWriter(os.Stderr)
	l.SetCaptureCallers(true)

	sl := slog.NewTextHandler(l.out, &slog.HandlerOptions{})
	l.slog = &slogEmu{
		Handler: sl,
		Logger:  slog.New(sl),
	}

	return l
}

func NewIQLogger(jsonMode bool) *logger {
	logger := &logger{
		debug:             false,
		logFields:         NewLogFields(),
		captureCallers:    false,
		IncludeTimePrefix: false,
		TimePrefixFormat:  time.StampMicro,
		out:               os.Stderr,
		newLine:           true,
	}
	debugStr := os.Getenv("IQLOG_DEBUG")
	if b, _ := strconv.ParseBool(debugStr); b {
		logger.SetDebugMode(true)
	}

	if jsonMode {
		logger.SetJSONMode(true)
	} else {
		logger.SetJSONMode(false)
		logger.useColour = terminal.IsTerminal(int(os.Stderr.Fd())) && (runtime.GOOS != "windows")
	}

	logger.bufferPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

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
		slog.SetDefault(GlobalLogger.slog.Logger)
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
	nl.baseRecord = l.baseRecord
	return &nl
}

// Add a context to the log entry.
func WithContext(ctx context.Context) *logger {
	if GlobalLogger == nil {
		return &logger{
			ctx:       ctx,
			jsonMode:  false,
			debug:     false,
			out:       os.Stderr,
			useColour: terminal.IsTerminal(int(os.Stderr.Fd())) && (runtime.GOOS != "windows"),
		}
	} else {
		nl := GlobalLogger
		nl.ctx = ctx
		return nl
	}
}
