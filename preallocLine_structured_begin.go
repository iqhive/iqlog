package iqlog

import (
	"time"
)

var nooppreallocLine = &preallocLine{}

func emptypreallocLine(l *logger) *preallocLine {
	pal := preallocLinePool.Get().(*preallocLine)
	pal.bytesUsed = 0
	pal.jsonMode = l.jsonMode.Load()
	pal.logger = l
	pal.includeTime = l.IncludeTime.Load()
	pal.captureCaller = int(l.CallerDepth.Load())
	pal.exitAfterWrite = false
	pal.panicAfterWrite = false
	pal.callerData.callerFuncLen = 0

	return pal
}

func (pal *preallocLine) writeInitialJSON(level Level) {

	if pal.includeTime {
		pal.bytesUsed += AddTimeJSONInPlaceCopy(time.Now(), pal.output[pal.bytesUsed:])
	} else {
		pal.output[0] = '{'
		pal.bytesUsed++
	}

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
	pal.AddCallers()
}

func (pal *preallocLine) AddCallers() {
	if pal.captureCaller == 0 || pal.callerData.callerFuncLen == 0 {
		return
	}

	if pal.jsonMode {
		// json mode
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, `,"func":"`)
		// pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], pal.callerData.callerFunc[:pal.callerData.callerFuncLen])
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, `","file":"`)
		// pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, file[fileOffset2ndLast:])
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], pal.callerData.callerFile[:pal.callerData.callerFileLen])
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, `"`)
	} else {
		// console mode
		if useColour.Load() {
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\x1b[32m[")
		} else {
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, `[`)
		}
		// pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], pal.callerData.callerFunc[:pal.callerData.callerFuncLen])
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ` `)
		// pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, file[fileOffset2ndLast:])
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], pal.callerData.callerFile[:pal.callerData.callerFileLen])
		if useColour.Load() {
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "]\x1b[0m ")
		} else {
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, `] `)
		}
	}
	return

}

func (pal *preallocLine) writeInitialConsole(level Level) {
	if pal.includeTime {
		pal.bytesUsed += AddTimeConsoleInPlaceCopy(time.Now(), pal.output[pal.bytesUsed:])
	}
	if useColour.Load() {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, string(ansiColourPrefix(level)))
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, string(levelPrefix(level)))
	}
	pal.AddCallers()
}

func (l *logger) WithPreallocLineTrace() *preallocLine {
	if l.Level() > LevelTrace {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
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

func (l *logger) WithPreallocLineDebug() *preallocLine {
	if l.Level() > LevelDebug {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
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

func (l *logger) WithPreallocLineInfo() *preallocLine {
	if l.Level() > LevelInfo {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
	if d := int(l.CallerDepth.Load()); d > 0 {
		var pc PC
		caller1(d+1, &pc, 1, 1)
		fillCallerData(pc, &pal.callerData)
	}
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
	if l.Level() > LevelWarn {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
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

func (l *logger) WithPreallocLineError() *preallocLine {
	if l.Level() > LevelError {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
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

func (l *logger) WithPreallocLinePanic() *preallocLine {
	if l.Level() > LevelPanic {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
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
	pal.panicAfterWrite = true
	return pal
}

func (l *logger) WithPreallocLineFatal() *preallocLine {
	if l.Level() > LevelFatal {
		return nooppreallocLine
	}
	pal := emptypreallocLine(l)
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
