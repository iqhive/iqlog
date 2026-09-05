package iqlog_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

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

func ExampleLogger_SlogHandler() {
	var out bytes.Buffer
	log := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: &out})
	logger := slog.New(log.SlogHandler()).WithGroup("http").With("method", "GET")
	logger.Info("handled", "status", 200)

	r := slog.NewRecord(time.Time{}, slog.LevelWarn, "hand built", 0)
	r.Add("user", "alice")
	_ = log.SlogHandler().Handle(context.Background(), r)
	fmt.Print(out.String())
	// Output:
	// {"level":"INFO","http.method":"GET","http.status":200,"message":"handled"}
	// {"level":"WARN","user":"alice","message":"hand built"}
}

func ExampleSlogHandler() {
	iqlog.SetDefault(iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: os.Stdout}))
	logger := slog.New(iqlog.SlogHandler()) // follows iqlog.SetDefault
	logger.Info("ready", "port", 8080)
	// Output: {"level":"INFO","port":8080,"message":"ready"}
}
