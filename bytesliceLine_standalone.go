package iqlog

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type bytesliceLine struct {
	out           io.Writer
	mu            *sync.Mutex
	jsonMode      bool
	output        []byte
	includeTime   bool
	captureCaller int
	callerData    callerData
}

func (l *logger) ByteSliceLineTrace(msg string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(LevelTrace)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(LevelTrace)
		sbl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) ByteSliceLineTracef(format string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLineTrace(msg)
}

func (l *logger) ByteSliceLineDebug(msg string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(LevelDebug)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(LevelDebug)
		sbl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) ByteSliceLineDebugf(format string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLineDebug(msg)
}

func (l *logger) ByteSliceLineInfo(msg string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(LevelInfo)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(LevelInfo)
		sbl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) ByteSliceLineInfof(format string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLineInfo(msg)
}

func (l *logger) ByteSliceLinePrint(msg string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(LevelPrint)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(LevelPrint)
		sbl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) ByteSliceLinePrintf(format string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLinePrint(msg)
}

func (l *logger) ByteSliceLineWarn(msg string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(LevelWarn)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(LevelWarn)
		sbl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) ByteSliceLineWarnf(format string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLineWarn(msg)
}

func (l *logger) ByteSliceLineError(msg string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(LevelError)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(LevelError)
		sbl.writeFinalConsole(msg, args...)
	}
	return
}

func (l *logger) ByteSliceLineErrorf(format string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLineError(msg)
}

func (l *logger) ByteSliceLinePanic(msg string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(LevelPanic)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(LevelPanic)
		sbl.writeFinalConsole(msg, args...)
	}
	// sbl.logger.Flush()
	os.Exit(1)
	return
}

func (l *logger) ByteSliceLinePanicf(format string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLinePanic(msg)
}

func (l *logger) ByteSliceLineFatal(msg string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(LevelFatal)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(LevelFatal)
		sbl.writeFinalConsole(msg, args...)
	}
	// sbl.logger.Flush()
	os.Exit(1)
	return
}

func (l *logger) ByteSliceLineFatalf(format string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLineFatal(msg)
}

// Log writes a log message at the level
func (l *logger) ByteSliceLineLog(level Level, msg string, args ...interface{}) {
	if l.Level > level {
		return
	}
	sbl := emptybytesliceLine(l)
	if sbl.jsonMode {
		sbl.writeInitialJSON(level)
		sbl.writeFinalJSON(msg, args...)
	} else {
		sbl.writeInitialConsole(level)
		sbl.writeFinalConsole(msg, args...)
	}
}

func (l *logger) ByteSliceLineLogf(level Level, format string, args ...interface{}) {
	if l.Level > level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.ByteSliceLineLog(level, msg)
}
