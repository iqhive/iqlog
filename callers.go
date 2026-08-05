package iqlog

import (
	"runtime/debug"
	"strings"
)

var mainModulePath string

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		mainModulePath = info.Main.Path
	}
}

// Maximum Number of paths to log
var NumPathsToLog = 1

var (
	FunctionsToSkip = []string{
		"/iqlog/",
		"runtime.",
		"testing.",
	}
)

const callerDataMaxLen = 100

type callerData struct {
	callerFunc    [callerDataMaxLen]byte
	callerFuncLen uint
	callerFile    [callerDataMaxLen]byte
	callerFileLen uint
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
