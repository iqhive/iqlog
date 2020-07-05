package iqlog

import (
	"bitbucket.org/iqhive/iqlog/v3/logruswrapper"
)

// GlobalLogger is used by the Global Logging functions
var GlobalLogger = NewGlobalIQLogger()

// NewIQLogger creates and return a new Logger
func NewGlobalIQLogger() *logruswrapper.LogrusWrapper {
	return logruswrapper.NewLogrusWrapper()
}
