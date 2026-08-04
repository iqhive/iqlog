//go:build !iqlog_safe_callers

// The default caller-capture path decodes runtime function metadata
// directly for speed. Build with -tags iqlog_safe_callers to use a
// slower implementation based on runtime.CallersFrames that does not
// depend on runtime internals.

package iqlog

import (
	"runtime"
	"strings"
	"unsafe"
)

func fillCallerData(pc PC, callerData *callerData) {
	f := findfunc(uintptr(pc))
	if f._func == nil {
		return
	}

	entry := funcInfoEntry(f)

	if pc > entry {
		pc--
	}

	var file string
	var line int

	file, line = (*runtime.Func)(unsafe.Pointer(f._func)).FileLine(uintptr(pc))
	callerData.callerFileLen = uint(len(file))
	if callerData.callerFileLen > callerDataMaxLen {
		callerData.callerFileLen = callerDataMaxLen
	}

	fileOffsetLast := 0
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			fileOffsetLast = i
			break
		}
	}
	// remove leading / before first directory
	if fileOffsetLast > 0 && fileOffsetLast < int(callerData.callerFileLen)-1 {
		fileOffsetLast++
	}

	copy(callerData.callerFile[:], file[fileOffsetLast:callerData.callerFileLen])
	callerData.callerFileLen -= uint(fileOffsetLast)
	if callerData.callerFileLen+8 < callerDataMaxLen {
		callerData.callerFile[callerData.callerFileLen] = ':'
		callerData.callerFileLen++
		i := writeIntDecimal(callerData.callerFile[callerData.callerFileLen:], int64(line))
		callerData.callerFileLen += uint(i)
	}

	// Get the function name
	zstr := &f.datap.funcnametab[f.nameOff]
	funcLen := findnull(zstr)

	// Strip the main module path so callers show package paths relative to the
	// module or workspace root (e.g. servers/datetime/pkg/datetime.loadCities
	// instead of github.com/iqhive/mcp/servers/datetime/pkg/datetime.loadCities).
	// We use the longest common prefix between the function name and the main
	// module path so nested Go workspaces/submodules still produce short paths.
	start := 0
	if prefixLen := len(mainModulePath); prefixLen > 0 && funcLen > 0 {
		nameBytes := unsafe.Slice((*byte)(unsafe.Pointer(zstr)), funcLen)

		common := 0
		maxCommon := prefixLen
		if funcLen < maxCommon {
			maxCommon = funcLen
		}
		for common < maxCommon && nameBytes[common] == mainModulePath[common] {
			common++
		}

		if common > 0 {
			if common < funcLen && (nameBytes[common] == '/' || nameBytes[common] == '.') {
				// Common prefix is a whole module/package component; strip the
				// separator too.
				start = common + 1
			} else if nameBytes[common-1] == '/' {
				// Common prefix ended right after a slash (diverging components
				// in the same parent workspace); strip up to that slash.
				start = common
			}
		}
	}

	copyLen := funcLen - start
	if copyLen > callerDataMaxLen {
		copyLen = callerDataMaxLen
	}

	callerData.callerFuncLen = uint(copyLen)
	zstrSlice := unsafe.Slice((*byte)(unsafe.Pointer(zstr)), funcLen)[start : start+copyLen]
	copy(callerData.callerFunc[:], zstrSlice)

	// callerData.callerFunc = [50]byte(name)
	// callerData.callerFile = [50]byte(file)
	// callerData.callerFileLen = uint(len(file))

	return
}

//go:nosplit
func findnull(s *byte) int {
	if s == nil {
		return 0
	}

	// pageSize is the unit we scan at a time looking for NULL.
	// It must be the minimum page size for any architecture Go
	// runs on. It's okay (just a minor performance loss) if the
	// actual system page size is larger than this value.
	const pageSize = 4096

	offset := 0
	ptr := unsafe.Pointer(s)
	// IndexByteString uses wide reads, so we need to be careful
	// with page boundaries. Call IndexByteString on
	// [ptr, endOfPage) interval.
	safeLen := int(pageSize - uintptr(ptr)%pageSize)

	for {
		t := *(*string)(unsafe.Pointer(&stringStruct{ptr, safeLen}))
		// Check one page at a time.
		if i := strings.IndexByte(t, 0); i != -1 {
			return offset + i
		}
		// Move to next page
		ptr = unsafe.Pointer(uintptr(ptr) + uintptr(safeLen))
		offset += safeLen
		safeLen = pageSize
	}
}

type stringStruct struct {
	str unsafe.Pointer
	len int
}
type funcInfo struct {
	*_func
	datap *moduledata
}

type srcFunc struct {
	datap     *moduledata
	nameOff   int32
	startLine int32
	funcID    uint8
}

type _func struct {
	entryOff uint32 // start pc, as offset from moduledata.text/pcHeader.textStart
	nameOff  int32  // function name, as index into moduledata.funcnametab.

	args        int32  // in/out args size
	deferreturn uint32 // offset of start of a deferreturn call instruction from entry, if any.

	pcsp      uint32
	pcfile    uint32
	pcln      uint32
	npcdata   uint32
	cuOffset  uint32 // runtime.cutab offset of this function's CU
	startLine int32  // line number of start of function (func keyword/TEXT directive)
	funcID    uint8  // set for certain special runtime functions
	flag      uint8
	_         [1]byte // pad
	nfuncdata uint8   // must be last, must end on a uint32-aligned boundary
}

type moduledata struct {
	pcHeader    unsafe.Pointer
	funcnametab []byte
	cutab       []uint32
	filetab     []byte
	pctab       []byte
	pclntable   []byte

	// omitted
}

//go:linkname findfunc runtime.findfunc
func findfunc(pc uintptr) funcInfo

//go:linkname funcInfoEntry runtime.funcInfo.entry
func funcInfoEntry(f funcInfo) PC
