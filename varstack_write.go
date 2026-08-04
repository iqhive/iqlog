package iqlog

import (
	"fmt"
	"strconv"
	"time"
)

func (vs *varStack) writeFinalConsole(msg string, args ...interface{}) {
	var output []byte

	// Add time if enabled
	if vs.includeTime {
		AddTimeConsoleAppend(time.Now(), &output)
	}

	// Add level prefix
	if useColour {
		output = append(output, ansiColourPrefix(vs.level)...)
	} else {
		output = append(output, levelPrefix(vs.level)...)
	}

	// Add caller info if configured
	// vs.AddCallers() - skipping for now as it's not in the structure

	// Add variable key=value pairs first
	for i := 0; i < vs.varCount; i++ {
		switch v := vs.varValue[i].(type) {
		case string:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, v...)
			output = append(output, ' ')
		case int:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, strconv.Itoa(v)...)
			output = append(output, ' ')
		case float32:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, strconv.FormatFloat(float64(v), 'f', 6, 32)...)
			output = append(output, ' ')
		case float64:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, strconv.FormatFloat(v, 'f', 6, 64)...)
			output = append(output, ' ')
		default:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, fmt.Sprintf("%v", v)...)
			output = append(output, ' ')
		}
	}

	// Add the message
	output = append(output, msg...)

	// Add any additional args
	for _, thisarg := range args {
		output = append(output, ' ')

		switch thisarg := thisarg.(type) {
		case string:
			output = append(output, thisarg...)
		case int:
			output = append(output, strconv.Itoa(thisarg)...)
		case int32:
			output = append(output, strconv.Itoa(int(thisarg))...)
		case uint32:
			output = append(output, strconv.FormatUint(uint64(thisarg), 10)...)
		case int64:
			output = append(output, strconv.FormatInt(thisarg, 10)...)
		case uint64:
			output = append(output, strconv.FormatUint(thisarg, 10)...)
		case float32:
			output = append(output, strconv.FormatFloat(float64(thisarg), 'f', -1, 32)...)
		case float64:
			output = append(output, strconv.FormatFloat(thisarg, 'f', -1, 64)...)
		case bool:
			output = append(output, strconv.FormatBool(thisarg)...)
		default:
			output = append(output, fmt.Sprintf("%v", thisarg)...)
		}
	}

	// Add newline
	output = append(output, '\n')

	// Write to output
	vs.out.Write(output)

	varStackPool.Put(vs)
}

func (vs *varStack) writeFinalConsoleF(format string, args ...interface{}) {
	var output []byte

	// Add time if enabled
	if vs.includeTime {
		AddTimeConsoleAppend(time.Now(), &output)
	}

	// Add level prefix
	if useColour {
		output = append(output, ansiColourPrefix(vs.level)...)
	} else {
		output = append(output, levelPrefix(vs.level)...)
	}

	// Add variable key=value pairs first
	for i := 0; i < vs.varCount; i++ {
		switch v := vs.varValue[i].(type) {
		case string:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, v...)
			output = append(output, ' ')
		case int:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, strconv.Itoa(v)...)
			output = append(output, ' ')
		case float32:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, strconv.FormatFloat(float64(v), 'f', 6, 32)...)
			output = append(output, ' ')
		case float64:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, strconv.FormatFloat(v, 'f', 6, 64)...)
			output = append(output, ' ')
		default:
			output = append(output, vs.varName[i]...)
			output = append(output, '=')
			output = append(output, fmt.Sprintf("%v", v)...)
			output = append(output, ' ')
		}
	}

	// Format the message
	formatted := fmt.Sprintf(format, args...)
	output = append(output, formatted...)

	// Add newline
	output = append(output, '\n')

	// Write to output
	vs.out.Write(output)

	varStackPool.Put(vs)
}

func (vs *varStack) writeFinalJSON(msg string, args ...interface{}) {
	var output []byte

	// Add time if enabled (AddTimeJSONAppend includes opening brace)
	if vs.includeTime {
		AddTimeJSONAppend(time.Now(), &output)
	} else {
		output = append(output, '{')
	}

	// Add level
	switch vs.level {
	case LevelDebug:
		output = append(output, []byte("\"level\":\"debug\"")...)
	case LevelInfo:
		output = append(output, []byte("\"level\":\"info\"")...)
	case LevelWarn:
		output = append(output, []byte("\"level\":\"warn\"")...)
	case LevelError:
		output = append(output, []byte("\"level\":\"error\"")...)
	case LevelFatal:
		output = append(output, []byte("\"level\":\"fatal\"")...)
	case LevelPanic:
		output = append(output, []byte("\"level\":\"panic\"")...)
	default:
		output = append(output, []byte("\"level\":\"unknown\"")...)
	}

	// Add variable key=value pairs
	for i := 0; i < vs.varCount; i++ {
		output = append(output, ',')
		output = append(output, '"')
		output = appendJSONEscaped(output, vs.varName[i])
		output = append(output, '"', ':')

		switch v := vs.varValue[i].(type) {
		case string:
			output = append(output, '"')
			output = appendJSONEscaped(output, v)
			output = append(output, '"')
		case int:
			output = append(output, strconv.Itoa(v)...)
		case float32:
			output = appendJSONFloat(output, float64(v), 32)
		case float64:
			output = appendJSONFloat(output, v, 64)
		default:
			output = append(output, '"')
			output = appendJSONEscaped(output, fmt.Sprintf("%v", v))
			output = append(output, '"')
		}
	}

	// Add message
	output = append(output, []byte(",\"message\":\"")...)
	output = appendJSONEscaped(output, msg)

	// Add any additional args
	for _, thisarg := range args {
		output = append(output, ' ')

		switch thisarg := thisarg.(type) {
		case string:
			output = appendJSONEscaped(output, thisarg)
		case int:
			output = append(output, strconv.Itoa(thisarg)...)
		case int32:
			output = append(output, strconv.Itoa(int(thisarg))...)
		case uint32:
			output = append(output, strconv.FormatUint(uint64(thisarg), 10)...)
		case int64:
			output = append(output, strconv.FormatInt(thisarg, 10)...)
		case uint64:
			output = append(output, strconv.FormatUint(thisarg, 10)...)
		case float32:
			output, _ = appendfastFloatFill(output, float64(thisarg), 6)
		case float64:
			output, _ = appendfastFloatFill(output, thisarg, 6)
		case bool:
			output = append(output, strconv.FormatBool(thisarg)...)
		default:
			output = appendJSONEscaped(output, fmt.Sprintf("%v", thisarg))
		}
	}

	// Close JSON and add newline
	output = append(output, []byte("\"}\n")...)

	// Write to output
	vs.out.Write(output)

	varStackPool.Put(vs)
}

func (vs *varStack) writeFinalJSONF(format string, args ...interface{}) {
	var output []byte

	// Add time if enabled (AddTimeJSONAppend includes opening brace)
	if vs.includeTime {
		AddTimeJSONAppend(time.Now(), &output)
	} else {
		output = append(output, '{')
	}

	// Add level
	switch vs.level {
	case LevelDebug:
		output = append(output, []byte("\"level\":\"debug\"")...)
	case LevelInfo:
		output = append(output, []byte("\"level\":\"info\"")...)
	case LevelWarn:
		output = append(output, []byte("\"level\":\"warn\"")...)
	case LevelError:
		output = append(output, []byte("\"level\":\"error\"")...)
	case LevelFatal:
		output = append(output, []byte("\"level\":\"fatal\"")...)
	case LevelPanic:
		output = append(output, []byte("\"level\":\"panic\"")...)
	default:
		output = append(output, []byte("\"level\":\"unknown\"")...)
	}

	// Add variable key=value pairs
	for i := 0; i < vs.varCount; i++ {
		output = append(output, ',')
		output = append(output, '"')
		output = appendJSONEscaped(output, vs.varName[i])
		output = append(output, '"', ':')

		switch v := vs.varValue[i].(type) {
		case string:
			output = append(output, '"')
			output = appendJSONEscaped(output, v)
			output = append(output, '"')
		case int:
			output = append(output, strconv.Itoa(v)...)
		case float32:
			output = appendJSONFloat(output, float64(v), 32)
		case float64:
			output = appendJSONFloat(output, v, 64)
		default:
			output = append(output, '"')
			output = appendJSONEscaped(output, fmt.Sprintf("%v", v))
			output = append(output, '"')
		}
	}

	// Add message with formatting
	output = append(output, []byte(",\"message\":\"")...)
	formatted := fmt.Sprintf(format, args...)
	output = appendJSONEscaped(output, formatted)

	// Close JSON and add newline
	output = append(output, []byte("\"}\n")...)

	// Write to output
	vs.out.Write(output)

	varStackPool.Put(vs)
}

// func (vs *varStack) AddCallers() {
// 	if !vs.captureCallers {
// 		return
// 	}
// 	// Get more stack frames to ensure we capture enough context
// 	var callers [32]uintptr
// 	n := runtime.Callers(1, callers[:]) // Changed from 0 to 1 to skip this frame
// 	frames := runtime.CallersFrames(callers[:n])

// 	// Skip frames until we find the actual caller
// 	var frame runtime.Frame
// 	more := true
// 	foundFrame := false

// 	for more {
// 		frame, more = frames.Next()
// 		// Skip internal logging packages and runtime frames
// 		skipFrame := false
// 		for _, skip := range FunctionsToSkip {
// 			if strings.Contains(frame.Function, skip) {
// 				skipFrame = true
// 				break
// 			}
// 		}
// 		if skipFrame {
// 			continue
// 		}
// 		foundFrame = true
// 		break
// 	}

// 	if foundFrame {
// 		fileOffsetLast := 0
// 		fileOffset2ndLast := 0
// 		for i := range frame.File {
// 			if frame.File[i] == '/' {
// 				fileOffset2ndLast = fileOffsetLast
// 				fileOffsetLast = i
// 			}
// 		}
// 		if vs.jsonMode {
// 			// json mode
// 			vs.output = append(vs.output, []byte(`,"func":"`)...)
// 			vs.output = append(vs.output, []byte(frame.Function)...)
// 			vs.output = append(vs.output, []byte(`","file":"`)...)
// 			vs.output = append(vs.output, []byte(frame.File[fileOffset2ndLast:])...)
// 			vs.output = append(vs.output, []byte(`"`)...)
// 		} else {
// 			// console mode
// 			if useColour {
// 				vs.output = append(vs.output, []byte("\x1b[32m[")...)
// 			} else {
// 				vs.output = append(vs.output, []byte(`[`)...)
// 			}
// 			vs.output = append(vs.output, []byte(frame.Function)...)
// 			vs.output = append(vs.output, []byte(` `)...)
// 			vs.output = append(vs.output, []byte(frame.File[fileOffset2ndLast:])...)
// 			if useColour {
// 				vs.output = append(vs.output, []byte("]\x1b[0m ")...)
// 			} else {
// 				vs.output = append(vs.output, []byte(`] `)...)
// 			}
// 		}
// 	} else {
// 		// Fallback if we couldn't find a suitable frame
// 		// originText = fmt.Sprintf("[%v]", l.applicationName)
// 		// fmt.Printf("[%v %v:%v]\n", name, file, frame.Line)
// 	}
// }
