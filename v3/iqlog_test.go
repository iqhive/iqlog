package iqlog_test

import (
	"testing"

	"bitbucket.org/iqhive/iqlog/v3"
)

func TestLog(t *testing.T) {
	iqlog.SetDebugMode(true)
	iqlog.Info("Hello World")
	iqlog.Warn("Hello World")
	iqlog.Error("Hello World")
	iqlog.Debug("Hello World")
}
