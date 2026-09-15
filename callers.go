package iqlog

import (
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

// mainModulePrefix is the main module path with a trailing slash. Function
// names inside the main module are reported without it. It is built once so
// the caller paths do not rebuild the string for every record.
var mainModulePrefix string

// familyPrefixes lists the import paths of this library and of its legacy
// home. A frame whose function lives in one of these packages, or in any
// package below them (such as a major-version suffix), is never reported as
// the caller. It is built once at package init and only read afterwards.
var familyPrefixes = [...]string{
	reflect.TypeOf(Logger{}).PkgPath(),
	"bitbucket.org/iqhive/iqlog",
}

func init() {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Path != "" {
		mainModulePrefix = info.Main.Path + "/"
	}
}

const callerDataMaxLen = 100

// maxCallerFrames bounds the stack walk performed by captureCallerScan. A
// caller deeper than this below the library boundary is reported as absent
// rather than resolved with a second walk.
const maxCallerFrames = 32

type callerData struct {
	callerFunc    [callerDataMaxLen]byte
	callerFuncLen uint
	callerFile    [callerDataMaxLen]byte
	callerFileLen uint
}

// isFamilyPackage reports whether the fully qualified function name belongs
// to prefix itself or to a package nested below it. The match stops at a
// package boundary so a sibling such as "iqlog-app" or "iqlogx" is not
// mistaken for the library.
func isFamilyPackage(name, prefix string) bool {
	if len(name) <= len(prefix) || name[:len(prefix)] != prefix {
		return false
	}
	switch name[len(prefix)] {
	case '.', '/':
		return true
	}
	return false
}

// isInternalFrame reports whether a frame belongs to this library's own
// non-test source. Frames from _test.go files are exempt so an in-package
// test helper is still attributed as the caller.
func isInternalFrame(name, file string) bool {
	if strings.HasSuffix(file, "_test.go") {
		return false
	}
	for _, prefix := range familyPrefixes {
		if isFamilyPackage(name, prefix) {
			return true
		}
	}
	return false
}

// isInfraFrame reports whether a frame belongs to the Go runtime or to the
// standard library logging front ends that forward to this library. The
// package path is compared exactly, so a user module whose path merely
// starts with "log." or "runtime." (such as "log.example.com/x") is still
// reported as the caller.
func isInfraFrame(name string) bool {
	dot := strings.LastIndexByte(name, '.')
	if dot < 0 {
		return false
	}
	pkg := name[:dot]
	// Method symbols include the receiver between the package and method.
	if receiver := strings.Index(pkg, ".("); receiver >= 0 {
		pkg = pkg[:receiver]
	}
	return pkg == "runtime" || pkg == "log" || pkg == "log/slog"
}

// captureCallerScan walks the stack outward from the library and records the
// callerDepth-th application frame: depth 1 is the first frame that is
// neither runtime, standard-library logging, nor this library's own source;
// depth 2 is its caller, and so on. When the stack runs out before that
// frame is reached nothing is recorded, so a wrong frame is never reported.
//
// Frames are resolved with runtime.CallersFrames rather than FuncForPC:
// FuncForPC heap-allocates a descriptor for every PC inside an inlined
// function, and the library's own wrappers between the entry point and this
// function are routinely inlined.
func captureCallerScan(callerDepth int, data *callerData) {
	if callerDepth <= 0 {
		return
	}
	var pcs [maxCallerFrames]uintptr
	// skip runtime.Callers itself and this function
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	remaining := callerDepth
	for {
		frame, more := frames.Next()
		if frame.PC != 0 && !isInfraFrame(frame.Function) && !isInternalFrame(frame.Function, frame.File) {
			remaining--
			if remaining == 0 {
				data.set(frame.Function, frame.File, frame.Line)
				return
			}
		}
		if !more {
			return
		}
	}
}

// captureCallerPC records the call site identified by pc, a return address
// as produced by runtime.Callers and carried by slog.Record.PC. The return
// address is moved back onto the call instruction before lookup, the same
// adjustment runtime.CallersFrames makes, so a call site inside an inlined
// function resolves to the inlined function rather than to whatever follows
// the call. Unlike CallersFrames this does not allocate for a call site that
// was not inlined. When the PC resolves to one of this library's own frames
// the record was produced from inside the library, and the caller is found
// by scanning the stack with captureCallerScan instead.
func captureCallerPC(pc uintptr, callerDepth int, data *callerData) {
	if pc == 0 {
		return
	}
	pc--
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return
	}
	name := fn.Name()
	file, line := fn.FileLine(pc)
	if isInfraFrame(name) || isInternalFrame(name, file) {
		captureCallerScan(callerDepth, data)
		return
	}
	data.set(name, file, line)
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
