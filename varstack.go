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

func emptyvarStack(l *logger, level Level) *varStack {
	if l.Level() > level {
		return noopvarStack
	}
	vs := varStackPool.Get().(*varStack)
	vs.applyDefaults(l, level)
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
	// release references held from a previous pooled use so they can be
	// garbage collected
	vs.varName = [maxVars]string{}
	vs.varValue = [maxVars]any{}
}
