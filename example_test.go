package iqlog_test

import (
	"bytes"
	"context"
	"fmt"

	"github.com/iqhive/iqlog"
)

func Example() {
	var out bytes.Buffer
	logger := iqlog.MustNew(iqlog.Config{Writer: &out})
	logger.Info("started")
	fmt.Print(out.String())
	// Output: INFO started
}

func ExampleLogger_InfoEvent() {
	var out bytes.Buffer
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: &out})
	logger.InfoEvent().Str("user", "alice").Int("count", 2).Msg("handled")
	fmt.Print(out.String())
	// Output: {"level":"INFO","user":"alice","count":2,"message":"handled"}
}

func ExampleLogger_LogContext() {
	type key struct{}
	var out bytes.Buffer
	logger := iqlog.MustNew(iqlog.Config{Writer: &out, ContextExtractor: func(ctx context.Context) map[string]any {
		return map[string]any{"request": ctx.Value(key{})}
	}})
	logger.LogContext(context.WithValue(context.Background(), key{}, 42), iqlog.LevelInfo, "handled")
	fmt.Print(out.String())
	// Output: INFO request=42 handled
}
