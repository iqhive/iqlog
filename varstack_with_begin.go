package iqlog

func (l *logger) WithVarStackTrace() *varStack {
	if l.Level() > LevelTrace {
		return noopvarStack
	}
	vs := emptyvarStack(l, LevelTrace)

	return vs
}

func (l *logger) WithVarStackDebug() *varStack {
	if l.Level() > LevelDebug {
		return noopvarStack
	}
	vs := emptyvarStack(l, LevelDebug)

	return vs
}

func (l *logger) WithVarStackInfo() *varStack {
	if l.Level() > LevelInfo {
		return noopvarStack
	}
	vs := emptyvarStack(l, LevelInfo)

	return vs
}

func (l *logger) WithVarStackWarn() *varStack {
	if l.Level() > LevelWarn {
		return noopvarStack
	}
	vs := emptyvarStack(l, LevelWarn)

	return vs
}

func (l *logger) WithVarStackError() *varStack {
	if l.Level() > LevelError {
		return noopvarStack
	}
	vs := emptyvarStack(l, LevelError)

	return vs
}

func (l *logger) WithVarStackPanic() *varStack {
	if l.Level() > LevelPanic {
		return noopvarStack
	}
	vs := emptyvarStack(l, LevelPanic)

	// vs.logger.Flush()
	// os.Exit(1)
	return vs
}

func (l *logger) WithVarStackFatal() *varStack {
	if l.Level() > LevelFatal {
		return noopvarStack
	}
	vs := emptyvarStack(l, LevelFatal)

	// vs.logger.Flush()
	// os.Exit(1)
	return vs
}

// func (l *logger) Str(name string, s string) *varStack {
// 	vs := &varStack{
// 		logger: l,
// 		level:  LevelUnknown,
// 	}
// 	if vs.jsonMode {
// 		vs.writeInitialJSON(LevelUnknown)
// 	} else {
// 		vs.writeInitialConsole(LevelUnknown)
// 	}
// 	return vs
// }

// func (l *logger) Any(name string, v any) *varStack {
// 	vs := &varStack{
// 		logger: l,
// 		level:  LevelUnknown,
// 	}
// 	if vs.jsonMode {
// 		vs.writeInitialJSON(LevelUnknown)
// 	} else {
// 		vs.writeInitialConsole(LevelUnknown)
// 	}
// 	return vs
// }
