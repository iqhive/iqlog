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
	bl.buffer.WriteByte('{')
	bl.AddTime() // adds a comma if it runs

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

}

func (bl *bufferLine) writeInitialConsole(level Level) {
	bl.AddTime()
	if useColour {
		bl.buffer.Write(ansiColourPrefix(level))
	} else {
		bl.buffer.Write(levelPrefix(level))
	}
}

func (bl *bufferLine) AddTime() {
	if !bl.logger.IncludeTime {
		return
	}

	timeNow := time.Now()
	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	if bl.logger.jsonMode {
		var prefixArr = [37]byte{
			',', '"', 't', 'i', 'm', 'e', '"', ':', '"',
			'0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
			'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
			'0', '0', '0', '0', '0', '0', '"', ',',
		}

		setIntBytes(prefixArr[9:], int64(year), 4)
		setIntBytes(prefixArr[14:], int64(month), 2)
		setIntBytes(prefixArr[17:], int64(day), 2)
		setIntBytes(prefixArr[20:], int64(hour), 2)
		setIntBytes(prefixArr[23:], int64(min), 2)
		setIntBytes(prefixArr[26:], int64(sec), 2)
		setIntBytes(prefixArr[29:], int64(usec), 6)
		if bl.buffer.Len() > 1 {
			// Use the header comma
			bl.buffer.Write(prefixArr[:37])
		} else {
			// Skip the header comma
			bl.buffer.Write(prefixArr[1:37])
		}

	} else {
		var consolePrefixFull = [29]byte{
			'[', '0', '0', '0', '0', '-', '0', '0', '-', '0', '0',
			'T', '0', '0', ':', '0', '0', ':', '0', '0', '.',
			'0', '0', '0', '0', '0', '0', ']', ' ',
		}

		setIntBytes(consolePrefixFull[1:], int64(year), 4)
		setIntBytes(consolePrefixFull[6:], int64(month), 2)
		setIntBytes(consolePrefixFull[9:], int64(day), 2)
		setIntBytes(consolePrefixFull[12:], int64(hour), 2)
		setIntBytes(consolePrefixFull[15:], int64(min), 2)
		setIntBytes(consolePrefixFull[18:], int64(sec), 2)
		setIntBytes(consolePrefixFull[21:], int64(usec), 6)
		bl.buffer.Write(consolePrefixFull[:29])
	}
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
