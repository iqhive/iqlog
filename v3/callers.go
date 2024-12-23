package iqlog

import (
	"runtime"
	"strings"
)

// Maximum Number of paths to log
var NumPathsToLog = 1

// How many callers are we going back
var CallersNum = 2

var (
	FunctionsToSkip = []string{
		"/iqlog/",
		"runtime.",
		"testing.",
	}
)

// GetCallerFields returns information about the caller, specifically
// the function name, the filename and line number in the file
func GetCallerFields() map[string]any {
	resp := make(map[string]any)

	OriginFile := ""
	OriginLine := 0
	OriginFunc := ""

	if pc, file, line, ok := runtime.Caller(CallersNum); ok {
		OriginFile = file
		OriginLine = line
		runtimeFuncPtr := runtime.FuncForPC(pc)
		OriginFunc = runtimeFuncPtr.Name()
	}

	OriginFunc = KeepNumDirs(OriginFunc, NumPathsToLog)
	OriginFile = KeepNumDirs(OriginFile, NumPathsToLog)

	if OriginFile != "" {
		resp["origin_file"] = OriginFile
	}
	if OriginLine > 0 {
		resp["origin_line"] = OriginLine
	}
	if OriginFunc != "" {
		resp["origin_func"] = OriginFunc
	}
	return resp
}

func keepNumDirs(str string, lastn int, startat int) string {
	numFound := strings.Count(str[startat:], "/")
	if numFound > lastn {
		return keepNumDirs(str, lastn, 1+startat+strings.Index(str[startat:], "/"))
	}
	return str[startat:]
}

// KeepNumDirs returns the lastn number of directories in a path+filename combo
func KeepNumDirs(str string, lastn int) string {
	return keepNumDirs(str, lastn, 0)
}
