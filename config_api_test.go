package iqlog

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

type nilBuffer struct{ bytes.Buffer }

func TestConfigValidationAndRoundTrip(t *testing.T) {
	l := MustNew(Config{})
	var typedNil *nilBuffer
	for name, cfg := range map[string]Config{
		"format": {Format: 99}, "depth": {CallerDepth: -1}, "size": {BufferSize: -1},
		"mode": {WriterMode: WriterMode(99)}, "policy": {OverflowPolicy: OverflowPolicy(99)}, "typed-nil": {Writer: typedNil},
	} {
		t.Run(name, func(t *testing.T) {
			if err := l.SetConfig(cfg); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	w := io.Discard
	want := Config{Format: FormatJSON, Level: LevelDebug, Writer: w, IncludeTime: true, TimestampLayout: "milliseconds", CallerDepth: 2, Color: true}
	if err := l.SetConfig(want); err != nil {
		t.Fatal(err)
	}
	got := l.Config()
	if got.Format != want.Format || got.Level != want.Level || got.Writer != w || !got.IncludeTime || got.CallerDepth != 2 || !got.Color {
		t.Fatalf("round trip: %#v", got)
	}
}

func TestNewAndMustNewErrors(t *testing.T) {
	if _, err := New(Config{Format: 99}); err == nil {
		t.Fatal("New accepted invalid config")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("MustNew did not panic")
		}
	}()
	MustNew(Config{Format: 99})
}

func TestSetConfigFailureIsTransactional(t *testing.T) {
	first := &syncBuffer{}
	l := MustNew(Config{Writer: first, Level: LevelInfo})
	before := l.Config()
	var typedNil *nilBuffer
	if err := l.SetConfig(Config{Writer: typedNil, Level: LevelDebug}); err == nil {
		t.Fatal("expected error")
	}
	after := l.Config()
	if after.Writer != before.Writer || after.Level != before.Level {
		t.Fatalf("configuration changed after failure: %#v", after)
	}
}

func TestFlushErrorsAreConsumedButLastErrorIsSticky(t *testing.T) {
	l := MustNew(Config{Writer: failingWriter{}, WriterMode: WriterAsync})
	l.Info("fails")
	if err := l.Flush(); err == nil {
		t.Fatal("expected flush error")
	}
	if err := l.Flush(); err != nil {
		t.Fatalf("second flush repeated old error: %v", err)
	}
	if err := l.LastWriteError(); err == nil {
		t.Fatal("last error should remain sticky")
	}
}

func TestCloseSharedAcrossDerivedLoggers(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	derived := l.WithFields(map[string]any{"key": "value"})
	if err := derived.Close(); err != nil {
		t.Fatal(err)
	}
	if l.Enabled(LevelInfo) || derived.Enabled(LevelInfo) {
		t.Fatal("closed logger remains enabled")
	}
	if err := l.Flush(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Flush=%v, want ErrClosed", err)
	}
	if err := l.SetConfig(Config{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("SetConfig=%v, want ErrClosed", err)
	}
}

func TestWriterModeRoundTrip(t *testing.T) {
	for _, mode := range []WriterMode{WriterSync, WriterAsync, WriterRing} {
		l := MustNew(Config{Writer: io.Discard, WriterMode: mode, OverflowPolicy: OverflowSync})
		if got := l.Config(); got.Writer != io.Discard || got.WriterMode != mode || got.OverflowPolicy != OverflowSync {
			t.Fatalf("round trip: %#v", got)
		}
		_ = l.Close()
	}
}

func TestLevelText(t *testing.T) {
	for _, want := range []Level{LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelPanic, LevelFatal} {
		got, err := ParseLevel(want.String())
		if err != nil || got != want {
			t.Fatalf("ParseLevel(%q)=%v,%v", want, got, err)
		}
		text, err := want.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		var decoded Level
		if err := decoded.UnmarshalText(text); err != nil || decoded != want {
			t.Fatalf("decoded=%v err=%v", decoded, err)
		}
	}
	if _, err := ParseLevel("invalid"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestConfiguredAsyncLifecycle(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Writer: sb, WriterMode: WriterAsync, BufferSize: 2, OverflowPolicy: OverflowBlock})
	l.Info("one")
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	if sb.String() != "INFO one\n" {
		t.Fatalf("output=%q", sb.String())
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestContextExtractorSkippedWhenDisabled(t *testing.T) {
	called := false
	l := MustNew(Config{Level: LevelError, Writer: io.Discard, ContextExtractor: func(context.Context) map[string]any { called = true; return nil }})
	l.LogContext(context.Background(), LevelDebug, "disabled")
	if called {
		t.Fatal("extractor called for disabled record")
	}
}
