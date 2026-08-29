package iqlog_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/iqhive/iqlog"
)

var (
	_ func(bool) *iqlog.Logger                                                          = iqlog.NewIQLogger
	_ func() *iqlog.Logger                                                              = iqlog.NewGlobalIQLogger
	_ func(string, string, bool)                                                        = iqlog.Init
	_ func(string, ...any)                                                              = iqlog.Warning
	_ func(string, ...any)                                                              = iqlog.Warningf
	_ func(...any)                                                                      = iqlog.Warningln
	_ func() *iqlog.Event                                                               = iqlog.InfoWith
	_ func(io.Writer) error                                                             = iqlog.SetWriter
	_ func() io.Writer                                                                  = iqlog.GetWriter
	_ func(iqlog.Level)                                                                 = iqlog.SetLevel
	_ func(bool)                                                                        = iqlog.SetDebugMode
	_ func(int)                                                                         = iqlog.SetCallerDepth
	_ func(bool)                                                                        = iqlog.SetUseColour
	_ func(iqlog.JSONTimeMode)                                                           = iqlog.SetJSONTimeMode
	_ func(bool)                                                                        = iqlog.SetNewLine
	_ func(string)                                                                      = iqlog.SetApplicationName
	_ func(string)                                                                      = iqlog.SetSyslogHost
	_ func() error                                                                      = iqlog.Flush
	_ func(*iqlog.Logger, string, ...any)                                               = (*iqlog.Logger).Warning
	_ func(*iqlog.Logger, string, ...any)                                               = (*iqlog.Logger).Warningf
	_ func(*iqlog.Logger, ...any)                                                       = (*iqlog.Logger).Warningln
	_ func(*iqlog.Logger) *iqlog.Event                                                  = (*iqlog.Logger).TraceWith
	_ func(*iqlog.Logger) *iqlog.Event                                                  = (*iqlog.Logger).DebugWith
	_ func(*iqlog.Logger) *iqlog.Event                                                  = (*iqlog.Logger).InfoWith
	_ func(*iqlog.Logger) *iqlog.Event                                                  = (*iqlog.Logger).WarnWith
	_ func(*iqlog.Logger) *iqlog.Event                                                  = (*iqlog.Logger).ErrorWith
	_ func(*iqlog.Logger) *iqlog.Event                                                  = (*iqlog.Logger).PanicWith
	_ func(*iqlog.Logger) *iqlog.Event                                                  = (*iqlog.Logger).FatalWith
	_ func(*iqlog.Logger, io.Writer) error                                              = (*iqlog.Logger).SetWriter
	_ func(*iqlog.Logger, io.Writer) error                                              = (*iqlog.Logger).SetAsyncWriter
	_ func(*iqlog.Logger, io.Writer) error                                              = (*iqlog.Logger).SetRingbufferWriter
	_ func(*iqlog.Logger, io.Writer) error                                              = (*iqlog.Logger).SetRingBufferWriter
	_ func(*iqlog.Logger) io.Writer                                                     = (*iqlog.Logger).GetWriter
	_ func(*iqlog.Logger, iqlog.Level)                                                  = (*iqlog.Logger).SetLevel
	_ func(*iqlog.Logger, bool)                                                         = (*iqlog.Logger).SetDebugMode
	_ func(*iqlog.Logger, int)                                                          = (*iqlog.Logger).SetCallerDepth
	_ func(*iqlog.Logger, bool)                                                         = (*iqlog.Logger).SetUseColour
	_ func(*iqlog.Logger, bool)                                                         = (*iqlog.Logger).SetUseColor
	_ func(*iqlog.Logger, iqlog.JSONTimeMode)                                           = (*iqlog.Logger).SetJSONTimeMode
	_ func(*iqlog.Logger, bool)                                                         = (*iqlog.Logger).SetNewLine
	_ func(*iqlog.Logger, string)                                                       = (*iqlog.Logger).SetApplicationName
	_ func(*iqlog.Logger, string)                                                       = (*iqlog.Logger).SetSyslogHost
	_ func(*iqlog.Logger, context.Context, iqlog.Level, map[string]any, string, ...any) = (*iqlog.Logger).LogWithFields
	_ func(*iqlog.Logger, context.Context, iqlog.Level, map[string]any, string, ...any) = (*iqlog.Logger).LogFWithFields
)

func TestLegacyCompatibilityAdapters(t *testing.T) {
	old := iqlog.Default()
	t.Cleanup(func() { iqlog.SetDefault(old) })

	buf := &bytes.Buffer{}
	logger := iqlog.NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetLevel(iqlog.LevelTrace)
	logger.SetUseColour(false)
	logger.SetCallerDepth(0)
	iqlog.SetDefault(logger)

	iqlog.Warning("warning")
	iqlog.Warningf("warning=%d", 2)
	iqlog.Warningln("warning", 3)
	iqlog.InfoWith().Str("key", "value").Msg("event")
	logger.WarnWith().Msg("method event")
	logger.LogWithFields(context.Background(), iqlog.LevelInfo, map[string]any{"field": 4}, "fields")
	logger.LogFWithFields(context.Background(), iqlog.LevelInfo, nil, "formatted=%d", 5)
	iqlog.Flush()

	if buf.Len() == 0 {
		t.Fatal("legacy adapters produced no output")
	}
	if iqlog.GetWriter() != buf {
		t.Fatal("legacy writer adapter returned wrong writer")
	}
}

func TestLegacyConfigurationAdapters(t *testing.T) {
	logger := iqlog.NewGlobalIQLogger()
	logger.SetDebugMode(true)
	if logger.Level() != iqlog.LevelDebug {
		t.Fatal("debug mode did not set debug level")
	}
	logger.SetJSONTimeMode(iqlog.JSONTimeUTC)
	if logger.Config().Format != iqlog.FormatJSON || !logger.Config().IncludeTime {
		t.Fatal("JSON mode not applied")
	}
	logger.SetApplicationName("app")
	logger.SetNewLine(false)
	if iqlog.LevelPrint != iqlog.LevelInfo {
		t.Fatal("LevelPrint alias changed")
	}
}
