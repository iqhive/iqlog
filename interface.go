package iqlog

import "context"

type Logger interface {
	// SetApplicationName(string)
	// SetDebugMode(bool)
	// SetUseColour(bool)
	// SetCaptureCallers(bool)
	// SetNewLine(bool)
	// SetJSONMode(bool)
	// SetSyslogHost(string)

	// TODO: implement these
	// WithFields(map[string]any) Logger
	// WithError(error) Logger

	Tracef(format string, args ...any)
	Trace(msg string, args ...any)
	Debugf(format string, args ...any)
	Debug(msg string, args ...any)
	Infof(format string, args ...any)
	Info(msg string, args ...any)
	Printf(format string, args ...any)
	Print(msg string, args ...any)
	Warnf(format string, args ...any)
	Warn(msg string, args ...any)
	Errorf(format string, args ...any)
	Error(msg string, args ...any)
	Fatalf(format string, args ...any)
	Fatal(msg string, args ...any)
	Panicf(format string, args ...any)
	Panic(msg string, args ...any)
	Log(ctx context.Context, level Level, msg string, args ...any)
	Logf(ctx context.Context, level Level, format string, args ...any)
}

// var _ Logger = new(logger)
