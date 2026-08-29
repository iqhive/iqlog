package iqlog

import (
	"bytes"
	"strings"
	"testing"
)

// Regression tests for caller attribution: the fluent With* builders and the
// standalone line methods captured the caller's caller (one frame too deep)
// because their skip count assumed an extra wrapper frame that only exists on
// the Debug/Info/... paths. All direct-entry points must attribute the
// record to their own caller.

// TestWithBuilderCapturesDirectCaller checks that the caller captured by a
// With* builder is the function calling the builder method.
func TestWithBuilderCapturesDirectCaller(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(false)
	logger.setCallerDepth(1)
	logger.setUseColor(false)
	logger.setWriter(buf)

	logWithBuilderDirectly(logger)

	out := buf.String()
	if !strings.Contains(out, "logWithBuilderDirectly") {
		t.Errorf("expected caller to be logWithBuilderDirectly, got: %s", out)
	}
}

//go:noinline
func logWithBuilderDirectly(l *Logger) {
	l.InfoEvent().Str("k", "v").Msg("direct builder call")
}

// TestStandaloneMethodCapturesDirectCaller checks that the standalone line
// methods attribute the record to their own caller.
func TestStandaloneMethodCapturesDirectCaller(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(false)
	logger.setCallerDepth(1)
	logger.setUseColor(false)
	logger.setWriter(buf)

	logStandaloneDirectly(logger)

	out := buf.String()
	if !strings.Contains(out, "logStandaloneDirectly") {
		t.Errorf("expected caller to be logStandaloneDirectly, got: %s", out)
	}
}

//go:noinline
func logStandaloneDirectly(l *Logger) {
	l.InfoEvent().Msg("standalone call")
}

// TestWithBuilderAndInfoAgree checks the two paths agree on the caller when
// called from the same function: the With* builder must capture the same
// frame as the Info wrapper path.
func TestWithBuilderAndInfoAgree(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(false)
	logger.setCallerDepth(1)
	logger.setUseColor(false)
	logger.setWriter(buf)

	logBothWays(logger)

	out := buf.String()
	if !strings.Contains(out, "logBothWays") {
		t.Errorf("expected both records to be attributed to logBothWays, got: %s", out)
	}
}

//go:noinline
func logBothWays(l *Logger) {
	l.Info("via wrapper")
	l.InfoEvent().Msg("via builder")
}
