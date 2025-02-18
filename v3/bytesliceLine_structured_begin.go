package iqlog

import (
	"os"
	"time"
)

var noopbytesliceLine = &bytesliceLine{}

func emptybytesliceLine(l *logger) *bytesliceLine {
	bsl := bytesliceLinePool.Get().(*bytesliceLine)
	// bsl.output = make([]byte, 0)
	bsl.output = bsl.output[:0]
	bsl.out = l.out
	bsl.jsonMode = l.jsonMode
	bsl.includeTime = l.IncludeTime
	bsl.captureCaller = l.CallerDepth
	bsl.callerData.callerFuncLen = 0
	bsl.callerData.callerFileLen = 0
	return bsl
}

func (bsl *bytesliceLine) writeInitialJSON(level Level) {
	if bsl.includeTime {
		AddTimeJSONAppend(time.Now(), &bsl.output)
	} else {
		bsl.output = append(bsl.output, '{')
	}

	// Convert Level to string
	switch level {
	case LevelDebug:
		bsl.output = append(bsl.output, []byte("\"level\":\"debug\"")...)
	case LevelInfo:
		bsl.output = append(bsl.output, []byte("\"level\":\"info\"")...)
	case LevelWarn:
		bsl.output = append(bsl.output, []byte("\"level\":\"warn\"")...)
	case LevelError:
		bsl.output = append(bsl.output, []byte("\"level\":\"error\"")...)
	case LevelFatal:
		bsl.output = append(bsl.output, []byte("\"level\":\"fatal\"")...)
	case LevelPanic:
		bsl.output = append(bsl.output, []byte("\"level\":\"panic\"")...)
	default:
		bsl.output = append(bsl.output, []byte("\"level\":\"unknown\"")...)
	}

	bsl.AddCallers()
}

func (bsl *bytesliceLine) AddCallers() {
	if bsl.captureCaller == 0 || bsl.callerData.callerFuncLen == 0 {
		return
	}

	if bsl.jsonMode {
		// json mode
		bsl.output = append(bsl.output, []byte(`,"func":"`)...)
		bsl.output = append(bsl.output, bsl.callerData.callerFunc[:bsl.callerData.callerFuncLen]...)
		bsl.output = append(bsl.output, []byte(`","file":"`)...)
		bsl.output = append(bsl.output, bsl.callerData.callerFile[:bsl.callerData.callerFileLen]...)
		bsl.output = append(bsl.output, []byte(`"`)...)
	} else {
		// console mode
		if useColour {
			bsl.output = append(bsl.output, []byte("\x1b[32m[")...)
		} else {
			bsl.output = append(bsl.output, []byte(`[`)...)
		}
		bsl.output = append(bsl.output, bsl.callerData.callerFunc[:bsl.callerData.callerFuncLen]...)
		bsl.output = append(bsl.output, []byte(` `)...)
		bsl.output = append(bsl.output, bsl.callerData.callerFile[:bsl.callerData.callerFileLen]...)
		if useColour {
			bsl.output = append(bsl.output, []byte("]\x1b[0m ")...)
		} else {
			bsl.output = append(bsl.output, []byte(`] `)...)
		}
	}
}

func (bsl *bytesliceLine) writeInitialConsole(level Level) {
	if bsl.includeTime {
		AddTimeConsoleAppend(time.Now(), &bsl.output)
	}
	if useColour {
		bsl.output = append(bsl.output, ansiColourPrefix(level)...)
	} else {
		bsl.output = append(bsl.output, levelPrefix(level)...)
	}
	bsl.AddCallers()
}

func (l *logger) WithByteSliceLineTrace() *bytesliceLine {
	if l.Level > LevelTrace {
		return noopbytesliceLine
	}
	bsl := emptybytesliceLine(l)

	if l.CallerDepth > 0 {
		var pc PC
		caller1(l.CallerDepth+1, &pc, 1, 1)
		fillCallerData(pc, &bsl.callerData)
	}

	if bsl.jsonMode {
		bsl.writeInitialJSON(LevelTrace)
	} else {
		bsl.writeInitialConsole(LevelTrace)
	}
	return bsl
}

func (l *logger) WithByteSliceLineDebug() *bytesliceLine {
	if l.Level > LevelDebug {
		return noopbytesliceLine
	}
	bsl := emptybytesliceLine(l)

	if l.CallerDepth > 0 {
		var pc PC
		caller1(l.CallerDepth+1, &pc, 1, 1)
		fillCallerData(pc, &bsl.callerData)
	}

	if bsl.jsonMode {
		bsl.writeInitialJSON(LevelDebug)
	} else {
		bsl.writeInitialConsole(LevelDebug)
	}
	return bsl
}

func (l *logger) WithByteSliceLineInfo() *bytesliceLine {
	if l.Level > LevelInfo {
		return noopbytesliceLine
	}

	bsl := emptybytesliceLine(l)

	if l.CallerDepth > 0 {
		var pc PC
		caller1(l.CallerDepth+1, &pc, 1, 1)
		fillCallerData(pc, &bsl.callerData)
	}

	if bsl.jsonMode {
		bsl.writeInitialJSON(LevelInfo)
	} else {
		bsl.writeInitialConsole(LevelInfo)
	}
	return bsl
}

func (l *logger) WithByteSliceLineWarn() *bytesliceLine {
	if l.Level > LevelWarn {
		return noopbytesliceLine
	}
	bsl := emptybytesliceLine(l)

	if l.CallerDepth > 0 {
		var pc PC
		caller1(l.CallerDepth+1, &pc, 1, 1)
		fillCallerData(pc, &bsl.callerData)
	}

	if bsl.jsonMode {
		bsl.writeInitialJSON(LevelWarn)
	} else {
		bsl.writeInitialConsole(LevelWarn)
	}
	return bsl
}

func (l *logger) WithByteSliceLineError() *bytesliceLine {
	if l.Level > LevelError {
		return noopbytesliceLine
	}
	bsl := emptybytesliceLine(l)

	if l.CallerDepth > 0 {
		var pc PC
		caller1(l.CallerDepth+1, &pc, 1, 1)
		fillCallerData(pc, &bsl.callerData)
	}

	if bsl.jsonMode {
		bsl.writeInitialJSON(LevelError)
	} else {
		bsl.writeInitialConsole(LevelError)
	}
	return bsl
}

func (l *logger) WithByteSliceLinePanic() *bytesliceLine {
	if l.Level > LevelPanic {
		return noopbytesliceLine
	}
	bsl := emptybytesliceLine(l)

	if l.CallerDepth > 0 {
		var pc PC
		caller1(l.CallerDepth+1, &pc, 1, 1)
		fillCallerData(pc, &bsl.callerData)
	}

	if bsl.jsonMode {
		bsl.writeInitialJSON(LevelPanic)
	} else {
		bsl.writeInitialConsole(LevelPanic)
	}
	// bsl.logger.Flush()
	os.Exit(1)
	return bsl
}

func (l *logger) WithByteSliceLineFatal() *bytesliceLine {
	if l.Level > LevelFatal {
		return noopbytesliceLine
	}
	bsl := emptybytesliceLine(l)

	if l.CallerDepth > 0 {
		var pc PC
		caller1(l.CallerDepth+1, &pc, 1, 1)
		fillCallerData(pc, &bsl.callerData)
	}

	if bsl.jsonMode {
		bsl.writeInitialJSON(LevelFatal)
	} else {
		bsl.writeInitialConsole(LevelFatal)
	}
	// bsl.logger.Flush()
	os.Exit(1)
	return bsl
}

// func (l *logger) Str(name string, s string) *bytesliceLine {
// 	bsl := &bytesliceLine{
// 		logger: l,
// 		level:  LevelUnknown,
// 	}
// 	if bsl.jsonMode {
// 		bsl.writeInitialJSON(LevelUnknown)
// 	} else {
// 		bsl.writeInitialConsole(LevelUnknown)
// 	}
// 	return bsl
// }

// func (l *logger) Any(name string, v any) *bytesliceLine {
// 	bsl := &bytesliceLine{
// 		logger: l,
// 		level:  LevelUnknown,
// 	}
// 	if bsl.jsonMode {
// 		bsl.writeInitialJSON(LevelUnknown)
// 	} else {
// 		bsl.writeInitialConsole(LevelUnknown)
// 	}
// 	return bsl
// }
