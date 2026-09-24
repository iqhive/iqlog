package iqlog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// slogGoldenMsg is the message the benchmarks in bench/ log.
const slogGoldenMsg = "The quick brown fox jumps over the lazy dog"

var slogGoldenFixed = time.Date(2026, 3, 4, 5, 6, 7, 8000, time.UTC)

// slogGoldenSource logs from a named, non-inlined function so the caller
// fields have a stable function name and a line this file controls.
//
//go:noinline
func slogGoldenSource(lg *slog.Logger) int {
	_, _, line, _ := runtime.Caller(0)
	lg.LogAttrs(context.Background(), slog.LevelInfo, slogGoldenMsg, slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
	return line + 1
}

type slogGoldenCase struct {
	name string
	cfg  Config
	// with wraps the handler like the benchmark's WithGroup/With chain.
	with func(slog.Handler) slog.Handler
	// log emits through the slog.Logger, or handle feeds a hand-built record
	// straight to the handler when the record time must be fixed.
	log    func(*slog.Logger)
	handle func(slog.Handler) error
	want   string
}

type slogGoldenErr struct{}

func (*slogGoldenErr) Error() string { return "typed nil" }

func slogGoldenCases() []slogGoldenCase {
	ctx := context.Background()
	boom := errors.New("boom")
	simple := func(lg *slog.Logger) {
		lg.LogAttrs(ctx, slog.LevelInfo, slogGoldenMsg, slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
	}
	grouped := func(h slog.Handler) slog.Handler {
		return slog.New(h).WithGroup("http").With("method", "GET", "path", "/orders").Handler()
	}
	groupedLog := func(lg *slog.Logger) {
		lg.LogAttrs(ctx, slog.LevelInfo, slogGoldenMsg, slog.Int("status", 200), slog.String("client", "10.0.0.1"))
	}
	timed := func(h slog.Handler) error {
		r := slog.NewRecord(slogGoldenFixed, slog.LevelInfo, slogGoldenMsg, 0)
		r.AddAttrs(slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
		return h.Handle(ctx, r)
	}
	// slog.Group boxes its arguments; built once so the cases stay simple.
	peer := slog.Group("peer", slog.String("ip", "10.0.0.1"), slog.Int("port", 443), slog.Group("tls", slog.Bool("on", true)))
	return []slogGoldenCase{
		{name: "bench simple json", cfg: Config{Format: FormatJSON}, log: simple,
			want: `{"level":"INFO","rate":"15","low":16,"high":123.2,"message":"The quick brown fox jumps over the lazy dog"}` + "\n"},
		{name: "bench grouped json", cfg: Config{Format: FormatJSON}, with: grouped, log: groupedLog,
			want: `{"level":"INFO","http.method":"GET","http.path":"/orders","http.status":200,"http.client":"10.0.0.1","message":"The quick brown fox jumps over the lazy dog"}` + "\n"},
		{name: "bench utc timestamp json", cfg: Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC}, handle: timed,
			want: `{"time":"2026-03-04T05:06:07.000008Z","level":"INFO","rate":"15","low":16,"high":123.2,"message":"The quick brown fox jumps over the lazy dog"}` + "\n"},
		{name: "bench console", cfg: Config{DisableColor: true}, log: simple,
			want: "INFO rate=15 low=16 high=123.2 The quick brown fox jumps over the lazy dog\n"},
		{name: "bench console grouped", cfg: Config{DisableColor: true}, with: grouped, log: groupedLog,
			want: "INFO http.method=GET http.path=/orders http.status=200 http.client=10.0.0.1 The quick brown fox jumps over the lazy dog\n"},
		{name: "console with record time", cfg: Config{DisableColor: true, IncludeTime: true}, handle: timed,
			want: "[2026-03-04T05:06:07.000008] INFO rate=15 low=16 high=123.2 The quick brown fox jumps over the lazy dog\n"},
		{name: "reserved and near-reserved keys", cfg: Config{Format: FormatJSON}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m",
				slog.String("time", "t"), slog.String("level", "l"), slog.String("message", "m"), slog.String("func", "f"), slog.String("file", "x"),
				slog.String("timer", "ok"), slog.String("field_time", "ft"), slog.String("tim", "short"), slog.Int("levels", 1), slog.String("messages", "s"),
				slog.String("Time", "T"), slog.String("fun", "n"), slog.String("files", "z"), slog.String("", "empty"), slog.String("t", "1"))
		}, want: `{"level":"INFO","field_time":"t","field_level":"l","field_message":"m","field_func":"f","field_file":"x","timer":"ok","field_time":"ft","tim":"short","levels":1,"messages":"s","Time":"T","fun":"n","files":"z","":"empty","t":"1","message":"m"}` + "\n"},
		{name: "reserved keys inside groups stay dotted", cfg: Config{Format: FormatJSON}, with: func(h slog.Handler) slog.Handler {
			return h.WithGroup("g").WithAttrs([]slog.Attr{slog.String("level", "pre")})
		}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.String("time", "t"), slog.Group("message", slog.String("func", "f")))
		}, want: `{"level":"INFO","g.level":"pre","g.time":"t","g.message.func":"f","message":"m"}` + "\n"},
		{name: "keys needing escaping", cfg: Config{Format: FormatJSON}, with: func(h slog.Handler) slog.Handler {
			return h.WithGroup("a\"b").WithAttrs([]slog.Attr{slog.String("k\\", "v")}).WithGroup("c\nd")
		}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.String("we\"ird", "v"), slog.String("tab\tkey", "v"), slog.String("ünï", "v"), slog.String("bad\xffkey", "v"), slog.Group("in\"g", slog.Int("n", 1)))
		}, want: "{\"level\":\"INFO\",\"a\\\"b.k\\\\\":\"v\",\"a\\\"b.c\\nd.we\\\"ird\":\"v\",\"a\\\"b.c\\nd.tab\\tkey\":\"v\",\"a\\\"b.c\\nd.ünï\":\"v\",\"a\\\"b.c\\nd.bad\\ufffdkey\":\"v\",\"a\\\"b.c\\nd.in\\\"g.n\":1,\"message\":\"m\"}\n"},
		{name: "top-level escaped key", cfg: Config{Format: FormatJSON}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.String("we\"ird", "v"), slog.String("tab\tkey", "v"), slog.String("ünï", "v"), slog.String("bad\xffkey", "v"), slog.String("\x01", "v"))
		}, want: "{\"level\":\"INFO\",\"we\\\"ird\":\"v\",\"tab\\tkey\":\"v\",\"ünï\":\"v\",\"bad\\ufffdkey\":\"v\",\"\\u0001\":\"v\",\"message\":\"m\"}\n"},
		{name: "escape field names option", cfg: Config{Format: FormatJSON, EscapeFieldNames: true}, with: grouped, log: groupedLog,
			want: `{"level":"INFO","http.method":"GET","http.path":"/orders","http.status":200,"http.client":"10.0.0.1","message":"The quick brown fox jumps over the lazy dog"}` + "\n"},
		{name: "record groups", cfg: Config{Format: FormatJSON}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", peer, slog.Group("", slog.String("inline", "yes")), slog.Group("empty"), slog.Group("", slog.Group("", slog.Int("deep", 1))))
		}, want: `{"level":"INFO","peer.ip":"10.0.0.1","peer.port":443,"peer.tls.on":true,"inline":"yes","deep":1,"message":"m"}` + "\n"},
		{name: "empty key and value dropped", cfg: Config{Format: FormatJSON}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.Attr{}, slog.String("", ""), slog.Int("", 0), slog.Any("", nil), slog.Group("", slog.Attr{}))
		}, want: `{"level":"INFO","":"","":0,"message":"m"}` + "\n"},
		{name: "log valuers", cfg: Config{Format: FormatJSON}, with: func(h slog.Handler) slog.Handler {
			return h.WithAttrs([]slog.Attr{slog.Any("pre", slogNameValuer("resolved-pre"))})
		}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.Any("name", slogNameValuer("resolved")), slog.Any("gv", slogGroupValuer{}), slog.Group("g", slog.Any("inner", slogNameValuer("x"))))
		}, want: `{"level":"INFO","pre":"resolved-pre","name":"resolved","gv.a":"1","gv.b":2,"g.inner":"x","message":"m"}` + "\n"},
		{name: "all kinds", cfg: Config{Format: FormatJSON}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "kinds",
				slog.String("str", "va\"l\nue\x00\xff"),
				slog.Int64("i64", -2), slog.Int("min", math.MinInt64), slog.Uint64("u64", math.MaxUint64),
				slog.Float64("f64", 2.5), slog.Float64("big", 1e21), slog.Float64("small", 1e-7), slog.Float64("neg0", math.Copysign(0, -1)), slog.Float64("nan", math.NaN()), slog.Float64("inf", math.Inf(-1)),
				slog.Bool("ok", true), slog.Bool("no", false),
				slog.Duration("duration", time.Second+2*time.Millisecond), slog.Duration("zero", 0),
				slog.Time("when", slogGoldenFixed), slog.Time("local", slogGoldenFixed.In(time.FixedZone("X", 3600))), slog.Time("zero", time.Time{}),
				slog.Any("err", boom), slog.Any("nilerr", (*slogGoldenErr)(nil)), slog.Any("empty", nil),
				slog.Any("bytes", []byte("hi\x00")), slog.Any("raw", json.RawMessage(`{"nested":true}`)),
				slog.Any("struct", struct{ A int }{7}), slog.Any("map", map[string]int{"z": 1, "a": 2}), slog.Any("i32", int32(5)), slog.Any("u8", uint8(9)), slog.Any("f32", float32(1.5)), slog.Any("ptr", &boom),
			)
		}, want: "{\"level\":\"INFO\",\"str\":\"va\\\"l\\nue\\u0000\\ufffd\",\"i64\":-2,\"min\":-9223372036854775808,\"u64\":18446744073709551615,\"f64\":2.5,\"big\":1000000000000000000000,\"small\":0.0000001,\"neg0\":-0,\"nan\":\"NaN\",\"inf\":\"-Infinity\",\"ok\":true,\"no\":false,\"duration\":\"1.002s\",\"zero\":\"0s\",\"when\":\"2026-03-04T05:06:07.000008Z\",\"local\":\"2026-03-04T06:06:07.000008+01:00\",\"zero\":\"0001-01-01T00:00:00Z\",\"err\":\"boom\",\"nilerr\":null,\"empty\":null,\"bytes\":\"aGkA\",\"raw\":{\"nested\":true},\"struct\":{\"A\":7},\"map\":{\"a\":2,\"z\":1},\"i32\":5,\"u8\":9,\"f32\":1.5,\"ptr\":{},\"message\":\"kinds\"}\n"},
		{name: "message escaping", cfg: Config{Format: FormatJSON}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "quote \" slash \\ nl \n tab \t ctl \x01 uni ünï bad \xff")
		}, want: "{\"level\":\"INFO\",\"message\":\"quote \\\" slash \\\\ nl \\n tab \\t ctl \\u0001 uni ünï bad \\ufffd\"}\n"},
		{name: "empty message and no attrs", cfg: Config{Format: FormatJSON}, log: func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "")
		}, want: `{"level":"INFO","message":""}` + "\n"},
		{name: "level mapping", cfg: Config{Format: FormatJSON, Level: LevelTrace}, log: func(lg *slog.Logger) {
			for _, lv := range []slog.Level{slog.LevelDebug - 4, slog.LevelDebug - 1, slog.LevelDebug, slog.LevelDebug + 3, slog.LevelInfo, slog.LevelWarn - 1, slog.LevelWarn, slog.LevelError - 1, slog.LevelError, slog.LevelError + 8} {
				lg.LogAttrs(ctx, lv, "m", slog.Int("v", int(lv)))
			}
		}, want: `{"level":"TRACE","v":-8,"message":"m"}` + "\n" + `{"level":"TRACE","v":-5,"message":"m"}` + "\n" + `{"level":"DEBUG","v":-4,"message":"m"}` + "\n" + `{"level":"DEBUG","v":-1,"message":"m"}` + "\n" + `{"level":"INFO","v":0,"message":"m"}` + "\n" + `{"level":"INFO","v":3,"message":"m"}` + "\n" + `{"level":"WARN","v":4,"message":"m"}` + "\n" + `{"level":"WARN","v":7,"message":"m"}` + "\n" + `{"level":"ERROR","v":8,"message":"m"}` + "\n" + `{"level":"ERROR","v":16,"message":"m"}` + "\n"},
		{name: "level gate", cfg: Config{Format: FormatJSON, Level: LevelWarn}, log: func(lg *slog.Logger) {
			for _, lv := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn - 1, slog.LevelWarn, slog.LevelError} {
				lg.LogAttrs(ctx, lv, "m", slog.Int("v", int(lv)))
			}
		}, want: `{"level":"WARN","v":4,"message":"m"}` + "\n" + `{"level":"ERROR","v":8,"message":"m"}` + "\n"},
		{name: "console level mapping and escaping", cfg: Config{DisableColor: true, Level: LevelTrace}, log: func(lg *slog.Logger) {
			for _, lv := range []slog.Level{slog.LevelDebug - 4, slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
				lg.LogAttrs(ctx, lv, "m\x1b[31m\n", slog.String("k\x00", "v\x1b"), slog.Bool("b", true), slog.Time("t", slogGoldenFixed), slog.Duration("d", time.Second), slog.Any("e", boom), slog.Any("n", nil), slog.Any("b", []byte("hi")), slog.Uint64("u", 3), slog.Float64("f", 1.5), slog.Group("g", slog.Int("i", 1)))
			}
		}, want: "DBUG k\\x00=v\\x1b b=true t=2026-03-04T05:06:07.000008Z d=1s e=boom n=null b=aGk= u=3 f=1.5 g.i=1 m\\x1b[31m\\n\nDBUG k\\x00=v\\x1b b=true t=2026-03-04T05:06:07.000008Z d=1s e=boom n=null b=aGk= u=3 f=1.5 g.i=1 m\\x1b[31m\\n\nINFO k\\x00=v\\x1b b=true t=2026-03-04T05:06:07.000008Z d=1s e=boom n=null b=aGk= u=3 f=1.5 g.i=1 m\\x1b[31m\\n\nWARN k\\x00=v\\x1b b=true t=2026-03-04T05:06:07.000008Z d=1s e=boom n=null b=aGk= u=3 f=1.5 g.i=1 m\\x1b[31m\\n\nERRR k\\x00=v\\x1b b=true t=2026-03-04T05:06:07.000008Z d=1s e=boom n=null b=aGk= u=3 f=1.5 g.i=1 m\\x1b[31m\\n\n"},
		{name: "console color", cfg: Config{Color: true}, with: grouped, log: groupedLog,
			want: "\x1b[34mINFO\x1b[0m http.method=GET http.path=/orders http.status=200 http.client=10.0.0.1 The quick brown fox jumps over the lazy dog\n"},
		{name: "context extractor", cfg: Config{Format: FormatJSON, ContextExtractor: func(context.Context) map[string]any { return map[string]any{"req": "r1", "n": 2} }}, log: simple,
			want: `{"level":"INFO","n":2,"req":"r1","rate":"15","low":16,"high":123.2,"message":"The quick brown fox jumps over the lazy dog"}` + "\n"},
	}
}

func runSlogGolden(t *testing.T, c slogGoldenCase) string {
	t.Helper()
	buf := &syncBuffer{}
	cfg := c.cfg
	cfg.Writer = buf
	l := MustNew(cfg)
	defer l.Close()
	h := l.SlogHandler()
	if c.with != nil {
		h = c.with(h)
	}
	if c.handle != nil {
		if err := c.handle(h); err != nil {
			t.Fatalf("%s: handle: %v", c.name, err)
		}
	} else {
		c.log(slog.New(h))
	}
	return buf.String()
}

// TestSlogHandlerGolden pins the exact bytes the handler writes for the
// benchmark payloads and for the inputs around them: reserved and escaped
// keys, groups, LogValuers, every value kind, level mapping, and both
// formats. The encoding is a public contract, so a change here is a
// deliberate format change, never a side effect of an optimisation.
func TestSlogHandlerGolden(t *testing.T) {
	for _, c := range slogGoldenCases() {
		t.Run(c.name, func(t *testing.T) {
			if got := runSlogGolden(t, c); got != c.want {
				t.Fatalf("output mismatch\n got: %q\nwant: %q", got, c.want)
			}
		})
	}
}

// TestSlogHandlerGoldenSource pins the caller envelope for the WithSource
// benchmark: the record PC names the direct caller of the slog call.
func TestSlogHandlerGoldenSource(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	defer l.Close()
	line := slogGoldenSource(slog.New(l.SlogHandler()))
	want := `{"level":"INFO","func":"iqlog.slogGoldenSource","file":"slog_golden_test.go:` + strconv.Itoa(line) + `","rate":"15","low":16,"high":123.2,"message":"` + slogGoldenMsg + `"}` + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("output mismatch\n got: %q\nwant: %q", got, want)
	}

	// A hand-built record with PC zero scans the stack instead; the caller
	// is still this test.
	buf.Reset()
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "none", 0)
	_, _, here, _ := runtime.Caller(0)
	if err := l.SlogHandler().Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	want = `{"level":"INFO","func":"iqlog.TestSlogHandlerGoldenSource","file":"slog_golden_test.go:` + strconv.Itoa(here+1) + `","message":"none"}` + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("zero PC output mismatch\n got: %q\nwant: %q", got, want)
	}

	// Console shape of the same envelope.
	buf.Reset()
	c := MustNew(Config{Writer: buf, DisableColor: true, CallerDepth: 1})
	defer c.Close()
	line = slogGoldenSource(slog.New(c.SlogHandler()))
	want = "INFO [iqlog.slogGoldenSource slog_golden_test.go:" + strconv.Itoa(line) + "] rate=15 low=16 high=123.2 " + slogGoldenMsg + "\n"
	if got := buf.String(); got != want {
		t.Fatalf("console output mismatch\n got: %q\nwant: %q", got, want)
	}
}

// Handle gates on the logger level and on a closed logger itself, so a
// record fed to it directly, without slog.Logger's Enabled check in front,
// is still dropped.
func TestSlogHandlerHandleGatesDirectly(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, Level: LevelInfo})
	h := l.SlogHandler()
	ctx := context.Background()
	below := slog.NewRecord(time.Time{}, slog.LevelDebug, "dropped", 0)
	if err := h.Handle(ctx, below); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("record below the level was written: %q", got)
	}
	at := slog.NewRecord(time.Time{}, slog.LevelInfo, "kept", 0)
	if err := h.Handle(ctx, at); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != `{"level":"INFO","message":"kept"}`+"\n" {
		t.Fatalf("record at the level: %q", got)
	}
	buf.Reset()
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if err := h.Handle(ctx, at); err != nil {
		t.Fatalf("closed logger: %v", err)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("record after Close was written: %q", got)
	}
}

// Integer attributes carry exactly strconv's text for every magnitude and
// sign.
func TestSlogHandlerIntegersMatchStrconv(t *testing.T) {
	buf := &syncBuffer{}
	lg := slog.New(MustNew(Config{Format: FormatJSON, Writer: buf}).SlogHandler())
	for _, v := range []int64{0, 1, 9, 10, 11, 99, 100, 101, 999, 1000, 1001, 9999, 10000, 65535, 1 << 31, math.MaxInt64, -1, -9, -10, -99, -100, -999, -1000, math.MinInt64} {
		lg.LogAttrs(context.Background(), slog.LevelInfo, "n", slog.Int64("i", v), slog.Uint64("u", uint64(v)))
		want := `{"level":"INFO","i":` + strconv.FormatInt(v, 10) + `,"u":` + strconv.FormatUint(uint64(v), 10) + `,"message":"n"}` + "\n"
		if got := buf.String(); got != want {
			t.Fatalf("%d: %q want %q", v, got, want)
		}
		buf.Reset()
	}
}

// TestSlogHandlerGoldenDump prints every case's current output as a Go
// literal; it runs only under -v so the table above can be regenerated after
// a deliberate format change.
func TestSlogHandlerGoldenDump(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("verbose only")
	}
	for _, c := range slogGoldenCases() {
		fmt.Printf("CASE %s\n%s\n", c.name, strconv.Quote(runSlogGolden(t, c)))
	}
}
