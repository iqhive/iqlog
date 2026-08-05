package iqlog

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/iqhive/iqlog/ringbuffer"
)

func TestVarStackEmitsCallerFields(t *testing.T) {
	l := newTestLogger(true)
	l.SetCallerDepth(1)
	sb := &syncBuffer{}
	l.SetWriter(sb)

	l.WithVarStackInfo().Str("k", "v").Msg("m")
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

	sb2 := &syncBuffer{}
	l.SetWriter(sb2)
	l.VarStackInfo("standalone")
	out = sb2.String()
	if !strings.Contains(out, `"func":"`) || !strings.Contains(out, "hardening3_test") {
		t.Fatalf("expected caller fields in standalone varstack line, got %q", out)
	}
}

func TestRingBufferRejectsEnqueueAfterClose(t *testing.T) {
	rb := ringbuffer.NewRingBuffer[int](8)
	if !rb.Enqueue(1) {
		t.Fatal("expected enqueue to succeed before close")
	}
	rb.Close()
	if rb.Enqueue(2) {
		t.Fatal("expected enqueue to fail after close")
	}
	if v, ok := rb.Dequeue(); !ok || v != 1 {
		t.Fatalf("expected queued item to drain after close, got %v %v", v, ok)
	}
}

func TestAsyncWriterErrorRecorded(t *testing.T) {
	l := newTestLogger(true)
	l.SetAsyncWriter(failingWriter{})
	l.Info("fails in background")
	l.Flush()
	if err := l.LastWriteError(); err == nil || err.Error() != "disk full" {
		t.Fatalf("expected disk full error from async write, got %v", err)
	}
	l.SetWriter(&syncBuffer{})
}

func TestRingWriterErrorRecorded(t *testing.T) {
	l := newTestLogger(true)
	l.SetRingbufferWriter(failingWriter{})
	l.Info("fails in background")
	deadline := time.Now().Add(2 * time.Second)
	for l.LastWriteError() == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if err := l.LastWriteError(); err == nil || err.Error() != "disk full" {
		t.Fatalf("expected disk full error from ring write, got %v", err)
	}
	l.SetWriter(&syncBuffer{})
}

func TestLogWithFieldsUnknownLevel(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.SetWriter(sb)

	l.LogWithFields(context.Background(), Level(42), map[string]any{"k": "v"}, "odd level")
	out := sb.String()
	if !strings.Contains(out, "odd level") {
		t.Fatalf("expected record to be logged for unknown level, got %q", out)
	}
	if !json.Valid([]byte(strings.TrimSuffix(out, "\n"))) {
		t.Fatalf("invalid JSON: %q", out)
	}
}
