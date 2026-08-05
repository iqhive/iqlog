package iqlog

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEventEmitsCallerFields(t *testing.T) {
	l := newTestLogger(true)
	l.setCallerDepth(1)
	sb := &syncBuffer{}
	l.setWriter(sb)

	l.InfoEvent().Str("k", "v").Msg("m")
	out := sb.String()
	if !strings.Contains(out, `"func":"`) || !strings.Contains(out, `"file":"`) {
		t.Fatalf("expected caller fields in %q", out)
	}
	if !strings.Contains(out, "hardening3_test") {
		t.Fatalf("expected caller file to reference this test, got %q", out)
	}
	if !json.Valid([]byte(strings.TrimSuffix(out, "\n"))) {
		t.Fatalf("invalid JSON: %q", out)
	}

}

func TestAsyncWriterErrorRecorded(t *testing.T) {
	l := newTestLogger(true)
	l.setAsyncWriterLegacy(failingWriter{})
	l.Info("fails in background")
	l.Flush()
	if err := l.LastWriteError(); err == nil || err.Error() != "disk full" {
		t.Fatalf("expected disk full error from async write, got %v", err)
	}
	l.setWriter(&syncBuffer{})
}

func TestRingWriterErrorRecorded(t *testing.T) {
	l := newTestLogger(true)
	_ = l.setRingBufferWriterLegacy(failingWriter{})
	l.Info("fails in background")
	deadline := time.Now().Add(2 * time.Second)
	for l.LastWriteError() == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if err := l.LastWriteError(); err == nil || err.Error() != "disk full" {
		t.Fatalf("expected disk full error from ring write, got %v", err)
	}
	l.setWriter(&syncBuffer{})
}

func TestLogWithFieldsUnknownLevel(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.setWriter(sb)

	l.WithFields(map[string]any{"k": "v"}).LogContext(context.Background(), Level(42), "odd level")
	out := sb.String()
	if !strings.Contains(out, "odd level") {
		t.Fatalf("expected record to be logged for unknown level, got %q", out)
	}
	if !json.Valid([]byte(strings.TrimSuffix(out, "\n"))) {
		t.Fatalf("invalid JSON: %q", out)
	}
}
