//go:build iqlog_safe_callers

// Safe caller-capture path selected with -tags iqlog_safe_callers. It uses
// runtime.CallersFrames instead of decoding runtime function metadata
// directly, trading some speed for independence from runtime internals.

package iqlog

import (
	"runtime"
	"strings"
)

func fillCallerData(pc PC, callerData *callerData) {
	frames := runtime.CallersFrames([]uintptr{uintptr(pc)})
	f, _ := frames.Next()
	if f.Function == "" {
		return
	}

	file := f.File
	if i := strings.LastIndexByte(file, '/'); i >= 0 {
		file = file[i+1:]
	}
	n := copy(callerData.callerFile[:], file)
	if uint(n)+8 < callerDataMaxLen {
		callerData.callerFile[n] = ':'
		n++
		n += writeIntDecimal(callerData.callerFile[n:], int64(f.Line))
	}
	callerData.callerFileLen = uint(n)

	// Strip the main module path so callers show package paths relative to
	// the module or workspace root, matching the fast path.
	name := f.Function
	start := 0
	if mp := mainModulePath; mp != "" {
		common := 0
		for common < len(mp) && common < len(name) && name[common] == mp[common] {
			common++
		}
		if common > 0 {
			if common < len(name) && (name[common] == '/' || name[common] == '.') {
				start = common + 1
			} else if name[common-1] == '/' {
				start = common
			}
		}
	}
	callerData.callerFuncLen = uint(copy(callerData.callerFunc[:], name[start:]))
}
