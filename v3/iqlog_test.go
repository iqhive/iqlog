package iqlog_test

import (
	"bytes"
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

func TestJSON(t *testing.T) {
	output := bytes.NewBuffer(nil)
	logger := iqlog.NewIQLogger(true)
	logger.SetWriter(output)
	logger.SetJSONMode(true)
	logger.SetDebugMode(true)
	logger.Info("Hello World")
	logger.Warn("Hello World")
	logger.Error("Hello World")
	logger.Debug("Hello World")

	gotStr := output.String()
	if gotStr != `{"level":"info","message":"Hello World"}
{"level":"warn","message":"Hello World"}
{"level":"error","message":"Hello World"}
{"level":"debug","message":"Hello World"}
` {
		t.Fatalf("gotStr = %s", gotStr)
	}
}
