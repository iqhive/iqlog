package iqlog

func (l *logger) Trace(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.WithByteSliceLineTrace().Msgs(msg, args...)
	} else {
		l.WithByteSliceLineTrace().Msg(msg)
	}
}
func Trace(msg string, args ...interface{}) {
	if len(args) > 0 {
		GlobalLogger.WithByteSliceLineTrace().Msgs(msg, args...)
	} else {
		GlobalLogger.WithByteSliceLineTrace().Msg(msg)
	}
}
func (l *logger) Tracef(format string, args ...interface{}) {
	l.WithByteSliceLineTrace().Msgf(format, args...)
}
func Tracef(format string, args ...interface{}) {
	GlobalLogger.Tracef(format, args...)
}

func (l *logger) Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.WithByteSliceLineDebug().Msgs(msg, args...)
	} else {
		l.WithByteSliceLineDebug().Msg(msg)
	}
}
func Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		GlobalLogger.WithByteSliceLineDebug().Msgs(msg, args...)
	} else {
		GlobalLogger.WithByteSliceLineDebug().Msg(msg)
	}
}
func (l *logger) Debugf(format string, args ...interface{}) {
	l.WithByteSliceLineDebug().Msgf(format, args...)
}
func Debugf(format string, args ...interface{}) {
	GlobalLogger.Debugf(format, args...)
}

func (l *logger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.WithByteSliceLineInfo().Msgs(msg, args...)
	} else {
		l.WithByteSliceLineInfo().Msg(msg)
	}
}
func Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		GlobalLogger.WithByteSliceLineInfo().Msgs(msg, args...)
	} else {
		GlobalLogger.WithByteSliceLineInfo().Msg(msg)
	}
}
func (l *logger) Infof(format string, args ...interface{}) {
	l.WithByteSliceLineInfo().Msgf(format, args...)
}
func Infof(format string, args ...interface{}) {
	GlobalLogger.Infof(format, args...)
}

func (l *logger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.WithByteSliceLineWarn().Msgs(msg, args...)
	} else {
		l.WithByteSliceLineWarn().Msg(msg)
	}
}
func Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		GlobalLogger.WithByteSliceLineWarn().Msgs(msg, args...)
	} else {
		GlobalLogger.WithByteSliceLineWarn().Msg(msg)
	}
}
func (l *logger) Warnf(format string, args ...interface{}) {
	l.WithByteSliceLineWarn().Msgf(format, args...)
}
func Warnf(format string, args ...interface{}) {
	GlobalLogger.Warnf(format, args...)
}

func (l *logger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.WithByteSliceLineError().Msgs(msg, args...)
	} else {
		l.WithByteSliceLineError().Msg(msg)
	}
}
func Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		GlobalLogger.WithByteSliceLineError().Msgs(msg, args...)
	} else {
		GlobalLogger.WithByteSliceLineError().Msg(msg)
	}
}
func (l *logger) Errorf(format string, args ...interface{}) {
	l.WithByteSliceLineError().Msgf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	GlobalLogger.Errorf(format, args...)
}

func (l *logger) Fatal(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.WithByteSliceLineFatal().Msgs(msg, args...)
	} else {
		l.WithByteSliceLineFatal().Msg(msg)
	}
}
func Fatal(msg string, args ...interface{}) {
	if len(args) > 0 {
		GlobalLogger.WithByteSliceLineFatal().Msgs(msg, args...)
	} else {
		GlobalLogger.WithByteSliceLineFatal().Msg(msg)
	}
}
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.WithByteSliceLineFatal().Msgf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	GlobalLogger.Fatalf(format, args...)
}

func (l *logger) Panic(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.WithByteSliceLinePanic().Msgs(msg, args...)
	} else {
		l.WithByteSliceLinePanic().Msg(msg)
	}
}
func Panic(msg string, args ...interface{}) {
	if len(args) > 0 {
		GlobalLogger.WithByteSliceLinePanic().Msgs(msg, args...)
	} else {
		GlobalLogger.WithByteSliceLinePanic().Msg(msg)
	}
}
func (l *logger) Panicf(format string, args ...interface{}) {
	l.WithByteSliceLinePanic().Msgf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	GlobalLogger.Panicf(format, args...)
}
