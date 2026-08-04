package iqlog

import (
	"time"
)

var nooppreallocLine2 = &preallocLine2{}

func emptypreallocLine2(l *logger) *preallocLine2 {
	pal := preallocLine2Pool.Get().(*preallocLine2)
	pal.jsonMode = l.jsonMode.Load()
	pal.logger = l
	pal.bytesUsed = 0
	pal.includeTime = l.IncludeTime.Load()
	pal.captureCaller = int(l.CallerDepth.Load())
	pal.exitAfterWrite = false
	pal.callerData.callerFuncLen = 0
	// zero the output
	// pal.output = pal.output[:0]
	return pal
}

func (pal *preallocLine2) writeInitialJSON(level Level) {
	if pal.includeTime {
		pal.bytesUsed += AddTimeJSONInPlaceCopy(time.Now(), pal.output[pal.bytesUsed:])
	} else {
		pal.output[0] = '{'
		pal.bytesUsed++
	}

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
	if pal.captureCaller == 0 || pal.callerData.callerFuncLen == 0 {
		return
	}

	if pal.jsonMode {
		// json mode
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `,"func":"`)
		// pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, frame.Function)
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], pal.callerData.callerFunc[:pal.callerData.callerFuncLen])
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `","file":"`)
		// pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, frame.File[fileOffset2ndLast:])
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], pal.callerData.callerFile[:pal.callerData.callerFileLen])
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `"`)
	} else {
		// console mode
		if useColour.Load() {
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\x1b[32m[")
		} else {
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `[`)
		}
		// pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, frame.Function)
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], pal.callerData.callerFunc[:pal.callerData.callerFuncLen])
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ` `)
		// pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, frame.File[fileOffset2ndLast:])
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], pal.callerData.callerFile[:pal.callerData.callerFileLen])

		if useColour.Load() {
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "]\x1b[0m ")
		} else {
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, `] `)
		}
	}

}

func (pal *preallocLine2) writeInitialConsole(level Level) {
	// fmt.Printf("before time output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))
	if pal.includeTime {
		pal.bytesUsed += AddTimeConsoleInPlaceCopy(time.Now(), pal.output[pal.bytesUsed:])
	}
	// fmt.Printf("after time output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))
	if useColour.Load() {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, string(ansiColourPrefix(level)))
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, string(levelPrefix(level)))
	}
	// fmt.Printf("after writeInitialConsole: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))
	pal.AddCallers()
}

func (l *logger) WithPreallocLine2Trace() *preallocLine2 {
	if l.Level() > LevelTrace {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &pal.callerData)
	}
	if pal.jsonMode {
		pal.writeInitialJSON(LevelTrace)
	} else {
		pal.writeInitialConsole(LevelTrace)
	}
	return pal
}

func (l *logger) WithPreallocLine2Debug() *preallocLine2 {
	if l.Level() > LevelDebug {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &pal.callerData)
	}
	if pal.jsonMode {
		pal.writeInitialJSON(LevelDebug)
	} else {
		pal.writeInitialConsole(LevelDebug)
	}
	return pal
}

func (l *logger) WithPreallocLine2Info() *preallocLine2 {
	if l.Level() > LevelInfo {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &pal.callerData)
	}
	// fmt.Printf("Info1 output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))

	// l.getCaller()

	// pc := loc.Caller(1)
	// ok := pc != 0
	// name, file, line := pc.NameFileLine()
	// e := pc.FuncEntry()

	if pal.jsonMode {
		pal.writeInitialJSON(LevelInfo)
	} else {
		pal.writeInitialConsole(LevelInfo)
	}
	// fmt.Printf("Info2 output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))

	return pal
}

func (l *logger) WithPreallocLine2Warn() *preallocLine2 {
	if l.Level() > LevelWarn {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &pal.callerData)
	}
	if pal.jsonMode {
		pal.writeInitialJSON(LevelWarn)
	} else {
		pal.writeInitialConsole(LevelWarn)
	}
	return pal
}

func (l *logger) WithPreallocLine2Error() *preallocLine2 {
	if l.Level() > LevelError {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &pal.callerData)
	}
	if pal.jsonMode {
		pal.writeInitialJSON(LevelError)
	} else {
		pal.writeInitialConsole(LevelError)
	}
	return pal
}

func (l *logger) WithPreallocLine2Panic() *preallocLine2 {
	if l.Level() > LevelPanic {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &pal.callerData)
	}
	if pal.jsonMode {
		pal.writeInitialJSON(LevelPanic)
	} else {
		pal.writeInitialConsole(LevelPanic)
	}
	// exit after the final write so the record is not lost
	pal.exitAfterWrite = true
	return pal
}

func (l *logger) WithPreallocLine2Fatal() *preallocLine2 {
	if l.Level() > LevelFatal {
		return nooppreallocLine2
	}
	pal := emptypreallocLine2(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &pal.callerData)
	}
	if pal.jsonMode {
		pal.writeInitialJSON(LevelFatal)
	} else {
		pal.writeInitialConsole(LevelFatal)
	}
	// exit after the final write so the record is not lost
	pal.exitAfterWrite = true
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
