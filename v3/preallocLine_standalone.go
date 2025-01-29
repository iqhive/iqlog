package iqlog

import (
	"fmt"
	"io"
	"os"
)

type preallocLine struct {
	out         io.Writer
	output      *[maxLineLen]byte
	bytesUsed   int
	jsonMode    bool
	includeTime bool
}

func (l *logger) PreAllocLineTrace(msg string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelTrace)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelTrace)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLineTracef(format string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLineTrace(msg)
}

func (l *logger) PreAllocLineDebug(msg string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelDebug)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelDebug)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLineDebugf(format string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLineDebug(msg)
}

func (l *logger) PreAllocLineInfo(msg string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelInfo)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelInfo)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLineInfof(format string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLineInfo(msg)
}

func (l *logger) PreAllocLinePrint(msg string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelPrint)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelPrint)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLinePrintf(format string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLinePrint(msg)
}

func (l *logger) PreAllocLineWarn(msg string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelWarn)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelWarn)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLineWarnf(format string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLineWarn(msg)
}

func (l *logger) PreAllocLineError(msg string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelError)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelError)
		pal.writeFinalConsole(msg, args...)
	}
	return
}

func (l *logger) PreAllocLineErrorf(format string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLineError(msg)
}

func (l *logger) PreAllocLinePanic(msg string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelPanic)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelPanic)
		pal.writeFinalConsole(msg, args...)
	}
	// pal.logger.Flush()
	os.Exit(1)
	return
}

func (l *logger) PreAllocLinePanicf(format string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLinePanic(msg)
}

func (l *logger) PreAllocLineFatal(msg string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelFatal)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelFatal)
		pal.writeFinalConsole(msg, args...)
	}
	// pal.logger.Flush()
	os.Exit(1)
	return
}

func (l *logger) PreAllocLineFatalf(format string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLineFatal(msg)
}

// Log writes a log message at the level
func (l *logger) PreAllocLineLog(level Level, msg string, args ...interface{}) {
	if l.Level > level {
		return
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(level)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(level)
		pal.writeFinalConsole(msg, args...)
	}
}

func (l *logger) PreAllocLineLogf(level Level, format string, args ...interface{}) {
	if l.Level > level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLineLog(level, msg)
}
