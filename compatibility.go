package iqlog

import (
	"context"
	"io"
	"os"
	"strconv"
)

// NewIQLogger returns a logger using JSON or console output.
// Deprecated: use New with Config.
func NewIQLogger(jsonMode bool) *Logger {
	format := FormatConsole
	if jsonMode {
		format = FormatJSON
	}
	level := LevelInfo
	if enabled, _ := strconv.ParseBool(os.Getenv("IQLOG_DEBUG")); enabled {
		level = LevelDebug
	}
	return MustNew(Config{Format: format, IncludeTime: true, Level: level})
}

// NewGlobalIQLogger returns a logger configured like the package default.
// Deprecated: use New with Config.
func NewGlobalIQLogger() *Logger { return MustNew(Config{IncludeTime: true, CallerDepth: 1}) }

// Init configures the package default logger.
// Deprecated: use SetDefault(New(Config{...})).
func Init(applicationName, syslogHost string, debugMode bool) {
	level := LevelInfo
	if debugMode {
		level = LevelDebug
	}
	SetDefault(MustNew(Config{ApplicationName: applicationName, SyslogHost: syslogHost, Level: level, IncludeTime: true, CallerDepth: 1}))
}

// Warning is an alias for Warn.
// Deprecated: use Warn.
func Warning(msg string, args ...any) { Warn(msg, args...) }

// Warningf is an alias for Warnf.
// Deprecated: use Warnf.
func Warningf(format string, args ...any) { Warnf(format, args...) }

// Warningln is an alias for Warnln.
// Deprecated: use Warnln.
func Warningln(args ...any) { Warnln(args...) }

func (l *Logger) Warning(msg string, args ...any)     { l.Warn(msg, args...) }
func (l *Logger) Warningf(format string, args ...any) { l.Warnf(format, args...) }
func (l *Logger) Warningln(args ...any)               { l.Warnln(args...) }

// TraceWith is an alias for TraceEvent.
// Deprecated: use TraceEvent.
func TraceWith() *Event { return TraceEvent() }

// DebugWith is an alias for DebugEvent.
// Deprecated: use DebugEvent.
func DebugWith() *Event { return DebugEvent() }

// InfoWith is an alias for InfoEvent.
// Deprecated: use InfoEvent.
func InfoWith() *Event { return InfoEvent() }

// WarnWith is an alias for WarnEvent.
// Deprecated: use WarnEvent.
func WarnWith() *Event { return WarnEvent() }

// ErrorWith is an alias for ErrorEvent.
// Deprecated: use ErrorEvent.
func ErrorWith() *Event { return ErrorEvent() }

// PanicWith is an alias for PanicEvent.
// Deprecated: use PanicEvent.
func PanicWith() *Event { return PanicEvent() }

// FatalWith is an alias for FatalEvent.
// Deprecated: use FatalEvent.
func FatalWith() *Event { return FatalEvent() }

func (l *Logger) TraceWith() *Event { return l.TraceEvent() }
func (l *Logger) DebugWith() *Event { return l.DebugEvent() }
func (l *Logger) InfoWith() *Event  { return l.InfoEvent() }
func (l *Logger) WarnWith() *Event  { return l.WarnEvent() }
func (l *Logger) ErrorWith() *Event { return l.ErrorEvent() }
func (l *Logger) PanicWith() *Event { return l.PanicEvent() }
func (l *Logger) FatalWith() *Event { return l.FatalEvent() }

func (l *Logger) SetWriter(w io.Writer) error      { return l.setWriter(w) }
func (l *Logger) SetAsyncWriter(w io.Writer) error { return l.setAsyncWriter(w, 1000, OverflowBlock) }
func (l *Logger) SetRingbufferWriter(w io.Writer) error {
	return l.setRingBufferWriter(w, 10000, OverflowSync)
}
func (l *Logger) SetRingBufferWriter(w io.Writer) error {
	return l.setRingBufferWriter(w, 10000, OverflowSync)
}
func (l *Logger) GetWriter() io.Writer { return l.getWriter() }
func SetWriter(w io.Writer) error      { return Default().SetWriter(w) }
func GetWriter() io.Writer             { return Default().GetWriter() }

func (l *Logger) SetLevel(level Level) { l.setLevel(level) }
func SetLevel(level Level)             { Default().SetLevel(level) }
func (l *Logger) SetDebugMode(enabled bool) {
	if enabled {
		l.setLevel(LevelDebug)
	} else {
		l.setLevel(LevelInfo)
	}
}
func SetDebugMode(enabled bool)             { Default().SetDebugMode(enabled) }
func (l *Logger) SetCallerDepth(depth int)  { l.setCallerDepth(depth) }
func SetCallerDepth(depth int)              { Default().SetCallerDepth(depth) }
func (l *Logger) SetUseColour(enabled bool) { l.setUseColor(enabled) }
func SetUseColour(enabled bool)             { Default().SetUseColour(enabled) }
func (l *Logger) SetUseColor(enabled bool)  { l.setUseColor(enabled) }
func SetUseColor(enabled bool)              { Default().SetUseColor(enabled) }
func (l *Logger) SetJSONTimeMode(mode JSONTimeMode) {
	l.updateConfig(func(cfg *loggerConfig) {
		cfg.format = FormatJSON
		cfg.jsonTimeMode = mode
		cfg.includeTime = mode != JSONTimeDisabled
	})
}
func SetJSONTimeMode(mode JSONTimeMode)          { Default().SetJSONTimeMode(mode) }
func (l *Logger) SetNewLine(bool)                {}
func SetNewLine(enabled bool)                    { Default().SetNewLine(enabled) }
func (l *Logger) SetApplicationName(name string) { l.setApplicationName(name) }
func SetApplicationName(name string)             { Default().SetApplicationName(name) }
func (l *Logger) SetSyslogHost(host string)      { l.setSyslogHostLegacy(host) }
func SetSyslogHost(host string)                  { Default().SetSyslogHost(host) }

// Flush flushes the package default logger.
// Deprecated: use Default().Flush().
func Flush() error { return Default().Flush() }

func (l *Logger) LogWithFields(ctx context.Context, level Level, fields map[string]any, msg string, args ...any) {
	l.WithFields(fields).LogContext(ctx, level, msg, args...)
}
func (l *Logger) LogFWithFields(ctx context.Context, level Level, fields map[string]any, format string, args ...any) {
	l.WithFields(fields).LogContextf(ctx, level, format, args...)
}
