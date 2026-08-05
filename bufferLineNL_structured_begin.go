package iqlog

import (
	"time"
)

var noopbufferLineNL = &bufferLineNL{}

func emptybufferLineNL(l *logger) *bufferLineNL {
	bl := bufferLineNLPool.Get().(*bufferLineNL)
	bl.buffer.Reset()
	bl.logger = l
	bl.includeTime = l.IncludeTime.Load()
	bl.jsonMode = l.jsonMode.Load()
	bl.captureCaller = int(l.CallerDepth.Load())
	bl.exitAfterWrite = false
	bl.panicAfterWrite = false
	bl.callerData.callerFuncLen = 0

	return bl
}

func (bl *bufferLineNL) writeInitialJSON(level Level) {
	if bl.includeTime {
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

func (bl *bufferLineNL) AddCallers() {
	if bl.captureCaller == 0 || bl.callerData.callerFuncLen == 0 {
		return
	}

	if bl.jsonMode {
		// json mode
		bl.buffer.Write([]byte(`,"func":"`))
		bl.buffer.Write([]byte(bl.callerData.callerFunc[:bl.callerData.callerFuncLen]))
		bl.buffer.Write([]byte(`","file":"`))
		bl.buffer.Write([]byte(bl.callerData.callerFile[:bl.callerData.callerFileLen]))
		bl.buffer.Write([]byte(`"`))
	} else {
		// console mode
		if useColour.Load() {
			bl.buffer.Write([]byte("\x1b[32m["))
		} else {
			bl.buffer.Write([]byte(`[`))
		}
		bl.buffer.Write([]byte(bl.callerData.callerFunc[:bl.callerData.callerFuncLen]))
		bl.buffer.Write([]byte(` `))
		bl.buffer.Write([]byte(bl.callerData.callerFile[:bl.callerData.callerFileLen]))
		if useColour.Load() {
			bl.buffer.Write([]byte("]\x1b[0m "))
		} else {
			bl.buffer.Write([]byte(`] `))
		}
	}
}

func (bl *bufferLineNL) writeInitialConsole(level Level) {
	if bl.includeTime {
		AddTimeConsoleToBuffer(time.Now(), bl.buffer)
	}
	if useColour.Load() {
		bl.buffer.Write(ansiColourPrefix(level))
	} else {
		bl.buffer.Write(levelPrefix(level))
	}
	bl.AddCallers()
}

func (bl *bufferLineNL) AddTime() {
	if !bl.includeTime {
		return
	}

	timeNow := time.Now()
	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	if bl.jsonMode {
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
			// Use comma
			bl.buffer.Write(prefixArr[:37])
		} else {
			// Skip it
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

func (l *logger) WithBufferLineNLTrace() *bufferLineNL {
	if l.Level() > LevelTrace {
		return noopbufferLineNL
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelTrace)
	} else {
		bl.writeInitialConsole(LevelTrace)
	}
	return bl
}

func (l *logger) WithBufferLineNLDebug() *bufferLineNL {
	if l.Level() > LevelDebug {
		return noopbufferLineNL
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelDebug)
	} else {
		bl.writeInitialConsole(LevelDebug)
	}
	return bl
}

func (l *logger) WithBufferLineNLInfo() *bufferLineNL {
	if l.Level() > LevelInfo {
		return noopbufferLineNL
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelInfo)
	} else {
		bl.writeInitialConsole(LevelInfo)
	}
	return bl
}

func (l *logger) WithBufferLineNLWarn() *bufferLineNL {
	if l.Level() > LevelWarn {
		return noopbufferLineNL
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelWarn)
	} else {
		bl.writeInitialConsole(LevelWarn)
	}
	return bl
}

func (l *logger) WithBufferLineNLError() *bufferLineNL {
	if l.Level() > LevelError {
		return noopbufferLineNL
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelError)
	} else {
		bl.writeInitialConsole(LevelError)
	}
	return bl
}

func (l *logger) WithBufferLineNLPanic() *bufferLineNL {
	if l.Level() > LevelPanic {
		return noopbufferLineNL
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelPanic)
	} else {
		bl.writeInitialConsole(LevelPanic)
	}
	// exit after the final write so the record is not lost
	bl.panicAfterWrite = true
	return bl
}

func (l *logger) WithBufferLineNLFatal() *bufferLineNL {
	if l.Level() > LevelFatal {
		return noopbufferLineNL
	}
	bl := emptybufferLineNL(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &bl.callerData)
	}
	if bl.jsonMode {
		bl.writeInitialJSON(LevelFatal)
	} else {
		bl.writeInitialConsole(LevelFatal)
	}
	// exit after the final write so the record is not lost
	bl.exitAfterWrite = true
	return bl
}
