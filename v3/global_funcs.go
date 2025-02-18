package iqlog

// ------------------------------------------------------------
// Common with getter
// ------------------------------------------------------------
func (*logger) getWith(level Level) *bytesliceLine {
	if GlobalLogger.Level > level {
		return noopbytesliceLine
	}
	l := emptybytesliceLine(GlobalLogger)

	callerData := callerData{}
	if GlobalLogger.CallerDepth > 0 {
		var pc PC
		// +1 caller depth for the caller of the caller
		caller1(GlobalLogger.CallerDepth+1, &pc, 1, 1)
		fillCallerData(pc, &callerData)
	}
	l.callerData = callerData

	if l.jsonMode {
		l.writeInitialJSON(level)
	} else {
		l.writeInitialConsole(level)
	}

	return l
}

// ------------------------------------------------------------
// Trace
// ------------------------------------------------------------
func Trace(msg string, args ...interface{}) {
	GlobalLogger.getWith(LevelTrace).Msgs(msg, args...)
}
func Tracef(format string, args ...interface{}) {
	GlobalLogger.getWith(LevelTrace).Msgf(format, args...)
}
func TraceWith() *bytesliceLine {
	return GlobalLogger.getWith(LevelTrace)
}

// ------------------------------------------------------------
// Debug
// ------------------------------------------------------------
func Debug(msg string, args ...interface{}) {
	GlobalLogger.getWith(LevelDebug).Msgs(msg, args...)
}
func Debugf(format string, args ...interface{}) {
	GlobalLogger.getWith(LevelDebug).Msgf(format, args...)
}
func DebugWith() *bytesliceLine {
	return GlobalLogger.getWith(LevelDebug)
}

// ------------------------------------------------------------
// Info
// ------------------------------------------------------------
func Info(msg string, args ...interface{}) {
	GlobalLogger.getWith(LevelInfo).Msgs(msg, args...)
}
func Infof(format string, args ...interface{}) {
	GlobalLogger.getWith(LevelInfo).Msgf(format, args...)
}
func InfoWith() *bytesliceLine {
	return GlobalLogger.getWith(LevelInfo)
}

// ------------------------------------------------------------
// Warn
// ------------------------------------------------------------
func Warn(msg string, args ...interface{}) {
	GlobalLogger.getWith(LevelWarn).Msgs(msg, args...)
}
func Warnf(format string, args ...interface{}) {
	GlobalLogger.getWith(LevelWarn).Msgf(format, args...)
}
func WarnWith() *bytesliceLine {
	return GlobalLogger.getWith(LevelWarn)
}

// ------------------------------------------------------------
// Error
// ------------------------------------------------------------
func Error(msg string, args ...interface{}) {
	GlobalLogger.getWith(LevelError).Msgs(msg, args...)
}
func Errorf(format string, args ...interface{}) {
	GlobalLogger.getWith(LevelError).Msgf(format, args...)
}
func ErrorWith() *bytesliceLine {
	return GlobalLogger.getWith(LevelError)
}

// ------------------------------------------------------------
// Fatal
// ------------------------------------------------------------
func Fatal(msg string, args ...interface{}) {
	GlobalLogger.getWith(LevelFatal).Msgs(msg, args...)
}
func Fatalf(format string, args ...interface{}) {
	GlobalLogger.getWith(LevelFatal).Msgf(format, args...)
}
func FatalWith() *bytesliceLine {
	return GlobalLogger.getWith(LevelFatal)
}

// ------------------------------------------------------------
// Panic
// ------------------------------------------------------------
func Panic(msg string, args ...interface{}) {
	GlobalLogger.getWith(LevelPanic).Msgs(msg, args...)
}
func Panicf(format string, args ...interface{}) {
	GlobalLogger.getWith(LevelPanic).Msgf(format, args...)
}
func PanicWith() *bytesliceLine {
	return GlobalLogger.getWith(LevelPanic)
}
