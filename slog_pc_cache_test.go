package iqlog

import (
	"context"
	"log/slog"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Two call sites logging through the same handler on one goroutine reuse
// the same pooled event, so every record after the first is either a cache
// hit for the same PC or a miss that must displace the other site's text.

//go:noinline
func slogPCCacheSiteA(lg *slog.Logger) int {
	_, _, line, _ := runtime.Caller(0)
	lg.Info("a")
	return line + 1
}

//go:noinline
func slogPCCacheSiteB(lg *slog.Logger) int {
	_, _, line, _ := runtime.Caller(0)
	lg.Info("b")
	return line + 1
}

func slogPCCacheExpect(t *testing.T, buf *syncBuffer, fn string, line int, msg string) {
	t.Helper()
	rec := lastJSONMap(t, buf)
	if got, _ := rec["func"].(string); !strings.HasSuffix(got, "."+fn) {
		t.Fatalf("func %q want suffix %q (record %#v)", got, fn, rec)
	}
	if got, _ := rec["file"].(string); got != "slog_pc_cache_test.go:"+strconv.Itoa(line) {
		t.Fatalf("file %q want line %d (record %#v)", got, line, rec)
	}
	if rec["message"] != msg {
		t.Fatalf("message %v want %q", rec["message"], msg)
	}
	buf.Reset()
}

func TestSlogHandlerCallerCacheAlternatingSites(t *testing.T) {
	buf := &syncBuffer{}
	lg := slog.New(MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1}).SlogHandler())
	for i := 0; i < 4; i++ {
		line := slogPCCacheSiteA(lg)
		slogPCCacheExpect(t, buf, "slogPCCacheSiteA", line, "a")
		line = slogPCCacheSiteA(lg)
		slogPCCacheExpect(t, buf, "slogPCCacheSiteA", line, "a")
		line = slogPCCacheSiteB(lg)
		slogPCCacheExpect(t, buf, "slogPCCacheSiteB", line, "b")
	}
}

// The default JSON timestamp cache and the remembered PC resolution are both
// per-event state that outlives pooling. A logger without caller capture
// stamping a pooled event must leave the PC that event remembered either
// valid or cleared, so the next record with that PC never reports anything
// but its own call site; alternating the two loggers on one goroutine makes
// the same pooled event serve both.
func TestSlogHandlerCallerCacheSurvivesTimestampCache(t *testing.T) {
	fixed := time.Date(2026, 3, 4, 5, 6, 7, 8000, time.UTC)
	timed := &syncBuffer{}
	timedLog := MustNew(Config{Format: FormatJSON, Writer: timed, JSONTimeMode: JSONTimeUTC, Now: func() time.Time { return fixed }})
	sourced := &syncBuffer{}
	lg := slog.New(MustNew(Config{Format: FormatJSON, Writer: sourced, CallerDepth: 1}).SlogHandler())
	for i := 0; i < 4; i++ {
		line := slogPCCacheSiteA(lg)
		slogPCCacheExpect(t, sourced, "slogPCCacheSiteA", line, "a")
		timedLog.InfoEvent().Msg("tick")
		if got := timed.String(); !strings.HasPrefix(got, `{"time":"2026-03-04T05:06:07.000008Z","level":"INFO"`) {
			t.Fatalf("timestamp record %q", got)
		}
		timed.Reset()
	}
}

// An explicit caller from EventAt and a stack scan for a record without a PC
// both write the buffers through other paths; each must leave the next
// PC-carrying record resolving its own call site again.
func TestSlogHandlerCallerCacheDisplacedByOtherCallerPaths(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	lg := slog.New(l.SlogHandler())

	line := slogPCCacheSiteA(lg)
	slogPCCacheExpect(t, buf, "slogPCCacheSiteA", line, "a")

	l.EventAt(LevelInfo, "custom.Func", "custom.go:1").Msg("explicit")
	rec := lastJSONMap(t, buf)
	if rec["func"] != "custom.Func" || rec["file"] != "custom.go:1" {
		t.Fatalf("explicit caller: %#v", rec)
	}
	buf.Reset()

	line = slogPCCacheSiteA(lg)
	slogPCCacheExpect(t, buf, "slogPCCacheSiteA", line, "a")

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "scanned", 0)
	_, _, here, _ := runtime.Caller(0)
	if err := l.SlogHandler().Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	slogPCCacheExpect(t, buf, "TestSlogHandlerCallerCacheDisplacedByOtherCallerPaths", here+1, "scanned")

	line = slogPCCacheSiteB(lg)
	slogPCCacheExpect(t, buf, "slogPCCacheSiteB", line, "b")
}

// Console output reads the same buffers, so a hit must carry the lengths
// across the format as well.
func TestSlogHandlerCallerCacheConsole(t *testing.T) {
	buf := &syncBuffer{}
	lg := slog.New(MustNew(Config{Writer: buf, DisableColor: true, CallerDepth: 1}).SlogHandler())
	for i := 0; i < 3; i++ {
		line := slogPCCacheSiteA(lg)
		want := ".slogPCCacheSiteA slog_pc_cache_test.go:" + strconv.Itoa(line) + "] a\n"
		if got := buf.String(); !strings.HasPrefix(got, "INFO [") || !strings.HasSuffix(got, want) {
			t.Fatalf("console record %q want suffix %q", got, want)
		}
		buf.Reset()
	}
}

// Events are pooled across loggers, so loggers with different caller path
// modes, alternating on one goroutine, share the remembered PC resolution;
// each must still get its own form.
func TestSlogHandlerCallerCacheKeepsPathForm(t *testing.T) {
	pkg := reflect.TypeOf(Logger{}).PkgPath()
	type site struct {
		buf *syncBuffer
		lg  *slog.Logger
		fn  any // nil when no function is reported
	}
	var sites []site
	for mode, fn := range map[CallerPathMode]any{
		CallerPathRelative: pkg[strings.LastIndexByte(pkg, '/')+1:] + ".slogPCCacheSiteA",
		CallerPathLong:     pkg + ".slogPCCacheSiteA",
		CallerPathShort:    pkg[strings.LastIndexByte(pkg, '/')+1:] + ".slogPCCacheSiteA",
		CallerPathFile:     nil,
	} {
		buf := &syncBuffer{}
		lg := slog.New(MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1, CallerPathMode: mode}).SlogHandler())
		sites = append(sites, site{buf, lg, fn})
	}
	for i := 0; i < 3; i++ {
		for _, s := range sites {
			line := slogPCCacheSiteA(s.lg)
			rec := lastJSONMap(t, s.buf)
			if rec["func"] != s.fn {
				t.Fatalf("round %d: func %v, want %v", i, rec["func"], s.fn)
			}
			if got, want := rec["file"], "slog_pc_cache_test.go:"+strconv.Itoa(line); got != want {
				t.Fatalf("round %d: file %v, want %q", i, got, want)
			}
			s.buf.Reset()
		}
	}
}
