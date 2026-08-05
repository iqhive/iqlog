package iqlog

import (
	"bytes"
	"fmt"
	"os"
)

type bufferLineNL struct {
	logger         *logger
	buffer         *bytes.Buffer
	jsonMode       bool
	includeTime    bool
	captureCaller  int
	exitAfterWrite bool
	// panicAfterWrite panics with the completed line after it is written
	panicAfterWrite bool
	callerData      callerData
}

func (l *logger) BufferSliceLineTrace(msg string, args ...interface{}) {
	if l.Level() > LevelTrace {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelTrace)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelTrace)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferSliceLineTracef(format string, args ...interface{}) {
	if l.Level() > LevelTrace {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLineTrace(msg)
}

func (l *logger) BufferSliceLineDebug(msg string, args ...interface{}) {
	if l.Level() > LevelDebug {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelDebug)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelDebug)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferSliceLineDebugf(format string, args ...interface{}) {
	if l.Level() > LevelDebug {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLineDebug(msg)
}

func (l *logger) BufferSliceLineInfo(msg string, args ...interface{}) {
	if l.Level() > LevelInfo {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelInfo)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelInfo)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferSliceLineInfof(format string, args ...interface{}) {
	if l.Level() > LevelInfo {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLineInfo(msg)
}

func (l *logger) BufferSliceLinePrint(msg string, args ...interface{}) {
	if l.Level() > LevelPrint {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelPrint)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelPrint)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferSliceLinePrintf(format string, args ...interface{}) {
	if l.Level() > LevelPrint {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLinePrint(msg)
}

func (l *logger) BufferSliceLineWarn(msg string, args ...interface{}) {
	if l.Level() > LevelWarn {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelWarn)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelWarn)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferSliceLineWarnf(format string, args ...interface{}) {
	if l.Level() > LevelWarn {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLineWarn(msg)
}

func (l *logger) BufferSliceLineError(msg string, args ...interface{}) {
	if l.Level() > LevelError {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelError)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelError)
		bl.writeFinalConsole(msg, args...)
	}
	return
}

func (l *logger) BufferSliceLineErrorf(format string, args ...interface{}) {
	if l.Level() > LevelError {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLineError(msg)
}

func (l *logger) BufferSliceLinePanic(msg string, args ...interface{}) {
	if l.Level() > LevelPanic {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelPanic)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelPanic)
		bl.writeFinalConsole(msg, args...)
	}
	// bl.logger.Flush()
	panic(msg)
}

func (l *logger) BufferSliceLinePanicf(format string, args ...interface{}) {
	if l.Level() > LevelPanic {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLinePanic(msg)
}

func (l *logger) BufferSliceLineFatal(msg string, args ...interface{}) {
	if l.Level() > LevelFatal {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelFatal)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelFatal)
		bl.writeFinalConsole(msg, args...)
	}
	// bl.logger.Flush()
	os.Exit(1)
	return
}

func (l *logger) BufferSliceLineFatalf(format string, args ...interface{}) {
	if l.Level() > LevelFatal {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLineFatal(msg)
}

// Log writes a log message at the level
func (l *logger) BufferSliceLineLog(level Level, msg string, args ...interface{}) {
	if l.Level() > level {
		return
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(level)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(level)
		bl.writeFinalConsole(msg, args...)
	}
}

func (l *logger) BufferSliceLineLogf(level Level, format string, args ...interface{}) {
	if l.Level() > level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferSliceLineLog(level, msg)
}
