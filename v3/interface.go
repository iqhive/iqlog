package iqlog

type Logger interface {
	SetApplicationName(string)
	SetDebugMode(bool)
	SetUseColour(bool)
	SetCaptureCallers(bool)
	SetNewLine(bool)
	SetJSONMode(bool)
	SetSyslogHost(string)

	// TODO: implement these
	// WithFields(map[string]any) Logger
	// WithError(error) Logger

	Tracef(format string, args ...interface{})
	Trace(msg string, args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(msg string, args ...interface{})
	Infof(format string, args ...interface{})
	Info(msg string, args ...interface{})
	Printf(format string, args ...interface{})
	Print(msg string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warn(msg string, args ...interface{})
	Errorf(format string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatal(msg string, args ...interface{})
	Panicf(format string, args ...interface{})
	Panic(msg string, args ...interface{})
	Log(level Level, msg string, args ...interface{})
	Logf(level Level, format string, args ...interface{})
}

// var _ Logger = new(logger)
