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

func newTestLogger(jsonMode bool) *Logger {
	format := FormatConsole
	if jsonMode {
		format = FormatJSON
	}
	return MustNew(Config{Format: format})
}

func TestAsyncWriterConcurrentFlush(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.setAsyncWriterLegacy(sb)

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
		l.setAsyncWriterLegacy(sb)
		_ = l.setRingBufferWriterLegacy(sb)
	}
	l.setWriter(sb)

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
	_ = l.setRingBufferWriterLegacy(sb)
	for i := 0; i < 100; i++ {
		l.Info("ring drain")
	}
	l.setWriter(sb) // closes and drains the ring writer
	if got := strings.Count(sb.String(), "ring drain"); got != 100 {
		t.Fatalf("expected 100 drained lines, got %d", got)
	}
}

func TestLongEventJSONValid(t *testing.T) {
	long := strings.Repeat("x", 2000)
	longWithQuotes := strings.Repeat(`a"b\`, 500)

	for _, msg := range []string{long, longWithQuotes} {
		l := newTestLogger(true)
		sb := &syncBuffer{}
		l.setWriter(sb)
		l.InfoEvent().Str("key", msg).Msg(msg)
		out := strings.TrimSuffix(sb.String(), "\n")
		if out == "" {
			t.Fatal("no output")
		}
		if !json.Valid([]byte(out)) {
			t.Fatalf("line is not valid JSON: %q", out)
		}
	}
}

func TestConcurrentConfigAndLogging(t *testing.T) {
	l := newTestLogger(true)
	sb := &syncBuffer{}
	l.setWriter(sb)

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
					l.InfoEvent().Str("k", "v").Msg("builder")
				}
			}
		}()
	}
	for i := 0; i < 200; i++ {
		l.setLevel(LevelDebug)
		l.setLevel(LevelInfo)
		l.setFormat(FormatJSON)
		l.setCallerDepth(i % 3)
		l.setUseColor(i%2 == 0)
		l.setApplicationName("app")
	}
	close(stop)
	wg.Wait()
}

func TestConsoleNewlineSanitized(t *testing.T) {
	logFns := []func(l *Logger, s string){
		func(l *Logger, s string) { l.Info(s) },
		func(l *Logger, s string) { l.InfoEvent().Str("k", s).Msg("m") },
	}
	hostile := "line1\nFAKE forged log line\r\nline2"
	for i, log := range logFns {
		l := newTestLogger(false)
		sb := &syncBuffer{}
		l.setWriter(sb)
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

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("disk full")
}

func TestLastWriteError(t *testing.T) {
	l := newTestLogger(true)
	l.setWriter(failingWriter{})
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
	l.setWriter(sb)

	var wg sync.WaitGroup
	logFns := []func(){func() { l.InfoEvent().Str("builder", "event").Msg("m") }, func() { l.Info("plain") }}
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
	if len(lines) != 1000 {
		t.Fatalf("expected 1000 lines, got %d", len(lines))
	}
	for _, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatalf("interleaved/corrupt line: %q", line)
		}
	}
}
