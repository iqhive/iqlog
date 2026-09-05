package iqlog

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/slogtest"
	"time"
)

type slogNameValuer string

func (n slogNameValuer) LogValue() slog.Value {
	return slog.StringValue(string(n))
}

type slogGroupValuer struct{}

func (slogGroupValuer) LogValue() slog.Value {
	return slog.GroupValue(slog.String("a", "1"), slog.Int("b", 2))
}

func TestSlogHandlerNilLoggerPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("(*Logger)(nil).SlogHandler() did not panic")
		}
	}()
	var l *Logger
	l.SlogHandler()
}

func TestSlogHandlerEnabledAndClose(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, Level: LevelInfo})
	h := l.SlogHandler()
	ctx := context.Background()
	if h.Enabled(ctx, slog.LevelDebug) {
		t.Fatal("debug should be disabled")
	}
	if !h.Enabled(ctx, slog.LevelInfo) {
		t.Fatal("info should be enabled")
	}
	if !h.Enabled(ctx, slog.LevelError) {
		t.Fatal("error should be enabled")
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if h.Enabled(ctx, slog.LevelInfo) {
		t.Fatal("closed logger should disable Info")
	}
}

func TestSlogHandlerSetConfigLevel(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, Level: LevelInfo})
	h := l.SlogHandler()
	ctx := context.Background()
	if err := l.SetConfig(Config{Format: FormatJSON, Writer: buf, Level: LevelWarn}); err != nil {
		t.Fatal(err)
	}
	if h.Enabled(ctx, slog.LevelInfo) {
		t.Fatal("info still enabled after SetConfig to warn")
	}
	if !h.Enabled(ctx, slog.LevelWarn) {
		t.Fatal("warn disabled after SetConfig to warn")
	}
	slog.New(h).Info("dropped")
	slog.New(h).Warn("kept")
	if got := buf.String(); strings.Contains(got, "dropped") || !strings.Contains(got, "kept") {
		t.Fatalf("level change not applied: %q", got)
	}
}

func TestSlogHandlerDisabledWritesNothing(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, Level: LevelInfo})
	slog.New(l.SlogHandler()).Debug("x", "k", 1)
	if got := buf.String(); got != "" {
		t.Fatalf("debug wrote %q", got)
	}
}

func TestSlogHandlerSimpleJSON(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.New(l.SlogHandler()).Info("handled", "user", "alice", "n", 2)
	rec := lastJSONMap(t, buf)
	if rec["message"] != "handled" || rec["user"] != "alice" || rec["n"] != float64(2) {
		t.Fatalf("unexpected record: %#v", rec)
	}
	if _, ok := rec["msg"]; ok {
		t.Fatalf("unexpected msg key: %#v", rec)
	}
}

func TestSlogHandlerAllKinds(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	now := time.Date(2026, 1, 2, 3, 4, 5, 6, time.UTC)
	slog.New(l.SlogHandler()).LogAttrs(context.Background(), slog.LevelInfo, "kinds",
		slog.String("str", "value"),
		slog.Int64("i64", -2),
		slog.Uint64("u64", 4),
		slog.Float64("f64", 2.5),
		slog.Bool("ok", true),
		slog.Duration("duration", time.Second+2*time.Millisecond),
		slog.Time("when", now),
		slog.Any("err", errors.New("boom")),
		slog.Any("empty", nil),
		slog.Any("bytes", []byte("hi")),
		slog.Any("raw", json.RawMessage(`{"nested":true}`)),
		slog.Any("struct", struct{ A int }{7}),
		slog.Group("g", slog.String("k", "v")),
	)
	rec := lastJSONMap(t, buf)
	if rec["str"] != "value" || rec["i64"] != float64(-2) || rec["u64"] != float64(4) || rec["f64"] != 2.5 || rec["ok"] != true {
		t.Fatalf("unexpected primitives: %#v", rec)
	}
	if rec["duration"] != "1.002s" || rec["when"] != now.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected time fields: %#v", rec)
	}
	if rec["err"] != "boom" {
		t.Fatalf("error key: %#v", rec)
	}
	if rec["empty"] != nil {
		t.Fatalf("nil any: %#v", rec)
	}
	if rec["bytes"] != "aGk=" {
		t.Fatalf("bytes: %#v", rec)
	}
	if raw, ok := rec["raw"].(map[string]any); !ok || raw["nested"] != true {
		t.Fatalf("raw json: %#v", rec)
	}
	if st, ok := rec["struct"].(map[string]any); !ok || st["A"] != float64(7) {
		t.Fatalf("struct: %#v", rec)
	}
	if rec["g.k"] != "v" {
		t.Fatalf("group: %#v", rec)
	}
	if rec["message"] != "kinds" {
		t.Fatalf("message: %#v", rec)
	}
}

func TestSlogHandlerLogValuer(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.New(l.SlogHandler()).Info("m", "name", slogNameValuer("resolved"))
	rec := lastJSONMap(t, buf)
	if rec["name"] != "resolved" {
		t.Fatalf("LogValuer not resolved: %#v", rec)
	}
	buf.Reset()
	slog.New(l.SlogHandler().WithAttrs([]slog.Attr{slog.Any("name", slogNameValuer("stored"))})).Info("m")
	rec = lastJSONMap(t, buf)
	if rec["name"] != "stored" {
		t.Fatalf("LogValuer in WithAttrs not resolved: %#v", rec)
	}
}

// A LogValuer may resolve to a group; it has to be flattened like any other
// group rather than fall through to generic encoding of []slog.Attr.
func TestSlogHandlerLogValuerGroup(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	check := func(what string) {
		t.Helper()
		rec := lastJSONMap(t, buf)
		if rec["g.a"] != "1" || rec["g.b"] != float64(2) {
			t.Fatalf("%s: group valuer not flattened: %#v", what, rec)
		}
		if _, ok := rec["g"]; ok {
			t.Fatalf("%s: unflattened key present: %#v", what, rec)
		}
	}
	slog.New(l.SlogHandler()).Info("m", "g", slogGroupValuer{})
	check("record")
	buf.Reset()
	slog.New(l.SlogHandler().WithAttrs([]slog.Attr{slog.Any("g", slogGroupValuer{})})).Info("m")
	check("WithAttrs")
}

func TestSlogHandlerGroups(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	h := l.SlogHandler()

	slog.New(h.WithGroup("http").WithAttrs([]slog.Attr{slog.String("method", "GET")})).Info("x", "status", 200)
	rec := lastJSONMap(t, buf)
	if rec["http.method"] != "GET" || rec["http.status"] != float64(200) {
		t.Fatalf("group then attrs: %#v", rec)
	}

	buf.Reset()
	slog.New(h.WithGroup("http").WithGroup("req")).Info("x", "method", "POST")
	rec = lastJSONMap(t, buf)
	if rec["http.req.method"] != "POST" {
		t.Fatalf("nested WithGroup: %#v", rec)
	}

	buf.Reset()
	slog.New(h.WithAttrs([]slog.Attr{slog.String("svc", "api")}).WithGroup("http")).Info("x", "method", "PUT")
	rec = lastJSONMap(t, buf)
	if rec["svc"] != "api" || rec["http.method"] != "PUT" {
		t.Fatalf("attrs then group: %#v", rec)
	}

	buf.Reset()
	slog.New(h.WithGroup("http")).Info("x", slog.Group("peer", slog.String("ip", "10.0.0.1"), slog.Group("tls", slog.Int("v", 13))))
	rec = lastJSONMap(t, buf)
	if rec["http.peer.ip"] != "10.0.0.1" || rec["http.peer.tls.v"] != float64(13) {
		t.Fatalf("record groups under WithGroup: %#v", rec)
	}
}

func TestSlogHandlerEmptyWithGroup(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	h := l.SlogHandler()
	if h.WithGroup("") != h {
		t.Fatal("WithGroup(\"\") should return the same handler")
	}
	if h.WithAttrs(nil) != h {
		t.Fatal("WithAttrs(nil) should return the same handler")
	}
	slog.New(h.WithGroup("")).Info("x", "k", "v")
	rec := lastJSONMap(t, buf)
	if rec["k"] != "v" {
		t.Fatalf("empty group changed keys: %#v", rec)
	}
	if _, ok := rec[".k"]; ok {
		t.Fatalf("empty group introduced a dotted key: %#v", rec)
	}
}

func TestSlogHandlerInlineEmptyGroupKey(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.New(l.SlogHandler()).LogAttrs(context.Background(), slog.LevelInfo, "x",
		slog.String("a", "b"),
		slog.Group("", slog.String("c", "d")),
		slog.String("e", "f"),
	)
	rec := lastJSONMap(t, buf)
	if rec["a"] != "b" || rec["c"] != "d" || rec["e"] != "f" {
		t.Fatalf("inline group: %#v", rec)
	}
}

// Matches slog.JSONHandler: a fully empty Attr is dropped, an empty key with
// a value is not.
func TestSlogHandlerEmptyKeyAttrs(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.New(l.SlogHandler()).LogAttrs(context.Background(), slog.LevelInfo, "x",
		slog.Attr{},
		slog.String("", "anonymous"),
		slog.Group("empty"),
	)
	rec := lastJSONMap(t, buf)
	if rec[""] != "anonymous" {
		t.Fatalf("empty key with value dropped: %#v", rec)
	}
	if _, ok := rec["empty"]; ok {
		t.Fatalf("empty group emitted: %#v", rec)
	}
	if len(rec) != 3 {
		t.Fatalf("unexpected keys: %#v", rec)
	}
}

func TestSlogHandlerEscapedGroupName(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.New(l.SlogHandler().WithGroup(`a"b`).WithAttrs([]slog.Attr{slog.Int("n", 1)})).Info("x", slog.Group("c\nd", "k", "v"))
	rec := lastJSONMap(t, buf)
	if rec[`a"b.n`] != float64(1) || rec["a\"b.c\nd.k"] != "v" {
		t.Fatalf("escaped group names: %#v", rec)
	}
}

// A record group whose joined prefix outgrows the stack buffer still
// produces the right key.
func TestSlogHandlerLongInlineGroupPrefix(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	long := strings.Repeat("g", 150)
	slog.New(l.SlogHandler()).Info("x", slog.Group(long, slog.Group(long, "k", "v")))
	rec := lastJSONMap(t, buf)
	if rec[long+"."+long+".k"] != "v" {
		t.Fatalf("long prefix: %#v", rec)
	}
}

func TestSlogHandlerRecordTime(t *testing.T) {
	fixed := time.Date(2026, 3, 4, 5, 6, 7, 8000, time.UTC)
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, JSONTimeMode: JSONTimeUTC, Now: func() time.Time {
		t.Fatal("Config.Now consulted for a slog record")
		return time.Time{}
	}})
	r := slog.NewRecord(fixed, slog.LevelInfo, "timed", 0)
	if err := l.SlogHandler().Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	rec := lastJSONMap(t, buf)
	got, ok := rec["time"].(string)
	if !ok {
		t.Fatalf("missing time: %#v", rec)
	}
	if got != "2026-03-04T05:06:07.000008Z" {
		t.Fatalf("time %q", got)
	}

	buf.Reset()
	l2 := MustNew(Config{Format: FormatJSON, Writer: buf})
	if err := l2.SlogHandler().Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	rec = lastJSONMap(t, buf)
	if _, ok := rec["time"]; ok {
		t.Fatalf("JSONTimeDisabled still emitted time: %#v", rec)
	}

	buf.Reset()
	zero := slog.NewRecord(time.Time{}, slog.LevelInfo, "untimed", 0)
	l3 := MustNew(Config{Format: FormatJSON, Writer: buf, JSONTimeMode: JSONTimeUTC})
	if err := l3.SlogHandler().Handle(context.Background(), zero); err != nil {
		t.Fatal(err)
	}
	rec = lastJSONMap(t, buf)
	if _, ok := rec["time"]; ok {
		t.Fatalf("zero record time still emitted time: %#v", rec)
	}
}

func TestSlogHandlerConsoleRecordTime(t *testing.T) {
	fixed := time.Date(2026, 3, 4, 5, 6, 7, 8000, time.UTC)
	buf := &syncBuffer{}
	l := MustNew(Config{Writer: buf, DisableColor: true, IncludeTime: true, TimestampLayout: time.RFC3339Nano})
	r := slog.NewRecord(fixed, slog.LevelWarn, "timed", 0)
	r.AddAttrs(slog.String("k", "v"))
	if err := l.SlogHandler().Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "[2026-03-04T05:06:07.000008Z] WARN k=v timed\n"; got != want {
		t.Fatalf("console record time:\n got %q\nwant %q", got, want)
	}
}

func TestSlogHandlerSourceFollowsCallerDepth(t *testing.T) {
	buf := &syncBuffer{}
	off := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.New(off.SlogHandler()).Info("no caller")
	rec := lastJSONMap(t, buf)
	if _, ok := rec["func"]; ok {
		t.Fatalf("CallerDepth 0 emitted func: %#v", rec)
	}
	if _, ok := rec["file"]; ok {
		t.Fatalf("CallerDepth 0 emitted file: %#v", rec)
	}

	buf.Reset()
	on := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	_, _, line, _ := runtime.Caller(0)
	slog.New(on.SlogHandler()).Info("sourced")
	rec = lastJSONMap(t, buf)
	fn, _ := rec["func"].(string)
	file, _ := rec["file"].(string)
	if !strings.HasSuffix(fn, ".TestSlogHandlerSourceFollowsCallerDepth") {
		t.Fatalf("func %q", fn)
	}
	if want := "slog_test.go:" + strconv.Itoa(line+1); file != want {
		t.Fatalf("file %q want %q", file, want)
	}

	buf.Reset()
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "none", 0)
	if err := on.SlogHandler().Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	rec = lastJSONMap(t, buf)
	if _, ok := rec["func"]; ok {
		t.Fatalf("zero PC still has func: %#v", rec)
	}
	if _, ok := rec["file"]; ok {
		t.Fatalf("zero PC still has file: %#v", rec)
	}

	buf.Reset()
	console := MustNew(Config{Writer: buf, DisableColor: true, CallerDepth: 1})
	slog.New(console.SlogHandler()).Info("sourced")
	if got := buf.String(); !strings.HasPrefix(got, "INFO [") || !strings.Contains(got, ".TestSlogHandlerSourceFollowsCallerDepth slog_test.go:") {
		t.Fatalf("console caller: %q", got)
	}
}

func slogInlineHelper(lg *slog.Logger) { lg.Info("from helper") }

// The record PC is a return address. It has to be attributed to the function
// containing the call, including when that function was inlined into its
// caller, rather than to whatever the PC after the call belongs to.
func TestSlogHandlerSourceAttribution(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	slogInlineHelper(slog.New(l.SlogHandler()))
	rec := lastJSONMap(t, buf)
	fn, _ := rec["func"].(string)
	if !strings.HasSuffix(fn, ".slogInlineHelper") {
		t.Fatalf("func %q", fn)
	}
}

func slogFunctionWithANameLongEnoughToOverflowTheHundredByteCallerBufferWhenQualifiedByThePackagePath(lg *slog.Logger) {
	lg.Info("long")
}

func TestSlogHandlerLongFunctionNameTruncated(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	slogFunctionWithANameLongEnoughToOverflowTheHundredByteCallerBufferWhenQualifiedByThePackagePath(slog.New(l.SlogHandler()))
	rec := lastJSONMap(t, buf)
	fn, _ := rec["func"].(string)
	if len(fn) != callerDataMaxLen {
		t.Fatalf("func length %d want %d: %q", len(fn), callerDataMaxLen, fn)
	}
	if !strings.Contains(fn, "slogFunctionWithANameLongEnough") {
		t.Fatalf("func %q", fn)
	}
}

func TestSlogHandlerLevelMapping(t *testing.T) {
	cases := []struct {
		level slog.Level
		want  string
	}{
		{slog.LevelDebug - 4, "TRACE"},
		{slog.LevelDebug - 1, "TRACE"},
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo - 1, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn - 1, "INFO"},
		{slog.LevelWarn, "WARN"},
		{slog.LevelError - 1, "WARN"},
		{slog.LevelError, "ERROR"},
		{slog.LevelError + 8, "ERROR"},
	}
	buf := &syncBuffer{}
	exited := false
	l := MustNew(Config{Format: FormatJSON, Writer: buf, Level: LevelTrace, ExitFunc: func(int) { exited = true }})
	lg := slog.New(l.SlogHandler())
	for _, c := range cases {
		buf.Reset()
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("level %v panicked: %v", c.level, r)
				}
			}()
			lg.Log(context.Background(), c.level, "m")
		}()
		if exited {
			t.Fatalf("level %v terminated the process", c.level)
		}
		rec := lastJSONMap(t, buf)
		if rec["level"] != c.want {
			t.Fatalf("level %v: got %v want %s", c.level, rec["level"], c.want)
		}
	}
}

func TestSlogHandlerConsoleShape(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Writer: buf, DisableColor: true})
	h := l.SlogHandler().WithGroup("http").WithAttrs([]slog.Attr{slog.String("method", "GET")})
	slog.New(h).Info("handled", "status", 200, slog.Group("peer", "ip", "10.0.0.1"), "ok", true, "ratio", 0.5)
	if got, want := buf.String(), "INFO http.method=GET http.status=200 http.peer.ip=10.0.0.1 http.ok=true http.ratio=0.5 handled\n"; got != want {
		t.Fatalf("console shape:\n got %q\nwant %q", got, want)
	}
	buf.Reset()
	slog.New(l.SlogHandler()).Error("failed", "err", errors.New("boom"), "message", "x")
	if got, want := buf.String(), "ERRR err=boom field_message=x failed\n"; got != want {
		t.Fatalf("console error:\n got %q\nwant %q", got, want)
	}
}

func TestSlogHandlerFieldOrder(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf}).WithFields(map[string]any{"svc": "api"})
	h := l.SlogHandler().WithAttrs([]slog.Attr{slog.String("pre", "1")})
	slog.New(h).Info("m", "rec", "2")
	got := buf.String()
	svc, pre, rec, msg := strings.Index(got, `"svc"`), strings.Index(got, `"pre"`), strings.Index(got, `"rec"`), strings.Index(got, `"message"`)
	if svc < 0 || pre < 0 || rec < 0 || msg < 0 || !(svc < pre && pre < rec && rec < msg) {
		t.Fatalf("field order: %q", got)
	}
}

func TestSlogHandlerReservedKeyInGroup(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.New(l.SlogHandler().WithGroup("g")).Info("envelope", "message", "x", "time", "y")
	rec := lastJSONMap(t, buf)
	if rec["message"] != "envelope" || rec["g.message"] != "x" || rec["g.time"] != "y" {
		t.Fatalf("grouped reserved names: %#v", rec)
	}
}

func TestSlogHandlerBuildError(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "raw", 0)
	r.AddAttrs(slog.Any("payload", json.RawMessage(`{bad`)))
	err := l.SlogHandler().Handle(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), `"payload"`) {
		t.Fatalf("invalid raw JSON did not surface a build error naming the field: %v", err)
	}
	rec := lastJSONMap(t, buf)
	if rec["payload"] != nil || rec["message"] != "raw" {
		t.Fatalf("record after build error: %#v", rec)
	}

	buf.Reset()
	h := l.SlogHandler().WithAttrs([]slog.Attr{slog.Any("payload", json.RawMessage(`{bad`))})
	if err := h.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "pre", 0)); err == nil {
		t.Fatal("invalid raw JSON in WithAttrs did not surface a build error")
	}
	rec = lastJSONMap(t, buf)
	if rec["payload"] != nil {
		t.Fatalf("pre-encoded record after build error: %#v", rec)
	}
}

func TestSlogHandlerWithAttrsIsolation(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	base := l.SlogHandler()
	a := base.WithAttrs([]slog.Attr{slog.String("a", "1")})
	b := base.WithAttrs([]slog.Attr{slog.String("b", "2")})
	ac := a.WithAttrs([]slog.Attr{slog.String("c", "3")})

	slog.New(a).Info("m")
	rec := lastJSONMap(t, buf)
	if rec["a"] != "1" || rec["b"] != nil || rec["c"] != nil {
		t.Fatalf("a: %#v", rec)
	}
	buf.Reset()
	slog.New(b).Info("m")
	rec = lastJSONMap(t, buf)
	if rec["b"] != "2" || rec["a"] != nil {
		t.Fatalf("b: %#v", rec)
	}
	buf.Reset()
	slog.New(ac).Info("m")
	rec = lastJSONMap(t, buf)
	if rec["a"] != "1" || rec["c"] != "3" {
		t.Fatalf("ac: %#v", rec)
	}
	buf.Reset()
	slog.New(base).Info("m")
	rec = lastJSONMap(t, buf)
	if len(rec) != 2 {
		t.Fatalf("base gained attrs: %#v", rec)
	}
}

// Attributes are pre-encoded in both formats, so a handler built while the
// logger was in JSON mode keeps working after a switch to console output.
func TestSlogHandlerPreEncodedSurvivesFormatSwitch(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	h := l.SlogHandler().WithGroup("g").WithAttrs([]slog.Attr{slog.String("a", "1")})
	slog.New(h).Info("m")
	if rec := lastJSONMap(t, buf); rec["g.a"] != "1" {
		t.Fatalf("json: %#v", rec)
	}
	if err := l.SetConfig(Config{Format: FormatConsole, Writer: buf, DisableColor: true}); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	slog.New(h).Info("m", "b", 2)
	if got, want := buf.String(), "INFO g.a=1 g.b=2 m\n"; got != want {
		t.Fatalf("console after switch:\n got %q\nwant %q", got, want)
	}
}

func TestSlogHandlerContextExtractor(t *testing.T) {
	type ctxKey struct{}
	buf := &syncBuffer{}
	base := MustNew(Config{Format: FormatJSON, Writer: buf, ContextExtractor: func(ctx context.Context) map[string]any {
		return map[string]any{"request": ctx.Value(ctxKey{})}
	}})
	wrong := context.WithValue(context.Background(), ctxKey{}, "wrong")
	right := context.WithValue(context.Background(), ctxKey{}, "right")
	h := base.WithContext(wrong).SlogHandler()
	slog.New(h).InfoContext(right, "m")
	rec := lastJSONMap(t, buf)
	if rec["request"] != "right" {
		t.Fatalf("Handle ctx lost: %#v", rec)
	}
}

func TestSlogHandlerLoggerFields(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf}).WithFields(map[string]any{"svc": "api"})
	slog.New(l.SlogHandler()).Info("m")
	rec := lastJSONMap(t, buf)
	if rec["svc"] != "api" {
		t.Fatalf("persistent fields: %#v", rec)
	}
}

func TestSlogHandlerColor(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Writer: buf, Color: true})
	slog.New(l.SlogHandler()).Error("fail")
	if !strings.Contains(buf.String(), "\x1b[31m") {
		t.Fatalf("error colour missing: %q", buf.String())
	}
	buf.Reset()
	slog.New(l.SlogHandler()).Info("ok")
	if !strings.Contains(buf.String(), "\x1b[34m") {
		t.Fatalf("info colour missing: %q", buf.String())
	}

	jbuf := &syncBuffer{}
	jl := MustNew(Config{Format: FormatJSON, Writer: jbuf, Color: true})
	slog.New(jl.SlogHandler()).Error("fail")
	if strings.Contains(jbuf.String(), "\x1b") {
		t.Fatalf("JSON contained ESC: %q", jbuf.String())
	}
}

func TestSlogHandlerReservedKey(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.New(l.SlogHandler()).Info("envelope", "message", "x")
	rec := lastJSONMap(t, buf)
	if rec["message"] != "envelope" {
		t.Fatalf("envelope message overwritten: %#v", rec)
	}
	if rec["field_message"] != "x" {
		t.Fatalf("reserved key not prefixed: %#v", rec)
	}
}

func TestPackageSlogHandlerFollowsSetDefault(t *testing.T) {
	old := Default()
	t.Cleanup(func() { SetDefault(old) })

	first := &syncBuffer{}
	SetDefault(MustNew(Config{Format: FormatJSON, Writer: first}))
	logger := slog.New(SlogHandler())
	logger.Info("hi")
	if !strings.Contains(first.String(), `"message":"hi"`) {
		t.Fatalf("first default: %q", first.String())
	}

	second := &syncBuffer{}
	SetDefault(MustNew(Config{Format: FormatJSON, Writer: second}))
	logger.Info("there")
	if second.String() == "" || strings.Contains(first.String(), `"message":"there"`) {
		t.Fatalf("SetDefault not followed: first=%q second=%q", first.String(), second.String())
	}

	// A handler bound to a specific logger keeps that logger.
	bound := slog.New(Default().SlogHandler())
	third := &syncBuffer{}
	SetDefault(MustNew(Config{Format: FormatJSON, Writer: third}))
	bound.Info("bound")
	if !strings.Contains(second.String(), `"message":"bound"`) || third.String() != "" {
		t.Fatalf("bound handler moved: second=%q third=%q", second.String(), third.String())
	}
}

// slog.SetDefault also routes the standard log package into the handler.
func TestSlogHandlerStdLogBridge(t *testing.T) {
	oldSlog := slog.Default()
	oldOut, oldFlags := log.Writer(), log.Flags()
	t.Cleanup(func() {
		slog.SetDefault(oldSlog)
		log.SetOutput(oldOut)
		log.SetFlags(oldFlags)
	})
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	slog.SetDefault(slog.New(l.SlogHandler()))
	log.Printf("bridged %d", 1)
	rec := lastJSONMap(t, buf)
	if rec["level"] != "INFO" || rec["message"] != "bridged 1" {
		t.Fatalf("std log bridge: %#v", rec)
	}
}

func TestSlogHandlerClosedDoesNotPanic(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf})
	logger := slog.New(l.SlogHandler())
	logger.Info("before")
	before := buf.String()
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	logger.Info("after")
	if buf.String() != before {
		t.Fatalf("closed logger still wrote: %q", buf.String())
	}
	if err := l.SlogHandler().Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "direct", 0)); err != nil {
		t.Fatalf("Handle after close returned %v", err)
	}
}

func TestSlogHandlerAsyncFlushAndClose(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, WriterMode: WriterAsync})
	logger := slog.New(l.SlogHandler().WithAttrs([]slog.Attr{slog.String("mode", "async")}))
	for i := 0; i < 50; i++ {
		logger.Info("queued", "i", i)
	}
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 50 {
		t.Fatalf("after flush got %d lines", len(lines))
	}
	for i, line := range lines {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("line %d: %v: %q", i, err, line)
		}
		if rec["mode"] != "async" || rec["i"] != float64(i) {
			t.Fatalf("line %d out of order or missing attrs: %#v", i, rec)
		}
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	logger.Info("after close")
	if strings.Contains(buf.String(), "after close") {
		t.Fatal("record written after Close")
	}
}

type slogBlockingWriter struct {
	unblock chan struct{}
}

func (w *slogBlockingWriter) Write(p []byte) (int, error) {
	<-w.unblock
	return len(p), nil
}

func TestSlogHandlerOverflowDropCounted(t *testing.T) {
	w := &slogBlockingWriter{unblock: make(chan struct{})}
	l := MustNew(Config{Format: FormatJSON, Writer: w, WriterMode: WriterAsync, BufferSize: 1, OverflowPolicy: OverflowDrop})
	logger := slog.New(l.SlogHandler())
	for i := 0; i < 20; i++ {
		logger.Info("burst", "i", i)
	}
	if l.Dropped() == 0 {
		t.Fatal("no records counted as dropped")
	}
	close(w.unblock)
	if err := l.Close(); err != nil && !errors.Is(err, ErrWriteDropped) {
		t.Fatal(err)
	}
}

func TestSlogHandlerRace(t *testing.T) {
	l := MustNew(Config{Format: FormatJSON, Writer: io.Discard})
	h := l.SlogHandler()
	logger := slog.New(h)
	shared := h.WithGroup("g").WithAttrs([]slog.Attr{slog.String("shared", "yes")})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			child := slog.New(shared.WithAttrs([]slog.Attr{slog.Int("i", i)}).WithGroup("g"))
			for j := 0; j < 200; j++ {
				logger.Info("x", "n", j)
				child.Info("y", "k", j)
				slog.New(shared).Info("z")
			}
		}(i)
	}
	wg.Wait()
}

func TestSlogHandlerEnabledDoesNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	l := MustNew(Config{Format: FormatJSON, Writer: io.Discard, Level: LevelInfo})
	h := l.SlogHandler()
	ctx := context.Background()
	if got := testing.AllocsPerRun(1000, func() {
		_ = h.Enabled(ctx, slog.LevelDebug)
	}); got != 0 {
		t.Fatalf("Enabled allocated %.2f times", got)
	}
	logger := slog.New(h)
	if got := testing.AllocsPerRun(1000, func() {
		logger.Debug("m", "k", "v")
	}); got != 0 {
		t.Fatalf("disabled slog.Logger call allocated %.2f times", got)
	}
}

// The handler must stay allocation-free for records made of typed
// attributes, in every configuration the logger supports, including caller
// capture from the record PC. Duration is excluded: time.Duration.String
// allocates, for native events as well.
func TestSlogHandlerHandleDoesNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	ctx := context.Background()
	now := time.Date(2026, 1, 2, 3, 4, 5, 6, time.UTC)
	boom := errors.New("boom")
	// slog.Group boxes its arguments, so the attr is built once here rather
	// than inside the measured call.
	peer := slog.Group("peer", slog.String("ip", "10.0.0.1"), slog.Int("port", 443))
	cases := []struct {
		name string
		cfg  Config
		with func(slog.Handler) slog.Handler
		log  func(*slog.Logger)
	}{
		{"json typed", Config{Format: FormatJSON, Writer: io.Discard}, nil, func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2), slog.Bool("ok", true))
		}},
		{"json time, error and bytes", Config{Format: FormatJSON, Writer: io.Discard}, nil, func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.Time("when", now), slog.Any("err", boom), slog.Uint64("u", 7))
		}},
		{"json group and pre-attached attrs", Config{Format: FormatJSON, Writer: io.Discard}, func(h slog.Handler) slog.Handler {
			return h.WithGroup("http").WithAttrs([]slog.Attr{slog.String("method", "GET"), slog.String("path", "/x")})
		}, func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.Int("status", 200))
		}},
		{"json inline record group", Config{Format: FormatJSON, Writer: io.Discard}, nil, func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", peer)
		}},
		{"json utc timestamp", Config{Format: FormatJSON, Writer: io.Discard, JSONTimeMode: JSONTimeUTC}, nil, func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.String("rate", "15"), slog.Int("low", 16))
		}},
		{"json caller from record pc", Config{Format: FormatJSON, Writer: io.Discard, CallerDepth: 1}, nil, func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.String("rate", "15"))
		}},
		{"console typed with timestamp", Config{Writer: io.Discard, DisableColor: true, IncludeTime: true}, func(h slog.Handler) slog.Handler {
			return h.WithGroup("http").WithAttrs([]slog.Attr{slog.String("method", "GET")})
		}, func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelInfo, "m", slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
		}},
		{"console colour", Config{Writer: io.Discard, Color: true}, nil, func(lg *slog.Logger) {
			lg.LogAttrs(ctx, slog.LevelError, "m", slog.String("rate", "15"), slog.Bool("ok", false))
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l := MustNew(c.cfg)
			h := l.SlogHandler()
			if c.with != nil {
				h = c.with(h)
			}
			lg := slog.New(h)
			c.log(lg) // warm the pools
			if got := testing.AllocsPerRun(1000, func() { c.log(lg) }); got != 0 {
				t.Fatalf("allocated %.2f times per record", got)
			}
		})
	}
}

func TestSlogHandlerSlogtest(t *testing.T) {
	var buf *syncBuffer
	slogtest.Run(t, func(t *testing.T) slog.Handler {
		buf = &syncBuffer{}
		l := MustNew(Config{Format: FormatJSON, Writer: buf, JSONTimeMode: JSONTimeUTC})
		t.Cleanup(func() { _ = l.Close() })
		return l.SlogHandler()
	}, func(t *testing.T) map[string]any {
		return remapSlogtestRecord(t, buf)
	})
}

func lastJSONMap(t *testing.T, buf *syncBuffer) map[string]any {
	t.Helper()
	s := strings.TrimSpace(buf.String())
	if s == "" {
		t.Fatal("empty output")
	}
	lines := strings.Split(s, "\n")
	last := lines[len(lines)-1]
	var rec map[string]any
	if err := json.Unmarshal([]byte(last), &rec); err != nil {
		t.Fatalf("json: %v in %q", err, last)
	}
	return rec
}

func remapSlogtestRecord(t *testing.T, buf *syncBuffer) map[string]any {
	t.Helper()
	raw := lastJSONMap(t, buf)
	out := make(map[string]any, len(raw)+1)
	if msg, ok := raw["message"]; ok {
		out[slog.MessageKey] = msg
	}
	if ts, ok := raw["time"].(string); ok {
		parsed, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			t.Fatalf("time %q: %v", ts, err)
		}
		out[slog.TimeKey] = parsed
	}
	if lv, ok := raw["level"].(string); ok {
		out[slog.LevelKey] = slogtestLevel(lv)
	}
	for k, v := range raw {
		switch k {
		case "time", "level", "message", "func", "file":
			continue
		}
		if strings.Contains(k, ".") {
			setNested(out, strings.Split(k, "."), v)
		} else {
			out[k] = v
		}
	}
	return out
}

func slogtestLevel(name string) slog.Level {
	switch name {
	case "TRACE":
		return slog.LevelDebug - 4
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func setNested(m map[string]any, parts []string, v any) {
	for i := 0; i < len(parts)-1; i++ {
		p := parts[i]
		next, ok := m[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[p] = next
		}
		m = next
	}
	m[parts[len(parts)-1]] = v
}
