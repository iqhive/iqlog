package iqlog

import (
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTestLogger(jsonMode bool) *logger {
	l := NewIQLogger(jsonMode)
	l.IncludeTime.Store(false)
	return l
}

func TestAsyncWriterConcurrentFlush(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.SetAsyncWriter(sb)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				l.Info("concurrent flush test")
			}
		}()
	}
	for i := 0; i < 20; i++ {
		l.Flush()
	}
	wg.Wait()
	l.Flush()

	if !strings.Contains(sb.String(), "concurrent flush test") {
		t.Fatal("expected log output after concurrent flushes")
	}
}

func TestWriterReplacementStopsGoroutines(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}

	before := runtime.NumGoroutine()
	for i := 0; i < 50; i++ {
		l.SetAsyncWriter(sb)
		l.SetRingbufferWriter(sb)
	}
	l.SetWriter(sb)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("goroutines leaked: before=%d after=%d", before, runtime.NumGoroutine())
}

func TestRingWriterDrainsOnClose(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.SetRingbufferWriter(sb)
	for i := 0; i < 100; i++ {
		l.Info("ring drain")
	}
	l.SetWriter(sb) // closes and drains the ring writer
	if got := strings.Count(sb.String(), "ring drain"); got != 100 {
		t.Fatalf("expected 100 drained lines, got %d", got)
	}
}

func TestPreallocTruncationKeepsJSONValid(t *testing.T) {
	long := strings.Repeat("x", 2000)
	longWithQuotes := strings.Repeat(`a"b\`, 500)

	cases := []func(l *logger, msg string){
		func(l *logger, msg string) { l.WithPreallocLineInfo().Str("key", msg).Msg(msg) },
		func(l *logger, msg string) { l.WithPreallocLine2Info().Str("key", msg).Msg(msg) },
		func(l *logger, msg string) { l.PreAllocLineInfo(msg) },
		func(l *logger, msg string) { l.PreAllocLine2Info(msg) },
	}
	for i, log := range cases {
		for _, msg := range []string{long, longWithQuotes} {
			l := newTestLogger(true)
			sb := &syncBuffer{}
			l.SetWriter(sb)
			log(l, msg)
			out := strings.TrimSuffix(sb.String(), "\n")
			if out == "" {
				t.Fatalf("case %d: no output", i)
			}
			if !json.Valid([]byte(out)) {
				t.Fatalf("case %d: truncated line is not valid JSON: %q", i, out)
			}
		}
	}
}

func TestConcurrentConfigAndLogging(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.SetWriter(sb)

	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					l.Info("config race test")
					l.WithByteSliceLineInfo().Str("k", "v").Msg("builder")
				}
			}
		}()
	}
	for i := 0; i < 200; i++ {
		l.SetLevel(LevelDebug)
		l.SetLevel(LevelInfo)
		l.SetJSONMode(true)
		l.SetCallerDepth(i % 3)
		l.SetNewLine(true)
		l.SetUseColour(i%2 == 0)
		l.SetApplicationName("app")
	}
	close(stop)
	wg.Wait()
}

func TestConsoleNewlineSanitized(t *testing.T) {
	logFns := []func(l *logger, s string){
		func(l *logger, s string) { l.Info(s) },
		func(l *logger, s string) { l.WithByteSliceLineInfo().Str("k", s).Msg("m") },
		func(l *logger, s string) { l.WithBufferLineInfo().Str("k", s).Msg("m") },
		func(l *logger, s string) { l.WithBufferLineNLInfo().Str("k", s).Msg("m") },
		func(l *logger, s string) { l.WithPreallocLineInfo().Str("k", s).Msg("m") },
		func(l *logger, s string) { l.WithPreallocLine2Info().Str("k", s).Msg("m") },
		func(l *logger, s string) { l.WithVarStackInfo().Str("k", s).Msg("m") },
	}
	hostile := "line1\nFAKE forged log line\r\nline2"
	for i, log := range logFns {
		l := newTestLogger(false)
		sb := &syncBuffer{}
		l.SetWriter(sb)
		log(l, hostile)
		out := sb.String()
		if got := strings.Count(out, "\n"); got != 1 {
			t.Fatalf("case %d: expected exactly 1 newline, got %d in %q", i, got, out)
		}
		if !strings.Contains(out, `\n`) {
			t.Fatalf("case %d: expected escaped newline in %q", i, out)
		}
	}
}

func TestPreallocBuildersEmitCallerFields(t *testing.T) {
	builders := []func(l *logger) string{
		func(l *logger) string {
			sb := &syncBuffer{}
			l.SetWriter(sb)
			l.WithPreallocLineInfo().Str("k", "v").Msg("m")
			return sb.String()
		},
		func(l *logger) string {
			sb := &syncBuffer{}
			l.SetWriter(sb)
			l.WithPreallocLine2Info().Str("k", "v").Msg("m")
			return sb.String()
		},
		func(l *logger) string {
			sb := &syncBuffer{}
			l.SetWriter(sb)
			l.WithBufferLineNLInfo().Str("k", "v").Msg("m")
			return sb.String()
		},
	}
	for i, run := range builders {
		l := newTestLogger(true)
		l.SetCallerDepth(1)
		out := run(l)
		if !strings.Contains(out, `"func":"`) || !strings.Contains(out, `"file":"`) {
			t.Fatalf("case %d: expected caller fields in %q", i, out)
		}
		if !json.Valid([]byte(strings.TrimSuffix(out, "\n"))) {
			t.Fatalf("case %d: invalid JSON: %q", i, out)
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("disk full")
}

func TestLastWriteError(t *testing.T) {
	l := newTestLogger(true)
	l.SetWriter(failingWriter{})
	if err := l.LastWriteError(); err != nil {
		t.Fatalf("expected nil before writes, got %v", err)
	}
	l.Info("this write fails")
	err := l.LastWriteError()
	if err == nil || err.Error() != "disk full" {
		t.Fatalf("expected disk full error, got %v", err)
	}
}

func TestMixedBuildersDoNotInterleave(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.SetWriter(sb)

	var wg sync.WaitGroup
	logFns := []func(){
		func() { l.WithByteSliceLineInfo().Str("builder", "byteslice").Msg("m") },
		func() { l.WithBufferLineInfo().Str("builder", "buffer").Msg("m") },
		func() { l.WithBufferLineNLInfo().Str("builder", "bufferNL").Msg("m") },
		func() { l.WithPreallocLineInfo().Str("builder", "prealloc").Msg("m") },
		func() { l.WithPreallocLine2Info().Str("builder", "prealloc2").Msg("m") },
		func() { l.WithVarStackInfo().Str("builder", "varstack").Msg("m") },
	}
	for _, fn := range logFns {
		wg.Add(1)
		go func(fn func()) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				fn()
			}
		}(fn)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSuffix(sb.String(), "\n"), "\n")
	if len(lines) != 3000 {
		t.Fatalf("expected 3000 lines, got %d", len(lines))
	}
	for _, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatalf("interleaved/corrupt line: %q", line)
		}
	}
}
