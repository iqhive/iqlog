package iqlog

import (
	"os"
	"time"
)

var nooppreallocLine = &preallocLine{}

func emptypreallocLine(l *logger) *preallocLine {
	pal := preallocLinePool.Get().(*preallocLine)
	pal.bytesUsed = 0
	pal.jsonMode = l.jsonMode
	pal.out = l.out
	pal.includeTime = l.IncludeTime
	return pal
}

func (pal *preallocLine) writeInitialJSON(level Level) {
	pal.output[0] = '{'
	pal.bytesUsed++
	pal.AddTime()

	// Convert Level to string
	switch level {
	case LevelDebug:
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"level\":\"debug\"")
	case LevelInfo:
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"level\":\"info\"")
	case LevelWarn:
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"level\":\"warn\"")
	case LevelError:
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"level\":\"error\"")
	case LevelFatal:
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"level\":\"fatal\"")
	case LevelPanic:
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"level\":\"panic\"")
	default:
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"level\":\"unknown\"")
	}
}

func (pal *preallocLine) writeInitialConsole(level Level) {
	pal.AddTime()
	if useColour {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, string(ansiColourPrefix(level)))
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, string(levelPrefix(level)))
	}
}

func (pal *preallocLine) AddTime() {
	if !pal.includeTime {
		return
	}

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

		if pal.bytesUsed > 1 {
			// Use the comma
			copy(pal.output[pal.bytesUsed:], prefixArr[:37])
			pal.bytesUsed += 37
		} else {
			// Skip the comma
			copy(pal.output[pal.bytesUsed:], prefixArr[1:37])
			pal.bytesUsed += 36
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
		copy(pal.output[pal.bytesUsed:], consolePrefixFull[:29])
		pal.bytesUsed += 29
	}
}

// func (pal *preallocLine) AddTime2() {
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
// 		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, `"time":"`)
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

func (l *logger) WithPreallocLineTrace() *preallocLine {
	if l.Level > LevelTrace {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelTrace)
	} else {
		pal.writeInitialConsole(LevelTrace)
	}
	return pal
}

func (l *logger) WithPreallocLineDebug() *preallocLine {
	if l.Level > LevelDebug {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelDebug)
	} else {
		pal.writeInitialConsole(LevelDebug)
	}
	return pal
}

func (l *logger) WithPreallocLineInfo() *preallocLine {
	if l.Level > LevelInfo {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
	// slice := fixedSlicePool2.Get().(*[maxLineLen]byte)
	// pal.output = slice
	if pal.jsonMode {
		pal.writeInitialJSON(LevelInfo)
	} else {
		pal.writeInitialConsole(LevelInfo)
	}
	return pal
}

func (l *logger) WithPreallocLineWarn() *preallocLine {
	if l.Level > LevelWarn {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelWarn)
	} else {
		pal.writeInitialConsole(LevelWarn)
	}
	return pal
}

func (l *logger) WithPreallocLineError() *preallocLine {
	if l.Level > LevelError {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelError)
	} else {
		pal.writeInitialConsole(LevelError)
	}
	return pal
}

func (l *logger) WithPreallocLinePanic() *preallocLine {
	if l.Level > LevelPanic {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelPanic)
	} else {
		pal.writeInitialConsole(LevelPanic)
	}
	// pal.logger.Flush()
	os.Exit(1)
	return pal
}

func (l *logger) WithPreallocLineFatal() *preallocLine {
	if l.Level > LevelFatal {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
	if pal.jsonMode {
		pal.writeInitialJSON(LevelFatal)
	} else {
		pal.writeInitialConsole(LevelFatal)
	}
	// pal.logger.Flush()
	os.Exit(1)
	return pal
}

// func (l *logger) Str(name string, s string) *preallocLine {
// 	pal := &preallocLine{
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

// func (l *logger) Any(name string, v any) *preallocLine {
// 	pal := &preallocLine{
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
