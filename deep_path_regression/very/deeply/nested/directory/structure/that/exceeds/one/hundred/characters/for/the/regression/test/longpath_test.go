package iqlog_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iqhive/iqlog"
)

// This file intentionally lives in a deeply nested directory so that its own
// source path (as recorded in the binary) exceeds the caller buffer limit
// with the last directory separator beyond the capped length.
//
// Regression test for a panic in the default caller-capture path
// (fillCallerData): slicing the caller's file path by an uncapped offset into
// a capped buffer produced "slice bounds out of range" whenever the last '/'
// in the path landed at index >= 101. Deep build/checkout paths (common in
// monorepos) crashed the whole application on every log line.

// TestLogFromDeepPathNoPanic logs through the public API from this deep
// file, which triggers caller capture with a long path end to end. The test
// binary used to crash with a slice-bounds panic here.
func TestLogFromDeepPathNoPanic(t *testing.T) {
	buf := &bytes.Buffer{}
	l := iqlog.MustNew(iqlog.Config{IncludeTime: true, Writer: buf, CallerDepth: 1})

	l.Info("deep path log") // must not panic

	out := buf.String()
	if !strings.Contains(out, "deep path log") {
		t.Fatalf("expected log output, got %q", out)
	}
	if !strings.Contains(out, "longpath_test.go") {
		t.Errorf("expected caller file basename longpath_test.go in output, got %q", out)
	}
}

// TestWithBuilderFromDeepPathNoPanic exercises the fluent builder from the
// deep path as well (it captures the caller itself rather than through the
// Info wrapper).
func TestWithBuilderFromDeepPathNoPanic(t *testing.T) {
	buf := &bytes.Buffer{}
	l := iqlog.MustNew(iqlog.Config{IncludeTime: true, Writer: buf, CallerDepth: 1})

	l.InfoEvent().Str("k", "v").Msg("deep path builder") // must not panic

	out := buf.String()
	if !strings.Contains(out, "deep path builder") {
		t.Fatalf("expected log output, got %q", out)
	}
	if !strings.Contains(out, "longpath_test.go") {
		t.Errorf("expected caller file basename longpath_test.go in output, got %q", out)
	}
}
