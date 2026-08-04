package iqlog

import (
	"io"
	"sync"
)

const maxVars = 8

type varStack struct {
	out           io.Writer
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
	if l.Level > level {
		return noopvarStack
	}
	vs := varStackPool.Get().(*varStack)
	vs.applyDefaults(l, level)
	// vs.level = level
	// vs.out = l.out
	// vs.jsonMode = l.jsonMode
	// vs.includeTime = l.IncludeTime
	// vs.captureCallers = l.captureCallers
	// vs.varCount = 0
	return vs
}

func (vs *varStack) applyDefaults(l *logger, level Level) {
	vs.level = level
	vs.out = l.out
	vs.jsonMode = l.jsonMode
	vs.includeTime = l.IncludeTime
	vs.captureCaller = l.CallerDepth
	vs.varCount = 0
}
