package iqlog

// ------------------------------------------------------------
// Trace
// ------------------------------------------------------------
func (l *logger) Trace(msg string, args ...interface{}) {
	l.getWith(LevelTrace).Msgs(msg, args...)
}
func (l *logger) Tracef(format string, args ...interface{}) {
	l.getWith(LevelTrace).Msgf(format, args...)
}
func (l *logger) TraceWith() *bytesliceLine {
	return l.getWith(LevelTrace)
}

// ------------------------------------------------------------
// Debug
// ------------------------------------------------------------
func (l *logger) Debug(msg string, args ...interface{}) {
	l.getWith(LevelDebug).Msgs(msg, args...)
}
func (l *logger) Debugf(format string, args ...interface{}) {
	l.getWith(LevelDebug).Msgf(format, args...)
}
func (l *logger) DebugWith() *bytesliceLine {
	return l.getWith(LevelDebug)
}

// ------------------------------------------------------------
// Info
// ------------------------------------------------------------
func (l *logger) Info(msg string, args ...interface{}) {
	l.getWith(LevelInfo).Msgs(msg, args...)
}
func (l *logger) Infof(format string, args ...interface{}) {
	l.getWith(LevelInfo).Msgf(format, args...)
}
func (l *logger) InfoWith() *bytesliceLine {
	return l.getWith(LevelInfo)
}

// ------------------------------------------------------------
// Warn
// ------------------------------------------------------------
func (l *logger) Warn(msg string, args ...interface{}) {
	l.getWith(LevelWarn).Msgs(msg, args...)
}
func (l *logger) Warnf(format string, args ...interface{}) {
	l.getWith(LevelWarn).Msgf(format, args...)
}
func (l *logger) WarnWith() *bytesliceLine {
	return l.getWith(LevelWarn)
}

// ------------------------------------------------------------
// Error
// ------------------------------------------------------------
func (l *logger) Error(msg string, args ...interface{}) {
	l.getWith(LevelError).Msgs(msg, args...)
}
func (l *logger) Errorf(format string, args ...interface{}) {
	l.getWith(LevelError).Msgf(format, args...)
}
func (l *logger) ErrorWith() *bytesliceLine {
	return l.getWith(LevelError)
}

// ------------------------------------------------------------
// Fatal
// ------------------------------------------------------------
func (l *logger) Fatal(msg string, args ...interface{}) {
	l.getWith(LevelFatal).Msgs(msg, args...)
}
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.getWith(LevelFatal).Msgf(format, args...)
}
func (l *logger) FatalWith() *bytesliceLine {
	return l.getWith(LevelFatal)
}

// ------------------------------------------------------------
// Panic
// ------------------------------------------------------------
func (l *logger) Panic(msg string, args ...interface{}) {
	l.getWith(LevelPanic).Msgs(msg, args...)
}
func (l *logger) Panicf(format string, args ...interface{}) {
	l.getWith(LevelPanic).Msgf(format, args...)
}
func (l *logger) PanicWith() *bytesliceLine {
	return l.getWith(LevelPanic)
}
