package iqlog

import (
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
)

// mainModulePrefix is the main module path with a trailing slash. Function
// names inside the main module are reported without it, so a symbol in a
// subpackage reads "sub/pkg.Func". Symbols in the module's root package have
// the module path followed by a dot instead; mainModuleRootPrefix matches
// those, and mainModuleRootTrim is how much to drop so they read "pkg.Func"
// with the package's own name kept. All three are built once so the caller
// paths do not rebuild strings for every record.
var (
	mainModulePrefix     string
	mainModuleRootPrefix string
	mainModuleRootTrim   int
)

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
		// a single-element module path ("app") is already just the package
		// name, so only multi-element paths need the root trimmed
		if i := strings.LastIndexByte(info.Main.Path, '/'); i >= 0 {
			mainModuleRootPrefix = info.Main.Path + "."
			mainModuleRootTrim = i + 1
		}
	}
}

// trimMainModule strips the main module path from a fully qualified function
// name, leaving the package-relative form the caller fields report.
func trimMainModule(name string) string {
	if mainModulePrefix != "" && strings.HasPrefix(name, mainModulePrefix) {
		return name[len(mainModulePrefix):]
	}
	if mainModuleRootPrefix != "" && strings.HasPrefix(name, mainModuleRootPrefix) {
		return name[mainModuleRootTrim:]
	}
	return name
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
	// pc is the record program counter whose resolution the two buffers
	// hold, with pcFuncLen and pcFileLen the lengths it was stored with, or
	// zero when they hold anything else. Events are pooled and a slog call
	// site presents the same PC on every record, so the pooled event's last
	// resolution is reused when the PC matches: the symbol-table lookups
	// cost more than the rest of the record. Every other write to the
	// buffers clears pc, so a match never reports stale text. Like the
	// Event's timestamp cache, this is per-event state and needs no
	// synchronisation.
	pc                   uintptr
	pcFuncLen, pcFileLen uint
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
	// Every infrastructure symbol starts with its package path and a dot.
	// Application frames, the common case, fail this prefix test at once and
	// skip the substring searches below, which measured as the larger cost of
	// resolving a record PC.
	if !strings.HasPrefix(name, "runtime.") && !strings.HasPrefix(name, "log.") && !strings.HasPrefix(name, "log/slog.") {
		return false
	}
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
	if data.pc == pc {
		// This pooled event last resolved the same call site and nothing
		// has written the buffers since; only the lengths were reset.
		data.callerFuncLen = data.pcFuncLen
		data.callerFileLen = data.pcFileLen
		return
	}
	lookup := pc - 1
	fn := runtime.FuncForPC(lookup)
	if fn == nil {
		return
	}
	name := fn.Name()
	file, line := fn.FileLine(lookup)
	if isInfraFrame(name) || isInternalFrame(name, file) {
		// The answer depends on the live stack, not on pc alone, so it is
		// found afresh every time and never remembered.
		captureCallerScan(callerDepth, data)
		return
	}
	data.set(name, file, line)
	data.pc = pc
	data.pcFuncLen = data.callerFuncLen
	data.pcFileLen = data.callerFileLen
}

// set stores the trimmed function name and "base:line" without allocating.
// Both are truncated to callerDataMaxLen. The line number is what makes the
// file reference useful, so an overlong basename is cut to make room for the
// complete ":line" suffix rather than the suffix being cut to a wrong number.
func (data *callerData) set(name, file string, line int) {
	// the buffers are about to hold something other than a remembered PC
	data.pc = 0
	name = trimMainModule(name)
	data.callerFuncLen = uint(copy(data.callerFunc[:], name))
	if index := strings.LastIndexByte(file, '/'); index >= 0 {
		file = file[index+1:]
	}
	// The line is formatted in place, least significant digit first, the
	// same text strconv would produce. A strconv call into scratch and a
	// second copy measured as a fifth of the cost of resolving a record PC.
	var digits [20]byte
	n := len(digits)
	u := uint64(line)
	if line < 0 {
		u = uint64(-line)
	}
	for {
		n--
		digits[n] = byte('0' + u%10)
		u /= 10
		if u == 0 {
			break
		}
	}
	if line < 0 {
		n--
		digits[n] = '-'
	}
	suffix := digits[n:]
	if limit := callerDataMaxLen - 1 - len(suffix); len(file) > limit {
		file = file[:limit]
	}
	n = copy(data.callerFile[:], file)
	data.callerFile[n] = ':'
	n++
	n += copy(data.callerFile[n:], suffix)
	data.callerFileLen = uint(n)
}
