package iqlog

import (
	"os"
	"time"
)

var nooppreallocLine2 = &preallocLine2{}

func emptypreallocLine2(l *logger) *preallocLine2 {
	pal := preallocLine2Pool.Get().(*preallocLine2)
	pal.jsonMode = l.jsonMode
	pal.out = l.out
	pal.bytesUsed = 0
	return pal
}

func (pal *preallocLine2) writeInitialJSON(level Level) {
	pal.output[0] = '{'
	pal.bytesUsed++
	pal.AddTime()

	// Convert Level to string
	switch level {
	case LevelDebug:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"debug\",\"message\":\"")
	case LevelInfo:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"info\",\"message\":\"")
	case LevelWarn:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"warn\",\"message\":\"")
	case LevelError:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"error\",\"message\":\"")
	case LevelFatal:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"fatal\",\"message\":\"")
	case LevelPanic:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"panic\",\"message\":\"")
	default:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"unknown\",\"message\":\"")
	}
}

func (pal *preallocLine2) writeInitialConsole(level Level) {
	if pal.bytesUsed == 0 {
		pal.AddTime()
		if useColour {
			pal.bytesUsed += safeOutputCopy(pal.output, 0, string(ansiColourPrefix(level)))
		} else {
			pal.bytesUsed += safeOutputCopy(pal.output, 0, string(levelPrefix(level)))
		}
	}
}

func (pal *preallocLine2) AddTime() {
	// if !pal.logger.IncludeTime {
	// 	return
	// }

	timeNow := time.Now()
	year, month, day := timeNow.Date()
	hour, min, sec := timeNow.Clock()
	usec := timeNow.Nanosecond() / 1000

	if pal.jsonMode {
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
		if len(pal.output) > 1 {
			// Use the comma
			copy(pal.output[pal.bytesUsed:], prefixArr[:37])
			pal.bytesUsed += 37
		} else {
			// Skip the comma
			copy(pal.output[pal.bytesUsed:], prefixArr[1:37])
			pal.bytesUsed += 36
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
		copy(pal.output[pal.bytesUsed:], consolePrefixFull[:28])
		pal.bytesUsed += 28
	}
}

// func (pal *preallocLine2) AddTime2() {
// 	// if !pal.logger.IncludeTime {
// 	// 	return
// 	// }

// 	timeNow := time.Now()
// 	year, month, day := timeNow.Date()
// 	hour, min, sec := timeNow.Clock()
// 	usec := timeNow.Nanosecond() / 1000
// 	if pal.jsonMode {
// 		if pal.bytesUsed > 1 {
// 			pal.output[pal.bytesUsed] = ','
// 			pal.bytesUsed++
// 		}
// 		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `"time":"`)
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(year), 4)
// 		pal.output[pal.bytesUsed] = '-'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(month), 2)
// 		pal.output[pal.bytesUsed] = '-'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(day), 2)
// 		pal.output[pal.bytesUsed] = 'T'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(hour), 2)
// 		pal.output[pal.bytesUsed] = ':'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(min), 2)
// 		pal.output[pal.bytesUsed] = ':'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(sec), 2)
// 		pal.output[pal.bytesUsed] = '.'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(usec), 6)
// 		pal.output[pal.bytesUsed] = '"'
// 		pal.bytesUsed++
// 	} else {
// 		if pal.bytesUsed > 0 {
// 			pal.output[pal.bytesUsed] = ' '
// 			pal.bytesUsed++
// 		}
// 		pal.output[pal.bytesUsed] = '['
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(year), 4)
// 		pal.output[pal.bytesUsed] = '-'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(month), 2)
// 		pal.output[pal.bytesUsed] = '-'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(day), 2)
// 		pal.output[pal.bytesUsed] = 'T'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(hour), 2)
// 		pal.output[pal.bytesUsed] = ':'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(min), 2)
// 		pal.output[pal.bytesUsed] = ':'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(sec), 2)
// 		pal.output[pal.bytesUsed] = '.'
// 		pal.bytesUsed++
// 		pal.bytesUsed += setIntBytes(pal.output[pal.bytesUsed:], int64(usec), 6)
// 		pal.output[pal.bytesUsed] = ']'
// 		pal.bytesUsed++
// 		pal.output[pal.bytesUsed] = ' '
// 		pal.bytesUsed++
// 	}

// }

func (l *logger) WithPreallocLine2Trace() *preallocLine2 {
	if l.Level > LevelTrace {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelTrace)
	} else {
		pal.writeInitialConsole(LevelTrace)
	}
	return pal
}

func (l *logger) WithPreallocLine2Debug() *preallocLine2 {
	if l.Level > LevelDebug {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelDebug)
	} else {
		pal.writeInitialConsole(LevelDebug)
	}
	return pal
}

func (l *logger) WithPreallocLine2Info() *preallocLine2 {
	if l.Level > LevelInfo {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelInfo)
		pal.writeFinalJSON("")
	} else {
		pal.writeInitialConsole(LevelInfo)
		pal.writeFinalConsole("")
	}
	return pal
}

func (l *logger) WithPreallocLine2Warn() *preallocLine2 {
	if l.Level > LevelWarn {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelWarn)
	} else {
		pal.writeInitialConsole(LevelWarn)
	}
	return pal
}

func (l *logger) WithPreallocLine2Error() *preallocLine2 {
	if l.Level > LevelError {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelError)
	} else {
		pal.writeInitialConsole(LevelError)
	}
	return pal
}

func (l *logger) WithPreallocLine2Panic() *preallocLine2 {
	if l.Level > LevelPanic {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelPanic)
	} else {
		pal.writeInitialConsole(LevelPanic)
	}
	// pal.logger.Flush()
	os.Exit(1)
	return pal
}

func (l *logger) WithPreallocLine2Fatal() *preallocLine2 {
	if l.Level > LevelFatal {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelFatal)
	} else {
		pal.writeInitialConsole(LevelFatal)
	}
	// pal.logger.Flush()
	os.Exit(1)
	return pal
}

// func (l *logger) Str(name string, s string) *preallocLine2 {
// 	pal := &preallocLine2{
// 		logger: l,
// 		level:  LevelUnknown,
// 	}
// 	if pal.jsonMode {
// 		pal.writeInitialJSON(LevelUnknown)
// 	} else {
// 		pal.writeInitialConsole(LevelUnknown)
// 	}
// 	return pal
// }

// func (l *logger) Any(name string, v any) *preallocLine2 {
// 	pal := &preallocLine2{
// 		logger: l,
// 		level:  LevelUnknown,
// 	}
// 	if pal.jsonMode {
// 		pal.writeInitialJSON(LevelUnknown)
// 	} else {
// 		pal.writeInitialConsole(LevelUnknown)
// 	}
// 	return pal
// }
