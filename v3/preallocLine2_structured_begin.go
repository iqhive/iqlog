package iqlog

import (
	"os"
	"runtime"
	"strings"
	"time"
)

var nooppreallocLine2 = &preallocLine2{}

func emptypreallocLine2(l *logger) *preallocLine2 {
	pal := preallocLine2Pool.Get().(*preallocLine2)
	pal.jsonMode = l.jsonMode
	pal.out = l.out
	pal.bytesUsed = 0
	pal.includeTime = l.IncludeTime
	pal.captureCallers = l.captureCallers
	// zero the output
	// pal.output = pal.output[:0]
	return pal
}

func (pal *preallocLine2) writeInitialJSON(level Level) {
	pal.output[0] = '{'
	pal.bytesUsed++
	pal.AddTime()

	// Convert Level to string
	switch level {
	case LevelDebug:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"debug\"")
	case LevelInfo:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"info\"")
	case LevelWarn:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"warn\"")
	case LevelError:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"error\"")
	case LevelFatal:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"fatal\"")
	case LevelPanic:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"panic\"")
	default:
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"level\":\"unknown\"")
	}
	pal.AddCallers()
}

func (pal *preallocLine2) AddCallers() {
	if !pal.captureCallers {
		return
	}
	// Get more stack frames to ensure we capture enough context
	var callers [32]uintptr
	n := runtime.Callers(1, callers[:]) // Changed from 0 to 1 to skip this frame
	frames := runtime.CallersFrames(callers[:n])

	// Skip frames until we find the actual caller
	var frame runtime.Frame
	more := true
	foundFrame := false

	for more {
		frame, more = frames.Next()
		// Skip internal logging packages and runtime frames
		skipFrame := false
		for _, skip := range FunctionsToSkip {
			if strings.Contains(frame.Function, skip) {
				skipFrame = true
				break
			}
		}
		if skipFrame {
			continue
		}
		foundFrame = true
		break
	}

	if foundFrame {
		fileOffsetLast := 0
		fileOffset2ndLast := 0
		for i := range frame.File {
			if frame.File[i] == '/' {
				fileOffset2ndLast = fileOffsetLast
				fileOffsetLast = i
			}
		}
		if pal.jsonMode {
			// json mode
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `,"func":"`)
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, frame.Function)
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `","file":"`)
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, frame.File[fileOffset2ndLast:])
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `"`)
		} else {
			// console mode
			if useColour {
				pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\x1b[32m[")
			} else {
				pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `[`)
			}
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, frame.Function)
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ` `)
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, frame.File[fileOffset2ndLast:])
			if useColour {
				pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "]\x1b[0m ")
			} else {
				pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `] `)
			}
		}
	} else {
		// Fallback if we couldn't find a suitable frame
		// originText = fmt.Sprintf("[%v]", l.applicationName)
		// fmt.Printf("[%v %v:%v]\n", name, file, frame.Line)
	}
}

func (pal *preallocLine2) writeInitialConsole(level Level) {
	// fmt.Printf("before time output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))
	pal.AddTime()
	// fmt.Printf("after time output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))
	if useColour {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, string(ansiColourPrefix(level)))
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, string(levelPrefix(level)))
	}
	// fmt.Printf("after writeInitialConsole: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))
	pal.AddCallers()
}

func (pal *preallocLine2) AddTime() {
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

		// fmt.Printf("during time output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))

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
	// fmt.Printf("Info1 output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))

	if pal.jsonMode {
		pal.writeInitialJSON(LevelInfo)
	} else {
		pal.writeInitialConsole(LevelInfo)
	}
	// fmt.Printf("Info2 output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))

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
