package iqlog

import (
	"fmt"
	"io"
	"os"
)

type preallocLine2 struct {
	out           io.Writer
	output        []byte
	bytesUsed     int
	jsonMode      bool
	includeTime   bool
	captureCaller int
	callerData    callerData
}

func (l *logger) PreAllocLine2Trace(msg string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelTrace)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelTrace)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLine2Tracef(format string, args ...interface{}) {
	if l.Level > LevelTrace {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Trace(msg)
}

func (l *logger) PreAllocLine2Debug(msg string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelDebug)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelDebug)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLine2Debugf(format string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Debug(msg)
}

func (l *logger) PreAllocLine2Info(msg string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelInfo)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelInfo)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLine2Infof(format string, args ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Info(msg)
}

func (l *logger) PreAllocLine2Print(msg string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelPrint)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelPrint)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLine2Printf(format string, args ...interface{}) {
	if l.Level > LevelPrint {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Print(msg)
}

func (l *logger) PreAllocLine2Warn(msg string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelWarn)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelWarn)
		pal.writeFinalConsole(msg, args...)
	}
	return
}
func (l *logger) PreAllocLine2Warnf(format string, args ...interface{}) {
	if l.Level > LevelWarn {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Warn(msg)
}

func (l *logger) PreAllocLine2Error(msg string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelError)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(LevelError)
		pal.writeFinalConsole(msg, args...)
	}
	return
}

func (l *logger) PreAllocLine2Errorf(format string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Error(msg)
}

func (l *logger) PreAllocLine2Panic(msg string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	pal := emptypreallocLine2(l)
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

func (l *logger) PreAllocLine2Panicf(format string, args ...interface{}) {
	if l.Level > LevelPanic {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Panic(msg)
}

func (l *logger) PreAllocLine2Fatal(msg string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	pal := emptypreallocLine2(l)
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

func (l *logger) PreAllocLine2Fatalf(format string, args ...interface{}) {
	if l.Level > LevelFatal {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Fatal(msg)
}

// Log writes a log message at the level
func (l *logger) PreAllocLine2Log(level Level, msg string, args ...interface{}) {
	if l.Level > level {
		return
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(level)
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeInitialConsole(level)
		pal.writeFinalConsole(msg, args...)
	}
}

func (l *logger) PreAllocLine2Logf(level Level, format string, args ...interface{}) {
	if l.Level > level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.PreAllocLine2Log(level, msg)
}
