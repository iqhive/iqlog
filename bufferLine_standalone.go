package iqlog

import (
	"bytes"
	"fmt"
	"os"
)

type bufferLine struct {
	logger *logger
	// level  Level
	buffer     *bytes.Buffer
	callerData callerData
}

func (l *logger) BufferLineTrace(msg string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelTrace)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelTrace)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferLineTracef(format string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLineTrace(msg)
}

func (l *logger) BufferLineDebug(msg string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelDebug)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelDebug)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferLineDebugf(format string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLineDebug(msg)
}

func (l *logger) BufferLineInfo(msg string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelInfo)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelInfo)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferLineInfof(format string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLineInfo(msg)
}

func (l *logger) BufferLinePrint(msg string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelPrint)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelPrint)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferLinePrintf(format string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLinePrint(msg)
}

func (l *logger) BufferLineWarn(msg string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelWarn)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelWarn)
		bl.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) BufferLineWarnf(format string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLineWarn(msg)
}

func (l *logger) BufferLineError(msg string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelError)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelError)
		bl.writeFinalConsole(msg, args...)
	}
	return
}

func (l *logger) BufferLineErrorf(format string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLineError(msg)
}

func (l *logger) BufferLinePanic(msg string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelPanic)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelPanic)
		bl.writeFinalConsole(msg, args...)
	}
	bl.logger.Flush()
	os.Exit(1)
	return
}

func (l *logger) BufferLinePanicf(format string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLinePanic(msg)
}

func (l *logger) BufferLineFatal(msg string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelFatal)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(LevelFatal)
		bl.writeFinalConsole(msg, args...)
	}
	bl.logger.Flush()
	os.Exit(1)
	return
}

func (l *logger) BufferLineFatalf(format string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLineFatal(msg)
}

// Log writes a log message at the level
func (l *logger) BufferLineLog(level Level, msg string, args ...interface{}) {
	if l.Level > level {
		return
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(level)
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeInitialConsole(level)
		bl.writeFinalConsole(msg, args...)
	}
}

func (l *logger) BufferLineLogf(level Level, format string, args ...interface{}) {
	if l.Level > level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.BufferLineLog(level, msg)
}
