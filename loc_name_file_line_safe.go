//go:build iqlog_safe_callers

package iqlog

import "runtime"

func (l PC) nameFileLine() (name, file string, line int) {
	f := l.frameSafe()

	return f.Function, f.File, f.Line
}

func (l PC) FuncEntry() PC {
	f := l.frameSafe()

	return PC(f.Entry)
}

func (l PC) frameSafe() runtime.Frame {
	frames := runtime.CallersFrames([]uintptr{uintptr(l)})
	f, _ := frames.Next()

	return f
}
