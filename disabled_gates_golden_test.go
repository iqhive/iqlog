package iqlog

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

// These tests pin the exact bytes and gate decisions of the paths exercised by
// the disabled-path benchmarks in bench/ (BenchmarkIQLogDisabled and
// BenchmarkIQLogSlogHandlerDisabled) and their enabled neighbours, so that a
// change to the level gates can be checked for byte-identical output and
// identical enable/disable decisions.

const gateGoldenMsg = "The quick brown fox jumps over the lazy dog"

const goldenBody = `"rate":"15","low":16,"high":123.2,"message":"The quick brown fox jumps over the lazy dog"}` + "\n"

var goldenNow = time.Date(2026, 3, 4, 5, 6, 7, 8000000, time.FixedZone("X", 2*3600))

func goldenPayload(e *Event) {
	e.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(gateGoldenMsg)
}

func goldenEnabled(t *testing.T, name string, l *Logger, levels []Level, want string) {
	t.Helper()
	got := make([]byte, 0, len(levels))
	for _, lv := range levels {
		if l.Enabled(lv) {
			got = append(got, 'T')
		} else {
			got = append(got, 'F')
		}
	}
	if string(got) != want {
		t.Fatalf("%s: Enabled over %v = %s, want %s", name, levels, got, want)
	}
}

func goldenSlogEnabled(t *testing.T, name string, h slog.Handler, levels []slog.Level, want string) {
	t.Helper()
	got := make([]byte, 0, len(levels))
	for _, lv := range levels {
		if h.Enabled(context.Background(), lv) {
			got = append(got, 'T')
		} else {
			got = append(got, 'F')
		}
	}
	if string(got) != want {
		t.Fatalf("%s: slog Enabled over %v = %s, want %s", name, levels, got, want)
	}
}

func TestGoldenTypedEventGates(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Level: LevelInfo, Writer: buf, JSONTimeMode: JSONTimeDisabled})
	levels := []Level{LevelUnknown, LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelPanic, LevelFatal, LevelFatal + 1, LevelFatal + 5, -3}

	// The benchmark payload on the disabled path: the gate returns nil and
	// the chained builder on a nil event must be a silent no-op.
	if e := l.DebugEvent(); e != nil {
		t.Fatalf("DebugEvent at Info returned %v, want nil", e)
	}
	if e := l.TraceEvent(); e != nil {
		t.Fatalf("TraceEvent at Info returned %v, want nil", e)
	}
	goldenPayload(l.DebugEvent())
	goldenPayload(l.TraceEvent())
	l.DebugEvent().Msg(gateGoldenMsg)
	l.DebugEvent().Discard()
	if got := buf.String(); got != "" {
		t.Fatalf("disabled events wrote %q", got)
	}
	goldenEnabled(t, "info", l, levels, "FFFTTTTTTTF")

	// The same payload through every enabled constructor.
	goldenPayload(l.InfoEvent())
	goldenPayload(l.WarnEvent())
	goldenPayload(l.ErrorEvent())
	goldenPayload(l.Event(LevelInfo))
	want := `{"level":"INFO",` + goldenBody + `{"level":"WARN",` + goldenBody + `{"level":"ERROR",` + goldenBody + `{"level":"INFO",` + goldenBody
	if got := buf.String(); got != want {
		t.Fatalf("enabled events wrote\n%q\nwant\n%q", got, want)
	}
	buf.Reset()

	// Raising the level flips Warn to nil without touching Error.
	if err := l.SetConfig(Config{Format: FormatJSON, Level: LevelError, Writer: buf, JSONTimeMode: JSONTimeDisabled}); err != nil {
		t.Fatal(err)
	}
	if e := l.WarnEvent(); e != nil {
		t.Fatal("WarnEvent at Error level returned an event")
	}
	goldenPayload(l.WarnEvent())
	goldenPayload(l.InfoEvent())
	goldenPayload(l.ErrorEvent())
	if got, want := buf.String(), `{"level":"ERROR",`+goldenBody; got != want {
		t.Fatalf("error-level events wrote\n%q\nwant\n%q", got, want)
	}
	buf.Reset()
	goldenEnabled(t, "error", l, levels, "FFFFFTTTTTF")

	// Lowering below Debug enables it again.
	if err := l.SetConfig(Config{Format: FormatJSON, Level: LevelTrace, Writer: buf, JSONTimeMode: JSONTimeDisabled}); err != nil {
		t.Fatal(err)
	}
	goldenPayload(l.TraceEvent())
	goldenPayload(l.DebugEvent())
	if got, want := buf.String(), `{"level":"TRACE",`+goldenBody+`{"level":"DEBUG",`+goldenBody; got != want {
		t.Fatalf("trace-level events wrote\n%q\nwant\n%q", got, want)
	}
	buf.Reset()
	goldenEnabled(t, "trace", l, levels, "FTTTTTTTTTF")

	// After Close every gate reports disabled, nothing is written, and
	// Level still reports the configured minimum.
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	goldenPayload(l.DebugEvent())
	goldenPayload(l.InfoEvent())
	goldenPayload(l.ErrorEvent())
	goldenEnabled(t, "closed", l, levels, "FFFFFFFFFFF")
	if got := l.Level(); got != LevelTrace {
		t.Fatalf("Level after close = %d, want %d", got, LevelTrace)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("closed logger wrote %q", got)
	}
}

func TestGoldenTypedEventTimestamps(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, JSONTimeMode: JSONTimeUTC, Now: func() time.Time { return goldenNow }})
	goldenPayload(l.InfoEvent())
	goldenPayload(l.DebugEvent())
	if got, want := buf.String(), `{"time":"2026-03-04T03:06:07.008000Z","level":"INFO",`+goldenBody; got != want {
		t.Fatalf("json utc wrote\n%q\nwant\n%q", got, want)
	}
	buf.Reset()

	c := MustNew(Config{Writer: buf, IncludeTime: true, DisableColor: true, Level: LevelDebug, Now: func() time.Time { return goldenNow }})
	goldenPayload(c.InfoEvent())
	goldenPayload(c.DebugEvent())
	goldenPayload(c.TraceEvent())
	want := "[2026-03-04T05:06:07.008000] INFO rate=15 low=16 high=123.2 " + gateGoldenMsg + "\n" +
		"[2026-03-04T05:06:07.008000] DBUG rate=15 low=16 high=123.2 " + gateGoldenMsg + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("console wrote\n%q\nwant\n%q", got, want)
	}
}

func TestGoldenSlogHandlerGates(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Level: LevelInfo, Writer: buf, JSONTimeMode: JSONTimeUTC})
	h := l.SlogHandler()
	ctx := context.Background()
	levels := []slog.Level{-100, -8, -5, -4, -3, -1, 0, 1, 3, 4, 5, 7, 8, 9, 12, 100}
	goldenSlogEnabled(t, "info", h, levels, "FFFFFFTTTTTTTTTT")

	logger := slog.New(h)
	// The benchmark payload on the disabled path.
	logger.LogAttrs(ctx, slog.LevelDebug, gateGoldenMsg, slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
	logger.Debug(gateGoldenMsg, "rate", "15")
	if got := buf.String(); got != "" {
		t.Fatalf("disabled slog records wrote %q", got)
	}
	// The same payload enabled, with a fixed record time.
	for _, lv := range []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError, 12} {
		r := slog.NewRecord(goldenNow, lv, gateGoldenMsg, 0)
		r.AddAttrs(slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
		if err := h.Handle(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	const stamp = `{"time":"2026-03-04T03:06:07.008000Z","level":"`
	want := stamp + `INFO",` + goldenBody + stamp + `WARN",` + goldenBody + stamp + `ERROR",` + goldenBody + stamp + `ERROR",` + goldenBody
	if got := buf.String(); got != want {
		t.Fatalf("enabled slog records wrote\n%q\nwant\n%q", got, want)
	}
	buf.Reset()

	if err := l.SetConfig(Config{Format: FormatJSON, Level: LevelWarn, Writer: buf, JSONTimeMode: JSONTimeUTC}); err != nil {
		t.Fatal(err)
	}
	goldenSlogEnabled(t, "warn", h, levels, "FFFFFFFFFTTTTTTT")
	r := slog.NewRecord(goldenNow, slog.LevelInfo, gateGoldenMsg, 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("Info record at Warn wrote %q", got)
	}

	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	goldenSlogEnabled(t, "closed", h, levels, "FFFFFFFFFFFFFFFF")
	r = slog.NewRecord(goldenNow, slog.LevelError, gateGoldenMsg, 0)
	if err := h.Handle(ctx, r); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("record after close wrote %q", got)
	}
}

// The package-level handler answers for whatever Default returns at call
// time, including before anything has built the default logger, when it
// must agree with the logger Default then builds.
func TestGoldenPackageSlogHandlerGates(t *testing.T) {
	old := Default()
	t.Cleanup(func() { SetDefault(old) })
	levels := []slog.Level{-100, -8, -5, -4, -3, -1, 0, 1, 3, 4, 5, 7, 8, 9, 12, 100}

	defaultLogger.Store(nil)
	h := SlogHandler()
	goldenSlogEnabled(t, "unbuilt default", h, levels, "FFFFFFTTTTTTTTTT")
	built := Default()
	for _, lv := range levels {
		if got, want := h.Enabled(context.Background(), lv), built.Enabled(slogLevelToIQ(lv)); got != want {
			t.Fatalf("unbound Enabled(%d) = %v before the default was built, but the built default says %v", lv, got, want)
		}
	}
	goldenSlogEnabled(t, "built default", h, levels, "FFFFFFTTTTTTTTTT")

	buf := &syncBuffer{}
	custom := MustNew(Config{Format: FormatJSON, Level: LevelWarn, Writer: buf, JSONTimeMode: JSONTimeDisabled})
	SetDefault(custom)
	goldenSlogEnabled(t, "custom warn default", h, levels, "FFFFFFFFFTTTTTTT")
	logger := slog.New(h)
	logger.LogAttrs(context.Background(), slog.LevelInfo, gateGoldenMsg, slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
	logger.LogAttrs(context.Background(), slog.LevelWarn, gateGoldenMsg, slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
	if got, want := buf.String(), `{"level":"WARN",`+goldenBody; got != want {
		t.Fatalf("package handler wrote\n%q\nwant\n%q", got, want)
	}
	if err := custom.Close(); err != nil {
		t.Fatal(err)
	}
	goldenSlogEnabled(t, "closed default", h, levels, "FFFFFFFFFFFFFFFF")
}
