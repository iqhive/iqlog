package iqlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"log/syslog"
	"os"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// GlobalLogger is used by the Global Logging functions
var GlobalLogger Logger = NewGlobalIQLogger()

// NewIQLogger creates and return a new Logger
func NewGlobalIQLogger() Logger {
	l := &logger{
		ctx:       nil,
		debug:     false,
		out:       os.Stderr,
		useColour: terminal.IsTerminal(int(os.Stderr.Fd())) && (runtime.GOOS != "windows"),
	}
	l.slog = slog.New(l)
	return l
}

// Init performs all the base configuration of the GlobalLogger
// This function is typically called when the application is starting up
// It sets log/slog and log's default output to use the iqlog format.
func Init(applicationName string, syslogHost string, debugMode bool) {
	SetApplicationName(applicationName)
	SetDebugMode(debugMode)
	SetSyslogHost(syslogHost)
	our, ok := GlobalLogger.(*logger)
	if ok {
		slog.SetDefault(our.slog)
	}
}

// SetApplicationName sets the application name on the GlobalLogger
func SetApplicationName(applicationName string) {
	GlobalLogger.SetApplicationName(applicationName)
}

// SetSyslogHost sets the syslog host on the GlobalLogger
func SetSyslogHost(host string) {
	GlobalLogger.SetSyslogHost(host)
}

// SetDebugMode sets the syslog host on the GlobalLogger
func SetDebugMode(debugMode bool) {
	GlobalLogger.SetDebugMode(debugMode)
}

// Add a context to the log entry.
func WithContext(ctx context.Context) Logger {
	existing, ok := GlobalLogger.(*logger)
	if !ok {
		return &logger{
			ctx:  ctx,
			slog: slog.Default(),
		}
	} else {
		return &logger{
			ctx:             ctx,
			slog:            slog.Default(),
			debug:           existing.debug,
			applicationName: existing.applicationName,
			syslogHost:      existing.syslogHost,
		}
	}
}

// Maximum Number of paths to log
var NumPathsToLog = 1

// How many callers are we going back
var CallersNum = 2

// GetCallerFields returns information about the caller, specifically
// the function name, the filename and line number in the file
func GetCallerFields() map[string]any {
	resp := make(map[string]any)

	OriginFile := ""
	OriginLine := 0
	OriginFunc := ""

	if pc, file, line, ok := runtime.Caller(CallersNum); ok {
		OriginFile = file
		OriginLine = line
		runtimeFuncPtr := runtime.FuncForPC(pc)
		OriginFunc = runtimeFuncPtr.Name()
	}

	OriginFunc = KeepNumDirs(OriginFunc, NumPathsToLog)
	OriginFile = KeepNumDirs(OriginFile, NumPathsToLog)

	if OriginFile != "" {
		resp["origin_file"] = OriginFile
	}
	if OriginLine > 0 {
		resp["origin_line"] = OriginLine
	}
	if OriginFunc != "" {
		resp["origin_func"] = OriginFunc
	}
	return resp
}

func keepNumDirs(str string, lastn int, startat int) string {
	numFound := strings.Count(str[startat:], "/")
	if numFound > lastn {
		return keepNumDirs(str, lastn, 1+startat+strings.Index(str[startat:], "/"))
	}
	return str[startat:]
}

// KeepNumDirs returns the lastn number of directories in a path+filename combo
func KeepNumDirs(str string, lastn int) string {
	return keepNumDirs(str, lastn, 0)
}

// Debugf is a global helper / convenience function for accessing the GlobalLogger object
func Debugf(format string, args ...any) { GlobalLogger.Debugf(format, args...) }

// Debug is a global helper / convenience function for accessing the GlobalLogger object
func Debug(args ...any) { GlobalLogger.Debug(args...) }

// Infof is a global helper / convenience function for accessing the GlobalLogger object
func Infof(format string, args ...any) { GlobalLogger.Infof(format, args...) }

// Info is a global helper / convenience function for accessing the GlobalLogger object
func Info(args ...any) { GlobalLogger.Info(args...) }

// Printf is a global helper / convenience function for accessing the GlobalLogger object
func Printf(format string, args ...any) { GlobalLogger.Printf(format, args...) }

// Print is a global helper / convenience function for accessing the GlobalLogger object
func Print(args ...any) { GlobalLogger.Print(args...) }

// Warnf is a global helper / convenience function for accessing the GlobalLogger object
func Warnf(format string, args ...any) { GlobalLogger.Warnf(format, args...) }

// Warn is a global helper / convenience function for accessing the GlobalLogger object
func Warn(args ...any) { GlobalLogger.Warn(args...) }

// Warningf is a global helper / convenience function for accessing the GlobalLogger object
func Warningf(format string, args ...any) { GlobalLogger.Warningf(format, args...) }

// Warning is a global helper / convenience function for accessing the GlobalLogger object
func Warning(args ...any) { GlobalLogger.Warning(args...) }

// Errorf is a global helper / convenience function for accessing the GlobalLogger object
func Errorf(format string, args ...any) { GlobalLogger.Errorf(format, args...) }

// Error is a global helper / convenience function for accessing the GlobalLogger object
func Error(args ...any) { GlobalLogger.Error(args...) }

// Fatalf is a global helper / convenience function for accessing the GlobalLogger object
func Fatalf(format string, args ...any) { GlobalLogger.Fatalf(format, args...) }

// Fatal is a global helper / convenience function for accessing the GlobalLogger object
func Fatal(args ...any) { GlobalLogger.Fatal(args...) }

// Panicf is a global helper / convenience function for accessing the GlobalLogger object
func Panicf(format string, args ...any) { GlobalLogger.Panicf(format, args...) }

// Panic is a global helper / convenience function for accessing the GlobalLogger object
func Panic(args ...any) { GlobalLogger.Panic(args...) }

// Log is a global helper / convenience function for accessing the GlobalLogger object
func Log(level slog.Level, args ...any) { GlobalLogger.Log(level, args...) }

// Trace is a global helper / convenience function for accessing the GlobalLogger object
func Trace(args ...any) { GlobalLogger.Trace(args) }

// Tracef is a global helper / convenience function for accessing the GlobalLogger object
func Tracef(format string, args ...any) { GlobalLogger.Tracef(format, args...) }

// Logln is a global helper / convenience function for accessing the GlobalLogger object
func Logln(level slog.Level, args ...any) { GlobalLogger.Logln(level, args) }

// Traceln is a global helper / convenience function for accessing the GlobalLogger object
func Traceln(args ...any) { GlobalLogger.Traceln(args) }

// Debugln is a global helper / convenience function for accessing the GlobalLogger object
func Debugln(args ...any) { GlobalLogger.Debugln(args) }

// Infoln is a global helper / convenience function for accessing the GlobalLogger object
func Infoln(args ...any) { GlobalLogger.Infoln(args) }

// Println is a global helper / convenience function for accessing the GlobalLogger object
func Println(args ...any) { GlobalLogger.Println(args) }

// Warnln is a global helper / convenience function for accessing the GlobalLogger object
func Warnln(args ...any) { GlobalLogger.Warnln(args) }

// Warningln is a global helper / convenience function for accessing the GlobalLogger object
func Warningln(args ...any) { GlobalLogger.Warningln(args) }

// Errorln is a global helper / convenience function for accessing the GlobalLogger object
func Errorln(args ...any) { GlobalLogger.Errorln(args) }

// Fatalln is a global helper / convenience function for accessing the GlobalLogger object
func Fatalln(args ...any) { GlobalLogger.Fatalln(args) }

// Panicln is a global helper / convenience function for accessing the GlobalLogger object
func Panicln(args ...any) { GlobalLogger.Panicln(args) }

type Logger interface {
	SetApplicationName(string)
	SetDebugMode(bool)
	SetSyslogHost(string)
	WithFields(map[string]any) Logger
	WithError(error) Logger

	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Debug(args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
	Printf(format string, args ...any)
	Print(args ...any)
	Warnf(format string, args ...any)
	Warn(args ...any)
	Warningf(format string, args ...any)
	Warning(args ...any)
	Errorf(format string, args ...any)
	Error(args ...any)
	Fatalf(format string, args ...any)
	Fatal(args ...any)
	Panicf(format string, args ...any)
	Panic(args ...any)
	Log(level slog.Level, args ...any)
	Trace(args ...any)
	Logln(level slog.Level, args ...any)
	Traceln(args ...any)
	Debugln(args ...any)
	Infoln(args ...any)
	Println(args ...any)
	Warnln(args ...any)
	Warningln(args ...any)
	Errorln(args ...any)
	Fatalln(args ...any)
	Panicln(args ...any)

	SetWriter(w io.Writer)
	Writer() io.Writer
}

const (
	levelTrace slog.Level = -3
	levelPrint slog.Level = -1
	levelPanic slog.Level = 9
	levelFatal slog.Level = 10
)

type logger struct {
	ctx context.Context
	err error

	slog *slog.Logger

	out io.Writer

	debug           bool
	applicationName string
	syslogHost      string

	IncludeTimePrefix bool
	TimePrefixFormat  string
	useColour         bool

	attrs []slog.Attr
}

var _ Logger = new(logger)

func (l *logger) SetApplicationName(name string) { l.applicationName = name }
func (l *logger) SetDebugMode(d bool)            { l.debug = d }

func (l *logger) SetSyslogHost(newhost string) {
	l.syslogHost = newhost
	if newhost != "" && !strings.Contains(newhost, ":") {
		// make sure we have a (UDP) port in the host definition
		newhost = newhost + ":514"
	}
	if l.syslogHost == newhost && newhost == "" {
		// no change
		l.Debugf("Syslog host not changed to (%s) - already set to that", newhost)
		return
	}
	if newhost == "" {
		l.Infof("Log output changed to StdErr", newhost)
		l.out = os.Stderr
		return
	}
	newSyslog, syslogErr := syslog.Dial("udp", newhost, syslog.LOG_DAEMON|syslog.LOG_INFO, l.applicationName)
	if syslogErr == nil && newSyslog != nil {
		l.out = newSyslog
		l.slog = slog.New(slog.NewJSONHandler(
			l.out,
			&slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: false,
			},
		))
		if l.applicationName != "" {
			l.Debugf("Log output for (%s) set to syslog to (%s)", l.applicationName, newhost)
		} else {
			l.Debugf("Log output set to syslog to (%s)", newhost)
		}
	} else {
		l.Errorf("ERROR: Unable to init syslog to (%s): %v", newhost, syslogErr)
		os.Exit(1)
	}
}

func (l logger) Tracef(format string, args ...any) {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	l.slog.Log(l.ctx, levelTrace, fmt.Sprintf(format, args...))
}

func (l logger) Debugf(format string, args ...any) {
	if !l.debug {
		return
	}
	format = strings.ReplaceAll(format, "%w", "%v")
	l.slog.DebugContext(l.ctx, fmt.Sprintf(format, args...))
}

func (l logger) Debug(args ...any) {
	if !l.debug {
		return
	}
	l.slog.DebugContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Debugln(args ...any) {
	if !l.debug {
		return
	}
	l.slog.DebugContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Infof(format string, args ...any) {
	format = strings.ReplaceAll(format, "%w", "%v")
	l.slog.InfoContext(l.ctx, fmt.Sprintf(format, args...))
}

func (l logger) Info(args ...any) {
	l.slog.InfoContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Printf(format string, args ...any) {
	format = strings.ReplaceAll(format, "%w", "%v")
	l.slog.Log(l.ctx, levelPrint, fmt.Sprintf(format, args...))
}

func (l logger) Print(args ...any) {
	l.slog.Log(l.ctx, levelPrint, fmt.Sprint(args...))
}

func (l logger) Warnf(format string, args ...any) {
	format = strings.ReplaceAll(format, "%w", "%v")
	l.slog.WarnContext(l.ctx, fmt.Sprintf(format, args...))
}

func (l logger) Warn(args ...any) {
	l.slog.WarnContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Warningf(format string, args ...any) {
	format = strings.ReplaceAll(format, "%w", "%v")
	l.slog.WarnContext(l.ctx, fmt.Sprintf(format, args...))
}

func (l logger) Warning(args ...any) {
	l.slog.WarnContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Errorf(format string, args ...any) {
	format = strings.ReplaceAll(format, "%w", "%v")
	l.slog.ErrorContext(l.ctx, fmt.Sprintf(format, args...))
}

func (l logger) Error(args ...any) {
	l.slog.ErrorContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Fatalf(format string, args ...any) {
	format = strings.ReplaceAll(format, "%w", "%v")
	l.slog.Log(l.ctx, levelFatal, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l logger) Fatal(args ...any) {
	l.slog.Log(l.ctx, levelFatal, fmt.Sprint(args...))
	os.Exit(1)
}

func (l logger) Panicf(format string, args ...any) {
	format = strings.ReplaceAll(format, "%w", "%v")
	l.slog.Log(l.ctx, levelPanic, fmt.Sprintf(format, args...))
	panic(fmt.Sprintf(format, args...))
}

func (l logger) Panic(args ...any) {
	l.slog.Log(l.ctx, levelPanic, fmt.Sprint(args...))
	panic(fmt.Sprint(args...))
}

func (l logger) Log(level slog.Level, args ...any) {
	l.slog.Log(l.ctx, level, fmt.Sprint(args...))
}

func (l logger) Trace(args ...any) {
	l.slog.Log(l.ctx, levelTrace, fmt.Sprint(args...))
}

func (l logger) Logln(level slog.Level, args ...any) {
	l.slog.Log(l.ctx, level, fmt.Sprint(args...))
}

func (l logger) Traceln(args ...any) {
	l.slog.Log(l.ctx, levelTrace, fmt.Sprint(args...))
}

func (l logger) Infoln(args ...any) {
	l.slog.InfoContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Println(args ...any) {
	l.slog.Log(l.ctx, levelPrint, fmt.Sprint(args...))
}

func (l logger) Warnln(args ...any) {
	l.slog.WarnContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Warningln(args ...any) {
	l.slog.WarnContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Errorln(args ...any) {
	l.slog.ErrorContext(l.ctx, fmt.Sprint(args...))
}

func (l logger) Fatalln(args ...any) {
	l.slog.Log(l.ctx, levelFatal, fmt.Sprint(args...))
	os.Exit(1)
}

func (l logger) Panicln(args ...any) {
	l.slog.Log(l.ctx, levelPanic, fmt.Sprint(args...))
	panic(fmt.Sprint(args...))
}

func (l logger) WithFields(fields map[string]any) Logger {
	attrs := make([]slog.Attr, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Attr{
			Key:   k,
			Value: slog.AnyValue(v),
		})
	}
	// l.slog = slog.New(l.WithAttrs(attrs))
	l.slog = slog.New(l.slog.Handler().WithAttrs(attrs))
	return &l
}

func (l logger) WithError(err error) Logger {
	l.slog = slog.New(l.slog.Handler().WithAttrs([]slog.Attr{{
		Key:   "error",
		Value: slog.AnyValue(err),
	}}))
	l.err = err
	return &l
}

func (l *logger) SetWriter(w io.Writer) {
	l.out = w
	l.slog = slog.New(
		slog.NewJSONHandler(
			w,
			&slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
}

func (l *logger) Writer() io.Writer { return l.out }

func SetWriter(w io.Writer) {
	GlobalLogger.SetWriter(w)
}
func GetWriter() io.Writer {
	return GlobalLogger.GetWriter()
}
