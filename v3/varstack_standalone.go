package iqlog

func (l *logger) VarStackTrace(msg string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	vs := emptyvarStack(l, LevelTrace)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) VarStackTracef(format string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	vs := emptyvarStack(l, LevelTrace)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}

func (l *logger) VarStackDebug(msg string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	vs := emptyvarStack(l, LevelDebug)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) VarStackDebugf(format string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	vs := emptyvarStack(l, LevelDebug)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}

func (l *logger) VarStackInfo(msg string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	vs := emptyvarStack(l, LevelInfo)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) VarStackInfof(format string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	vs := emptyvarStack(l, LevelInfo)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}

func (l *logger) VarStackPrint(msg string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	vs := emptyvarStack(l, LevelPrint)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) VarStackPrintf(format string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	vs := emptyvarStack(l, LevelPrint)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}

func (l *logger) VarStackWarn(msg string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	vs := emptyvarStack(l, LevelWarn)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) VarStackWarnf(format string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	vs := emptyvarStack(l, LevelWarn)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}

func (l *logger) VarStackError(msg string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	vs := emptyvarStack(l, LevelError)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) VarStackErrorf(format string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	vs := emptyvarStack(l, LevelError)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}

func (l *logger) VarStackFatal(msg string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	vs := emptyvarStack(l, LevelFatal)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) VarStackFatalf(format string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	vs := emptyvarStack(l, LevelFatal)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}

func (l *logger) VarStackPanic(msg string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	vs := emptyvarStack(l, LevelPanic)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) VarStackPanicf(format string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	vs := emptyvarStack(l, LevelPanic)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}

// Log writes a log message at the level
func (l *logger) VarStackLog(level Level, msg string, args ...interface{}) {
	if l.Level > level {
		return
	}

	vs := emptyvarStack(l, level)
	if vs.jsonMode {
		vs.writeFinalJSON(msg, args...)
	} else {
		vs.writeFinalConsole(msg, args...)
	}
}

func (l *logger) VarStackLogf(level Level, format string, args ...interface{}) {
	if l.Level > level {
		return
	}

	vs := emptyvarStack(l, level)
	if len(args) > 0 {
		if vs.jsonMode {
			vs.writeFinalJSONF(format, args...)
		} else {
			vs.writeFinalConsoleF(format, args...)
		}
	} else {
		if vs.jsonMode {
			vs.writeFinalJSON(format)
		} else {
			vs.writeFinalConsole(format)
		}
	}
}
