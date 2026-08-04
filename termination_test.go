package iqlog

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func expectPanic(t *testing.T, wantSubstr string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, wantSubstr) {
			t.Fatalf("expected panic containing %q, got %v", wantSubstr, r)
		}
	}()
	f()
}

func TestPanicBuildersPanicAfterWrite(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.SetWriter(sb)

	expectPanic(t, "boom-with", func() { l.WithByteSliceLinePanic().Str("k", "v").Msg("boom-with") })
	expectPanic(t, "boom-getwith", func() { l.Panic("boom-getwith") })
	expectPanic(t, "boom-standalone", func() { l.ByteSliceLinePanic("boom-standalone") })
	expectPanic(t, "boom-buffer", func() { l.BufferLinePanic("boom-buffer") })
	expectPanic(t, "boom-bufferwith", func() { l.WithBufferLinePanic().Str("k", "v").Msg("boom-bufferwith") })
	expectPanic(t, "boom-nl", func() { l.WithBufferLineNLPanic().Str("k", "v").Msg("boom-nl") })
	expectPanic(t, "boom-prealloc", func() { l.WithPreallocLinePanic().Str("k", "v").Msg("boom-prealloc") })
	expectPanic(t, "boom-prealloc2", func() { l.WithPreallocLine2Panic().Str("k", "v").Msg("boom-prealloc2") })
	expectPanic(t, "boom-varstack", func() { l.WithVarStackPanic().Str("k", "v").Msg("boom-varstack") })
	expectPanic(t, "boom-varstack-standalone", func() { l.VarStackPanic("boom-varstack-standalone") })

	out := sb.String()
	for _, want := range []string{"boom-with", "boom-getwith", "boom-standalone", "boom-buffer",
		"boom-bufferwith", "boom-nl", "boom-prealloc", "boom-prealloc2",
		"boom-varstack", "boom-varstack-standalone"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected record %q to be written before the panic; output: %q", want, out)
		}
	}
}

// TestFatalExitsProcess re-executes the test binary and asserts the global
// Fatal helper writes its record and terminates with exit code 1.
func TestFatalExitsProcess(t *testing.T) {
	if os.Getenv("IQLOG_TEST_FATAL") != "" {
		l := NewIQLogger(true)
		l.IncludeTime.Store(false)
		l.SetWriter(os.Stderr)
		switch os.Getenv("IQLOG_TEST_FATAL") {
		case "getwith":
			l.Fatal("fatal-record")
		case "varstack":
			l.VarStackFatal("fatal-record")
		case "withvarstack":
			l.WithVarStackFatal().Str("k", "v").Msg("fatal-record")
		}
		os.Exit(0) // should be unreachable
	}

	for _, mode := range []string{"getwith", "varstack", "withvarstack"} {
		cmd := exec.Command(os.Args[0], "-test.run", "TestFatalExitsProcess")
		cmd.Env = append(os.Environ(), "IQLOG_TEST_FATAL="+mode)
		out, err := cmd.CombinedOutput()
		exitErr, ok := err.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			t.Fatalf("mode %s: expected exit code 1, got err=%v output=%q", mode, err, out)
		}
		if !strings.Contains(string(out), "fatal-record") {
			t.Fatalf("mode %s: expected fatal record to be written, got %q", mode, out)
		}
	}
}
