package iqlog

import (
	"os"
	"time"
)

var noopbufferLine = &bufferLine{
	logger: &logger{
		Level: 999999,
	},
}

// var emptybufferLine = bufferLine{}

func emptybufferLine(l *logger) *bufferLine {
	bl := bufferLinePool.Get().(*bufferLine)
	bl.logger = l
	bl.buffer.Reset()
	return bl
}

func (bl *bufferLine) writeInitialJSON(level Level) {
	if bl.logger.IncludeTime {
		AddTimeJSONToBuffer(time.Now(), bl.buffer)
	} else {
		bl.buffer.WriteByte('{')
	}
	// Convert Level to string
	switch level {
	case LevelDebug:
		bl.buffer.Write([]byte("\"level\":\"debug\""))
	case LevelInfo:
		bl.buffer.Write([]byte("\"level\":\"info\""))
	case LevelWarn:
		bl.buffer.Write([]byte("\"level\":\"warn\""))
	case LevelError:
		bl.buffer.Write([]byte("\"level\":\"error\""))
	case LevelFatal:
		bl.buffer.Write([]byte("\"level\":\"fatal\""))
	case LevelPanic:
		bl.buffer.Write([]byte("\"level\":\"panic\""))
	default:
		bl.buffer.Write([]byte("\"level\":\"unknown\""))
	}
	bl.AddCallers()
}

func (bl *bufferLine) AddCallers() {
	if bl.logger.CallerDepth == 0 || bl.callerData.callerFuncLen == 0 {
		return
	}

	if bl.logger.jsonMode {
		// json mode
		bl.buffer.Write([]byte(`,"func":"`))
		bl.buffer.Write([]byte(bl.callerData.callerFunc[:bl.callerData.callerFuncLen]))
		bl.buffer.Write([]byte(`","file":"`))
		bl.buffer.Write([]byte(bl.callerData.callerFile[:bl.callerData.callerFileLen]))
		bl.buffer.Write([]byte(`"`))
	} else {
		// console mode
		if useColour {
			bl.buffer.Write([]byte("\x1b[32m["))
		} else {
			bl.buffer.Write([]byte(`[`))
		}
		bl.buffer.Write([]byte(bl.callerData.callerFunc[:bl.callerData.callerFuncLen]))
		bl.buffer.Write([]byte(` `))
		bl.buffer.Write([]byte(bl.callerData.callerFile[:bl.callerData.callerFileLen]))
		if useColour {
			bl.buffer.Write([]byte("]\x1b[0m "))
		} else {
			bl.buffer.Write([]byte(`] `))
		}
	}
	return
}

func (bl *bufferLine) writeInitialConsole(level Level) {
	if bl.logger.IncludeTime {
		AddTimeConsoleToBuffer(time.Now(), bl.buffer)
	}

	if useColour {
		bl.buffer.Write(ansiColourPrefix(level))
	} else {
		bl.buffer.Write(levelPrefix(level))
	}
	bl.AddCallers()
}

func (l *logger) WithBufferLineTrace() *bufferLine {
	if l.Level > LevelTrace {
		return noopbufferLine
	}
	bl := emptybufferLine(l)
	bl.logger = l
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelTrace)
	} else {
		bl.writeInitialConsole(LevelTrace)
	}
	return bl
}

func (l *logger) WithBufferLineDebug() *bufferLine {
	if l.Level > LevelDebug {
		return noopbufferLine
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelDebug)
	} else {
		bl.writeInitialConsole(LevelDebug)
	}
	return bl
}

func (l *logger) WithBufferLineInfo() *bufferLine {
	if l.Level > LevelInfo {
		return noopbufferLine
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelInfo)
	} else {
		bl.writeInitialConsole(LevelInfo)
	}
	return bl
}

func (l *logger) WithBufferLineWarn() *bufferLine {
	if l.Level > LevelWarn {
		return noopbufferLine
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelWarn)
	} else {
		bl.writeInitialConsole(LevelWarn)
	}
	return bl
}

func (l *logger) WithBufferLineError() *bufferLine {
	if l.Level > LevelError {
		return noopbufferLine
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelError)
	} else {
		bl.writeInitialConsole(LevelError)
	}
	return bl
}

func (l *logger) WithBufferLinePanic() *bufferLine {
	if l.Level > LevelPanic {
		return noopbufferLine
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelPanic)
	} else {
		bl.writeInitialConsole(LevelPanic)
	}
	bl.logger.Flush()
	os.Exit(1)
	return bl
}

func (l *logger) WithBufferLineFatal() *bufferLine {
	if l.Level > LevelFatal {
		return noopbufferLine
	}
	bl := emptybufferLine(l)
	if bl.logger.jsonMode {
		bl.writeInitialJSON(LevelFatal)
	} else {
		bl.writeInitialConsole(LevelFatal)
	}
	bl.logger.Flush()
	os.Exit(1)
	return bl
}
