package iqlog

func (vs *varStack) writeFinalConsole(msg string, args ...interface{}) {
	// vs.AddTime()
	// if useColour {
	// 	vs.output = append(vs.output, ansiColourPrefix(level)...)
	// } else {
	// 	vs.output = append(vs.output, levelPrefix(level)...)
	// }
	// vs.AddCallers()

	// //put the message
	// vs.output = append(vs.output, msg...)

	// // followed by args
	// for _, thisarg := range args {
	// 	vs.output = append(vs.output, ' ')

	// 	switch thisarg := thisarg.(type) {
	// 	case string:
	// 		vs.output = append(vs.output, thisarg...)
	// 	case int:
	// 		vs.output = append(vs.output, strconv.Itoa(thisarg)...)
	// 	case int32:
	// 		vs.output = append(vs.output, strconv.Itoa(int(thisarg))...)
	// 	case uint32:
	// 		vs.output = append(vs.output, strconv.FormatUint(uint64(thisarg), 10)...)
	// 	case int64:
	// 		vs.output = append(vs.output, strconv.FormatInt(thisarg, 10)...)
	// 	case uint64:
	// 		vs.output = append(vs.output, strconv.FormatUint(thisarg, 10)...)
	// 	case float32:
	// 		vs.output = append(vs.output, strconv.FormatFloat(float64(thisarg), 'f', -1, 32)...)
	// 	case float64:
	// 		vs.output = append(vs.output, strconv.FormatFloat(thisarg, 'f', -1, 64)...)
	// 	case bool:
	// 		vs.output = append(vs.output, strconv.FormatBool(thisarg)...)
	// 	default:
	// 		vs.output = append(vs.output, fmt.Sprintf("%v", thisarg)...)
	// 	}
	// }
	// // if vs.logger.newLine {
	// vs.output = append(vs.output, '\n')
	// // }

	// vs.out.Write(vs.output)

	varStackPool.Put(vs)
}

func (vs *varStack) writeFinalConsoleF(format string, args ...interface{}) {
	// vs.AddTime()
	// if useColour {
	// 	vs.output = append(vs.output, ansiColourPrefix(level)...)
	// } else {
	// 	vs.output = append(vs.output, levelPrefix(level)...)
	// }
	// vs.AddCallers()

	// // ba := baPool.Get().(*bytesAppender)
	// // ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	// // fmt.Fprintf(ba, format, args...)
	// // vs.output = append(vs.output, ba.Bytes...)
	// // baPool.Put(ba)

	// bia := biapool.Get().(*byteIndexAppender)
	// bia.Index = 0
	// fmt.Fprintf(bia, format, args...)
	// vs.output = append(vs.output, bia.Bytes[:bia.Index]...)
	// biapool.Put(bia)

	// // if vs.logger.newLine {
	// vs.output = append(vs.output, '\n')
	// // }

	// vs.out.Write(vs.output)

	varStackPool.Put(vs)
}

func (vs *varStack) writeFinalJSON(msg string, args ...interface{}) {
	// vs.output = append(vs.output, '{')

	// vs.AddTime() // adds a trailing comma if it outputs

	// // Convert Level to string
	// switch level {
	// case LevelDebug:
	// 	vs.output = append(vs.output, []byte("\"level\":\"debug\"")...)
	// case LevelInfo:
	// 	vs.output = append(vs.output, []byte("\"level\":\"info\"")...)
	// case LevelWarn:
	// 	vs.output = append(vs.output, []byte("\"level\":\"warn\"")...)
	// case LevelError:
	// 	vs.output = append(vs.output, []byte("\"level\":\"error\"")...)
	// case LevelFatal:
	// 	vs.output = append(vs.output, []byte("\"level\":\"fatal\"")...)
	// case LevelPanic:
	// 	vs.output = append(vs.output, []byte("\"level\":\"panic\"")...)
	// default:
	// 	vs.output = append(vs.output, []byte("\"level\":\"unknown\"")...)
	// }

	// vs.AddCallers()

	// // write JSON closer
	// vs.output = append(vs.output, []byte(",\"message\":\"")...)

	// vs.output = append(vs.output, msg...)

	// // followed by args
	// for _, thisarg := range args {
	// 	vs.output = append(vs.output, ' ')

	// 	switch thisarg := thisarg.(type) {
	// 	case string:
	// 		vs.output = append(vs.output, thisarg...)
	// 	case int:
	// 		vs.output = append(vs.output, strconv.Itoa(thisarg)...)
	// 	case int32:
	// 		vs.output = append(vs.output, strconv.Itoa(int(thisarg))...)
	// 	case uint32:
	// 		vs.output = append(vs.output, strconv.FormatUint(uint64(thisarg), 10)...)
	// 	case int64:
	// 		vs.output = append(vs.output, strconv.FormatInt(thisarg, 10)...)
	// 	case uint64:
	// 		vs.output = append(vs.output, strconv.FormatUint(thisarg, 10)...)
	// 	case float32:
	// 		vs.output, _ = appendfastFloatFill(vs.output, float64(thisarg), 6)
	// 	case float64:
	// 		vs.output, _ = appendfastFloatFill(vs.output, thisarg, 6)
	// 	case bool:
	// 		vs.output = append(vs.output, strconv.FormatBool(thisarg)...)
	// 	default:
	// 		vs.output = append(vs.output, fmt.Sprintf("%v", thisarg)...)
	// 	}
	// }

	// // if vs.newLine {
	// vs.output = append(vs.output, []byte("\"}\n")...)
	// // } else {
	// // 	vs.output = append(vs.output, []byte("\"}")...)
	// // }

	// vs.out.Write(vs.output)

	varStackPool.Put(vs)
}

func (vs *varStack) writeFinalJSONF(format string, args ...interface{}) {
	// vs.output = append(vs.output, []byte(",\"message\":\"")...)

	// ba := baPool.Get().(*bytesAppender)
	// ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	// fmt.Fprintf(ba, format, args...)
	// vs.output = append(vs.output, ba.Bytes...)
	// baPool.Put(ba)

	// // bia := biapool.Get().(*byteIndexAppender)
	// // bia.Index = 0
	// // fmt.Fprintf(bia, format, args...)
	// // vs.output = append(vs.output, bia.Bytes[:bia.Index]...)
	// // biapool.Put(bia)

	// // if vs.logger.newLine {
	// vs.output = append(vs.output, []byte("\"}\n")...)

	// vs.out.Write(vs.output)

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
