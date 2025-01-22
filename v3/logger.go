package iqlog

func (l *logger) Trace(msg string, args ...interface{}) {
	l.PreAllocLineTrace(msg, args...)
}

func (l *logger) Tracef(msg string, args ...interface{}) {
	l.PreAllocLineTracef(msg, args...)
}

func (l *logger) Debug(msg string, args ...interface{}) {
	l.PreAllocLineDebug(msg, args...)
}

func (l *logger) Debugf(msg string, args ...interface{}) {
	l.PreAllocLineDebugf(msg, args...)
}

func (l *logger) Info(msg string, args ...interface{}) {
	l.PreAllocLineInfo(msg, args...)
}

func (l *logger) Infof(msg string, args ...interface{}) {
	l.PreAllocLineInfof(msg, args...)
}

func (l *logger) Warn(msg string, args ...interface{}) {
	l.PreAllocLineWarn(msg, args...)
}

func (l *logger) Warnf(msg string, args ...interface{}) {
	l.PreAllocLineWarnf(msg, args...)
}

func (l *logger) Error(msg string, args ...interface{}) {
	l.PreAllocLineError(msg, args...)
}

func (l *logger) Errorf(msg string, args ...interface{}) {
	l.PreAllocLineErrorf(msg, args...)
}

func (l *logger) Fatal(msg string, args ...interface{}) {
	l.PreAllocLineFatal(msg, args...)
}

func (l *logger) Fatalf(msg string, args ...interface{}) {
	l.PreAllocLineFatalf(msg, args...)
}

func (l *logger) Panic(msg string, args ...interface{}) {
	l.PreAllocLinePanic(msg, args...)
}

func (l *logger) Panicf(msg string, args ...interface{}) {
	l.PreAllocLinePanicf(msg, args...)
}
