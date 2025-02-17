package iqlog

func (l *logger) DebugWith() *bytesliceLine {
	return l.WithByteSliceLineDebug()
}
func DebugWith() *bytesliceLine {
	return GlobalLogger.WithByteSliceLineDebug()
}

func (l *logger) InfoWith() *bytesliceLine {
	return l.WithByteSliceLineInfo()
}
func InfoWith() *bytesliceLine {
	return GlobalLogger.WithByteSliceLineInfo()
}

func (l *logger) WarnWith() *bytesliceLine {
	return l.WithByteSliceLineWarn()
}
func WarnWith() *bytesliceLine {
	return GlobalLogger.WithByteSliceLineWarn()
}

func (l *logger) ErrorWith() *bytesliceLine {
	return l.WithByteSliceLineError()
}
func ErrorWith() *bytesliceLine {
	return GlobalLogger.WithByteSliceLineError()
}

func (l *logger) PanicWith() *bytesliceLine {
	return l.WithByteSliceLinePanic()
}
func PanicWith() *bytesliceLine {
	return GlobalLogger.WithByteSliceLinePanic()
}

func (l *logger) FatalWith() *bytesliceLine {
	return l.WithByteSliceLineFatal()
}
func FatalWith() *bytesliceLine {
	return GlobalLogger.WithByteSliceLineFatal()
}
