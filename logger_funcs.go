package iqlog

// The ...ln wrappers join their arguments before handing the result to Msg.
// getWith returns nil when the level gate dropped the record, and checking
// that first keeps the disabled path free of the sprintln allocation. Fatal
// and Panic always get an event back, so they still format and terminate.

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
	if e := l.getWith(LevelTrace); e != nil {
		e.Msg(sprintln(args...))
	}
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
	if e := l.getWith(LevelDebug); e != nil {
		e.Msg(sprintln(args...))
	}
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
	if e := l.getWith(LevelInfo); e != nil {
		e.Msg(sprintln(args...))
	}
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
	if e := l.getWith(LevelWarn); e != nil {
		e.Msg(sprintln(args...))
	}
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
	if e := l.getWith(LevelError); e != nil {
		e.Msg(sprintln(args...))
	}
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
	if e := l.getWith(LevelFatal); e != nil {
		e.Msg(sprintln(args...))
	}
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
	if e := l.getWith(LevelPanic); e != nil {
		e.Msg(sprintln(args...))
	}
}
