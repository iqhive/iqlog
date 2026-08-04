package iqlog

import (
	"sync"
)

const maxVars = 8

type varStack struct {
	logger        *logger
	jsonMode      bool
	includeTime   bool
	captureCaller int
	level         Level
	varName       [maxVars]string
	varValue      [maxVars]any
	varCount      int
	callerData    callerData
	// exitAfterWrite / panicAfterWrite terminate after the completed
	// record has been written, so Fatal/Panic lines are not lost
	exitAfterWrite  bool
	panicAfterWrite bool
	// varValueStr    [maxVars]string
}

var varStackPool = sync.Pool{
	New: func() any {
		resp := varStack{
			varName:  [maxVars]string{},
			varValue: [maxVars]any{},
		}
		return &resp
	},
}

var noopvarStack = &varStack{varCount: maxVars, level: LevelUnknown}

// emptyvarStack must not be inlined: the caller-capture skip count relies
// on it occupying its own stack frame between caller1 and the public entry
// point.
//
//go:noinline
func emptyvarStack(l *logger, level Level) *varStack {
	if l.Level() > level {
		return noopvarStack
	}
	vs := varStackPool.Get().(*varStack)
	vs.applyDefaults(l, level)
	if vs.captureCaller > 0 {
		var pc PC
		// skip counts from emptyvarStack's frame: +1 skips the public
		// entry point, leaving depth 1 = the entry point's caller (same
		// convention as the other builders)
		caller1(vs.captureCaller+1, &pc, 1, 1)
		fillCallerData(pc, &vs.callerData)
	}
	// vs.level = level
	// vs.out = l.out
	// vs.jsonMode = l.jsonMode.Load()
	// vs.includeTime = l.IncludeTime.Load()
	// vs.captureCallers = l.captureCallers
	// vs.varCount = 0
	return vs
}

func (vs *varStack) applyDefaults(l *logger, level Level) {
	vs.level = level
	vs.logger = l
	vs.jsonMode = l.jsonMode.Load()
	vs.includeTime = l.IncludeTime.Load()
	vs.captureCaller = int(l.CallerDepth.Load())
	vs.varCount = 0
	vs.exitAfterWrite = level == LevelFatal
	vs.panicAfterWrite = level == LevelPanic
	// release references held from a previous pooled use so they can be
	// garbage collected
	vs.varName = [maxVars]string{}
	vs.varValue = [maxVars]any{}
	vs.callerData.callerFuncLen = 0
}
