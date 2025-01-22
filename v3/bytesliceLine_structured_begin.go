package iqlog

import (
	"os"
	"time"
)

var noopbytesliceLine = &bytesliceLine{}

func emptybytesliceLine(l *logger) *bytesliceLine {
	bsl := bytesliceLinePool.Get().(*bytesliceLine)
	bsl.output = make([]byte, 0)
	bsl.out = l.out
	bsl.jsonMode = l.jsonMode
	return bsl
}

func (bsl *bytesliceLine) writeInitialJSON(level Level) {
	bsl.output = append(bsl.output, '{')
	bsl.AddTime() // adds a trailing comma if it outputs

	// Convert Level to string
	switch level {
	case LevelDebug:
		bsl.output = append(bsl.output, []byte("\"level\":\"debug\",\"message\":\"")...)
	case LevelInfo:
		bsl.output = append(bsl.output, []byte("\"level\":\"info\",\"message\":\"")...)
	case LevelWarn:
		bsl.output = append(bsl.output, []byte("\"level\":\"warn\",\"message\":\"")...)
	case LevelError:
		bsl.output = append(bsl.output, []byte("\"level\":\"error\",\"message\":\"")...)
	case LevelFatal:
		bsl.output = append(bsl.output, []byte("\"level\":\"fatal\",\"message\":\"")...)
	case LevelPanic:
		bsl.output = append(bsl.output, []byte("\"level\":\"panic\",\"message\":\"")...)
	default:
		bsl.output = append(bsl.output, []byte("\"level\":\"unknown\",\"message\":\"")...)
	}

}

func (bsl *bytesliceLine) writeInitialConsole(level Level) {
	if len(bsl.output) == 0 {
		bsl.AddTime()
		if useColour {
			bsl.output = append(bsl.output, ansiColourPrefix(level)...)
		} else {
			bsl.output = append(bsl.output, levelPrefix(level)...)
		}
	}
}

func (bsl *bytesliceLine) AddTime() {
	// if !bsl.logger.IncludeTime {
	// 	return
	// }

	timeNow := time.Now()
	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	if bsl.jsonMode {
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
		if len(bsl.output) > 1 {
			// Use the comma
			bsl.output = append(bsl.output, prefixArr[:37]...)
		} else {
			// Skip the comma
			bsl.output = append(bsl.output, prefixArr[1:37]...)
		}

	} else {
		var consolePrefixFull = [37]byte{
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
		bsl.output = append(bsl.output, consolePrefixFull[:28]...)
	}
}

func (l *logger) WithByteSliceLineTrace() *bytesliceLine {
	if l.Level > LevelTrace {
		return noopbytesliceLine
	}
	bsl := emptybytesliceLine(l)
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
