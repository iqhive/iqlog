package iqlog

import (
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

var mainModulePath string

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		mainModulePath = info.Main.Path
	}
}

const callerDataMaxLen = 100

type callerData struct {
	callerFunc    [callerDataMaxLen]byte
	callerFuncLen uint
	callerFile    [callerDataMaxLen]byte
	callerFileLen uint
}

func captureCaller(skip int, data *callerData) {
	programCounter, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return
	}
	name := ""
	if fn := runtime.FuncForPC(programCounter); fn != nil {
		name = fn.Name()
	}
	name = strings.TrimPrefix(name, mainModulePath+"/")
	data.callerFuncLen = uint(copy(data.callerFunc[:], name))
	base := file
	if index := strings.LastIndexByte(file, '/'); index >= 0 {
		base = file[index+1:]
	}
	text := base + ":" + strconv.Itoa(line)
	data.callerFileLen = uint(copy(data.callerFile[:], text))
}
