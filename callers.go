package iqlog

import (
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

// mainModulePrefix is the main module path with a trailing slash. Function
// names inside the main module are reported without it. It is built once so
// the caller paths do not rebuild the string for every record.
var mainModulePrefix string

func init() {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Path != "" {
		mainModulePrefix = info.Main.Path + "/"
	}
}

const callerDataMaxLen = 100

type callerData struct {
	callerFunc    [callerDataMaxLen]byte
	callerFuncLen uint
	callerFile    [callerDataMaxLen]byte
	callerFileLen uint
}

// captureCaller records the frame skip levels above the caller of
// captureCaller.
func captureCaller(skip int, data *callerData) {
	programCounter, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return
	}
	name := ""
	if fn := runtime.FuncForPC(programCounter); fn != nil {
		name = fn.Name()
	}
	data.set(name, file, line)
}

// captureCallerPC records the call site identified by pc, a return address
// as produced by runtime.Callers and carried by slog.Record.PC. The return
// address is moved back onto the call instruction before lookup, the same
// adjustment runtime.CallersFrames makes, so a call site inside an inlined
// function resolves to the inlined function rather than to whatever follows
// the call. Unlike CallersFrames this does not allocate for a call site that
// was not inlined.
func captureCallerPC(pc uintptr, data *callerData) {
	if pc == 0 {
		return
	}
	pc--
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return
	}
	file, line := fn.FileLine(pc)
	data.set(fn.Name(), file, line)
}

// set stores the trimmed function name and "base:line" without allocating.
// Both are truncated to callerDataMaxLen.
func (data *callerData) set(name, file string, line int) {
	name = strings.TrimPrefix(name, mainModulePrefix)
	data.callerFuncLen = uint(copy(data.callerFunc[:], name))
	if index := strings.LastIndexByte(file, '/'); index >= 0 {
		file = file[index+1:]
	}
	n := copy(data.callerFile[:], file)
	if n < callerDataMaxLen {
		data.callerFile[n] = ':'
		n++
		var digits [20]byte
		n += copy(data.callerFile[n:], strconv.AppendInt(digits[:0], int64(line), 10))
	}
	data.callerFileLen = uint(n)
}
