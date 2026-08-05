package iqlog

func (l *Logger) getWith(level Level) *Event {
	return l.newEvent(level, 2)
}

// ------------------------------------------------------------
// Trace
// ------------------------------------------------------------
func Trace(msg string, args ...interface{}) {
	Default().Trace(msg, args...)
}
func Tracef(format string, args ...interface{}) {
	Default().Tracef(format, args...)
}
func Traceln(args ...interface{}) {
	Default().Traceln(args...)
}

// ------------------------------------------------------------
// Debug
// ------------------------------------------------------------
func Debug(msg string, args ...interface{}) {
	Default().Debug(msg, args...)
}
func Debugf(format string, args ...interface{}) {
	Default().Debugf(format, args...)
}
func Debugln(args ...interface{}) {
	Default().Debugln(args...)
}

// ------------------------------------------------------------
// Info
// ------------------------------------------------------------
func Info(msg string, args ...interface{}) {
	Default().Info(msg, args...)
}
func Infof(format string, args ...interface{}) {
	Default().Infof(format, args...)
}
func Infoln(args ...interface{}) {
	Default().Infoln(args...)
}

// ------------------------------------------------------------
// Warn
// ------------------------------------------------------------
func Warn(msg string, args ...interface{}) {
	Default().Warn(msg, args...)
}
func Warnf(format string, args ...interface{}) {
	Default().Warnf(format, args...)
}
func Warnln(args ...interface{}) {
	Default().Warnln(args...)
}

// ------------------------------------------------------------
// Error
// ------------------------------------------------------------
func Error(msg string, args ...interface{}) {
	Default().Error(msg, args...)
}
func Errorf(format string, args ...interface{}) {
	Default().Errorf(format, args...)
}
func Errorln(args ...interface{}) {
	Default().Errorln(args...)
}

// ------------------------------------------------------------
// Fatal
// ------------------------------------------------------------
func Fatal(msg string, args ...interface{}) {
	Default().Fatal(msg, args...)
}
func Fatalf(format string, args ...interface{}) {
	Default().Fatalf(format, args...)
}
func Fatalln(args ...interface{}) {
	Default().Fatalln(args...)
}

// ------------------------------------------------------------
// Panic
// ------------------------------------------------------------
func Panic(msg string, args ...interface{}) {
	Default().Panic(msg, args...)
}
func Panicf(format string, args ...interface{}) {
	Default().Panicf(format, args...)
}
func Panicln(args ...interface{}) {
	Default().Panicln(args...)
}
