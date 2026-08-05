package iqlog

// ------------------------------------------------------------
// Trace
// ------------------------------------------------------------
func (l *Logger) Trace(msg string, args ...interface{}) {
	l.getWith(LevelTrace).Msgs(msg, args...)
}
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.getWith(LevelTrace).Msgf(format, args...)
}
func (l *Logger) Traceln(args ...interface{}) {
	l.getWith(LevelTrace).Msg(sprintln(args...))
}

// ------------------------------------------------------------
// Debug
// ------------------------------------------------------------
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.getWith(LevelDebug).Msgs(msg, args...)
}
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.getWith(LevelDebug).Msgf(format, args...)
}
func (l *Logger) Debugln(args ...interface{}) {
	l.getWith(LevelDebug).Msg(sprintln(args...))
}

// ------------------------------------------------------------
// Info
// ------------------------------------------------------------
func (l *Logger) Info(msg string, args ...interface{}) {
	l.getWith(LevelInfo).Msgs(msg, args...)
}
func (l *Logger) Infof(format string, args ...interface{}) {
	l.getWith(LevelInfo).Msgf(format, args...)
}
func (l *Logger) Infoln(args ...interface{}) {
	l.getWith(LevelInfo).Msg(sprintln(args...))
}

// ------------------------------------------------------------
// Warn
// ------------------------------------------------------------
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.getWith(LevelWarn).Msgs(msg, args...)
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.getWith(LevelWarn).Msgf(format, args...)
}
func (l *Logger) Warnln(args ...interface{}) {
	l.getWith(LevelWarn).Msg(sprintln(args...))
}

// ------------------------------------------------------------
// Error
// ------------------------------------------------------------
func (l *Logger) Error(msg string, args ...interface{}) {
	l.getWith(LevelError).Msgs(msg, args...)
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.getWith(LevelError).Msgf(format, args...)
}
func (l *Logger) Errorln(args ...interface{}) {
	l.getWith(LevelError).Msg(sprintln(args...))
}

// ------------------------------------------------------------
// Fatal
// ------------------------------------------------------------
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.getWith(LevelFatal).Msgs(msg, args...)
}
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.getWith(LevelFatal).Msgf(format, args...)
}
func (l *Logger) Fatalln(args ...interface{}) {
	l.getWith(LevelFatal).Msg(sprintln(args...))
}

// ------------------------------------------------------------
// Panic
// ------------------------------------------------------------
func (l *Logger) Panic(msg string, args ...interface{}) {
	l.getWith(LevelPanic).Msgs(msg, args...)
}
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.getWith(LevelPanic).Msgf(format, args...)
}
func (l *Logger) Panicln(args ...interface{}) {
	l.getWith(LevelPanic).Msg(sprintln(args...))
}
