package iqlog

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"
	"time"
)

// The Msgf output format is a public contract. These cases pin the exact
// bytes Msgf produces for the comparative benchmark payload and for the
// neighbouring inputs that exercise every branch fmt takes -- clean text,
// text that needs JSON or console escaping, every plain verb on every basic
// type, and the diagnostic forms fmt emits for malformed formats -- so the
// formatting path can be optimised without changing a single byte.
//
// Each case is run three times: JSON with a fixed UTC timestamp, JSON with
// the timestamp disabled (the benchmark configuration), and console with a
// fixed timestamp. The expected records were captured from the
// implementation before the Msgf fast path existed.

type goldenStringer int

func (s goldenStringer) String() string { return "stringer!" }

type goldenNamedInt int

type goldenNamedStr string

type goldenFormatter struct{}

func (goldenFormatter) Format(f fmt.State, verb rune) {
	fmt.Fprintf(f, "formatted:%c", verb)
}

// goldenFormat hides a format string from vet's printf checker so the
// deliberately malformed cases below can pin fmt's diagnostic output.
func goldenFormat(s string) string { return s }

// goldenArgs likewise hides an empty argument list from vet.
func goldenArgs(a ...any) []any { return a }

var msgfGoldenClock = time.Date(2026, time.September, 24, 8, 30, 15, 123456789, time.UTC)

type msgfGoldenCase struct {
	name string
	log  func(l *Logger)
	// panics is the panic value a PanicEvent case must raise; exits marks a
	// FatalEvent case, which must call ExitFunc with status 1.
	panics string
	exits  bool
}

var msgfGoldenCases = []msgfGoldenCase{
	// the comparative benchmark payload, exactly as bench/benchmark_test.go logs it
	{name: "bench payload", log: func(l *Logger) {
		l.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("The race was %d minutes and %d seconds long", 16, 32)
	}},
	{name: "bench payload no fields", log: func(l *Logger) {
		l.InfoEvent().Msgf("The race was %d minutes and %d seconds long", 16, 32)
	}},
	{name: "no args percent", log: func(l *Logger) { l.InfoEvent().Msgf("100%% complete") }},
	{name: "no args no percent", log: func(l *Logger) { l.InfoEvent().Msgf("plain message") }},
	{name: "empty format", log: func(l *Logger) { l.InfoEvent().Msgf("") }},
	{name: "empty format with arg", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat(""), 1) }},
	{name: "empty string arg", log: func(l *Logger) { l.InfoEvent().Msgf("[%s]", "") }},
	{name: "percent literal around verb", log: func(l *Logger) { l.InfoEvent().Msgf("%%d %d%% %%", 1) }},
	{name: "only verb", log: func(l *Logger) { l.InfoEvent().Msgf("%d", 7) }},
	{name: "adjacent verbs", log: func(l *Logger) { l.InfoEvent().Msgf("%d%s%v", 1, "a", 2) }},
	{name: "escape needed in arg", log: func(l *Logger) {
		l.InfoEvent().Msgf("%s", "quote \" backslash \\ newline \n tab \t cr \r bell \x07 del \x7f end")
	}},
	{name: "escape needed in format", log: func(l *Logger) {
		l.InfoEvent().Msgf("say \"%s\" and\n%d", "hi", 3)
	}},
	{name: "escape needed at end", log: func(l *Logger) { l.InfoEvent().Msgf("%s", "trailing quote\"") }},
	{name: "escape needed at start", log: func(l *Logger) { l.InfoEvent().Msgf("\"%d", 5) }},
	{name: "utf8", log: func(l *Logger) { l.InfoEvent().Msgf("%s and %d", "héllo wörld ✓ 日本", -42) }},
	{name: "invalid utf8", log: func(l *Logger) { l.InfoEvent().Msgf("%s|%s", "bad\xffutf8", "\xc3") }},
	{name: "c1 control", log: func(l *Logger) { l.InfoEvent().Msgf("%s", "c1 \xc2\x85 next") }},
	{name: "line separator", log: func(l *Logger) { l.InfoEvent().Msgf("%s", "sep \u2028\u2029 end") }},
	{name: "v basic", log: func(l *Logger) { l.InfoEvent().Msgf("%v %v %v %v %v", "str", 7, true, false, nil) }},
	{name: "t bool", log: func(l *Logger) { l.InfoEvent().Msgf("%t %t %v", true, false, true) }},
	{name: "d all int types", log: func(l *Logger) {
		l.InfoEvent().Msgf("%d %d %d %d %d %d %d %d %d %d %d", int8(-8), int16(-16), int32(-32), int64(-64), uint(1), uint8(8), uint16(16), uint32(32), uint64(64), math.MaxInt64, math.MinInt64)
	}},
	{name: "v all int types", log: func(l *Logger) {
		l.InfoEvent().Msgf("%v %v %v %v %v %v %v %v %v %v %v", int8(-8), int16(-16), int32(-32), int64(-64), uint(1), uint8(8), uint16(16), uint32(32), uint64(math.MaxUint64), 0, -0)
	}},
	{name: "d uintptr", log: func(l *Logger) { l.InfoEvent().Msgf("%d %v", uintptr(7), uintptr(8)) }},
	{name: "s bytes", log: func(l *Logger) { l.InfoEvent().Msgf("%s %v", []byte("bytes"), []byte("b2")) }},
	{name: "floats", log: func(l *Logger) {
		l.InfoEvent().Msgf("%f %.2f %g %e %v %v %v", 3.14159, 2.5, 1e21, 123456.789, float32(0.1), math.Inf(1), math.NaN())
	}},
	{name: "float verbs", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%d %s %f"), 1.5, 2.5, 3) }},
	{name: "flags width precision", log: func(l *Logger) {
		l.InfoEvent().Msgf("%5d|%-5d|%05d|%+d|% d|%x|%X|%o|%b|%c|%q|%U|%08.3f|%.3s|%#x|%#v", 42, 42, 42, 42, 42, 255, 255, 8, 5, 'A', "q\"", 0x1F600, 3.14159, "abcdef", 255, "s")
	}},
	{name: "missing arg", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%s %d"), "only-one") }},
	{name: "missing all args", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%s %d"), goldenArgs()...) }},
	{name: "extra arg", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%s"), "a", "b") }},
	{name: "extra args no verbs", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("literal"), 1, "two", nil) }},
	{name: "bad verb type", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%d|%s|%t|%v"), "notint", 5, 1, struct{}{}) }},
	{name: "unknown verb", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%z %y"), 1, "a") }},
	{name: "no verb trailing", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("trailing %"), 1) }},
	{name: "no verb trailing no args", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("trailing %"), goldenArgs()...) }},
	{name: "non-ascii verb", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%é"), 1) }},
	{name: "arg index", log: func(l *Logger) { l.InfoEvent().Msgf("%[2]d %[1]d %d", 1, 2) }},
	{name: "bad arg index", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%[5]d"), 1) }},
	{name: "star width", log: func(l *Logger) { l.InfoEvent().Msgf("%*d|%-*d", 5, 42, 4, 7) }},
	{name: "nil s", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%s %d"), nil, nil) }},
	{name: "struct", log: func(l *Logger) {
		v := struct {
			A int
			B string
		}{1, "x"}
		l.InfoEvent().Msgf("%v %+v %#v", v, v, v)
	}},
	{name: "slice map", log: func(l *Logger) {
		l.InfoEvent().Msgf("%v %v %d", []int{1, 2, 3}, map[string]int{"b": 2, "a": 1}, []int{4, 5})
	}},
	{name: "pointer to struct", log: func(l *Logger) {
		v := &struct{ A int }{9}
		l.InfoEvent().Msgf("%v %+v", v, v)
	}},
	{name: "error and stringer", log: func(l *Logger) {
		l.InfoEvent().Msgf("%s %v %d %s", errors.New("err!"), goldenStringer(3), goldenStringer(4), goldenStringer(5))
	}},
	{name: "wrapped error verb", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("wrapped: %w"), errors.New("inner")) }},
	{name: "formatter", log: func(l *Logger) {
		l.InfoEvent().Msgf("%v %s %d", goldenFormatter{}, goldenFormatter{}, goldenFormatter{})
	}},
	{name: "named types", log: func(l *Logger) {
		l.InfoEvent().Msgf("%d %v %s %v", goldenNamedInt(5), goldenNamedInt(6), goldenNamedStr("ns"), goldenNamedStr("ns2"))
	}},
	{name: "duration and time", log: func(l *Logger) {
		l.InfoEvent().Msgf("%v %s %d", 1500*time.Millisecond, 2*time.Second, time.Duration(3))
	}},
	{name: "rune and byte", log: func(l *Logger) { l.InfoEvent().Msgf(goldenFormat("%c %d %v %s"), 'x', 'y', 'z', 'w') }},
	{name: "type verb", log: func(l *Logger) { l.InfoEvent().Msgf("%T %T %T", 1, "s", nil) }},
	{name: "long hostile", log: func(l *Logger) {
		l.InfoEvent().Msgf("prefix %s END", strings.Repeat(`a"b\c`, 300))
	}},
	{name: "long clean", log: func(l *Logger) {
		l.InfoEvent().Msgf("prefix %s %d END", strings.Repeat("abcdefgh", 300), 1)
	}},
	{name: "spaces", log: func(l *Logger) { l.InfoEvent().Msgf(" %d ", 5) }},
	{name: "with fields hostile", log: func(l *Logger) {
		l.InfoEvent().Str("k", "v\"q").Int("n", -1).Bool("b", true).Msgf("%s=%d", "x\ty", 2)
	}},
	{name: "warn level", log: func(l *Logger) { l.WarnEvent().Msgf("warn %d", 1) }},
	{name: "error level", log: func(l *Logger) { l.ErrorEvent().Msgf("error %s", "e") }},
	{name: "below level", log: func(l *Logger) { l.DebugEvent().Msgf("debug %d", 1) }},
	{name: "panic simple", panics: "boom 42", log: func(l *Logger) { l.PanicEvent().Msgf("boom %d", 42) }},
	{name: "panic complex", panics: "boom    42 x %!d(string=y)", log: func(l *Logger) {
		l.PanicEvent().Msgf(goldenFormat("boom %5d %s %d"), 42, "x", "y")
	}},
	{name: "panic no args", panics: "boom", log: func(l *Logger) { l.PanicEvent().Msgf("boom") }},
	{name: "fatal", exits: true, log: func(l *Logger) { l.FatalEvent().Msgf("fatal %s", "bye") }},
}

func TestMsgfGolden(t *testing.T) {
	configs := []struct {
		name string
		cfg  Config
	}{
		{"json", Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC}},
		{"json-notime", Config{Format: FormatJSON, JSONTimeMode: JSONTimeDisabled}},
		{"console", Config{Format: FormatConsole, IncludeTime: true}},
	}
	for _, c := range msgfGoldenCases {
		for i, cfg := range configs {
			t.Run(c.name+"/"+cfg.name, func(t *testing.T) {
				want, ok := msgfGoldenWant[c.name]
				if !ok {
					t.Fatalf("no golden record for %q", c.name)
				}
				got, panicked, exited := runMsgfGoldenCase(c, cfg.cfg)
				if got != want[i] {
					t.Fatalf("record mismatch\n got: %q\nwant: %q", got, want[i])
				}
				if panicked != c.panics {
					t.Fatalf("panic value = %q, want %q", panicked, c.panics)
				}
				if exited != c.exits {
					t.Fatalf("exited = %v, want %v", exited, c.exits)
				}
			})
		}
	}
	for name := range msgfGoldenWant {
		found := false
		for _, c := range msgfGoldenCases {
			found = found || c.name == name
		}
		if !found {
			t.Errorf("golden record %q has no case", name)
		}
	}
}

// runMsgfGoldenCase logs one case through a fresh logger with a fixed clock
// and reports the record, the recovered panic value, and whether ExitFunc ran.
func runMsgfGoldenCase(c msgfGoldenCase, cfg Config) (record, panicked string, exited bool) {
	sb := &syncBuffer{}
	cfg.Writer = sb
	cfg.Level = LevelInfo
	cfg.Now = func() time.Time { return msgfGoldenClock }
	cfg.ExitFunc = func(int) { exited = true }
	l := MustNew(cfg)
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked, _ = r.(string)
			}
		}()
		c.log(l)
	}()
	return sb.String(), panicked, exited
}

// The Msgf fast path must stay allocation-free for the benchmark payload
// and for the plain-verb formats it handles, and the fmt fallback must not
// add allocations beyond fmt's own zero-allocation formatting either.
func TestMsgfDoesNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	for _, format := range []Format{FormatJSON, FormatConsole} {
		l := MustNew(Config{Format: format, Writer: io.Discard, JSONTimeMode: JSONTimeDisabled})
		if got := testing.AllocsPerRun(1000, func() {
			l.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("The race was %d minutes and %d seconds long", 16, 32)
		}); got != 0 {
			t.Fatalf("%v Msgf benchmark payload allocated %.2f times", format, got)
		}
		if got := testing.AllocsPerRun(1000, func() {
			l.InfoEvent().Msgf("%s said \"%v\" %d times, %t", "she", "hi", 3, true)
		}); got != 0 {
			t.Fatalf("%v Msgf with escaping allocated %.2f times", format, got)
		}
		if got := testing.AllocsPerRun(1000, func() {
			l.InfoEvent().Msgf("%5d %.2f", 3, 2.5)
		}); got != 0 {
			t.Fatalf("%v Msgf fmt fallback allocated %.2f times", format, got)
		}
	}
}

// msgfGoldenWant holds, per case, the JSON record with a fixed UTC
// timestamp, the JSON record with the timestamp disabled, and the console
// record with a fixed timestamp.
var msgfGoldenWant = map[string][3]string{
	"bench payload": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"rate\":\"15\",\"low\":16,\"high\":123.2,\"message\":\"The race was 16 minutes and 32 seconds long\"}\n",
		"{\"level\":\"INFO\",\"rate\":\"15\",\"low\":16,\"high\":123.2,\"message\":\"The race was 16 minutes and 32 seconds long\"}\n",
		"[2026-09-24T08:30:15.123456] INFO rate=15 low=16 high=123.2 The race was 16 minutes and 32 seconds long\n",
	},
	"bench payload no fields": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"The race was 16 minutes and 32 seconds long\"}\n",
		"{\"level\":\"INFO\",\"message\":\"The race was 16 minutes and 32 seconds long\"}\n",
		"[2026-09-24T08:30:15.123456] INFO The race was 16 minutes and 32 seconds long\n",
	},
	"no args percent": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"100% complete\"}\n",
		"{\"level\":\"INFO\",\"message\":\"100% complete\"}\n",
		"[2026-09-24T08:30:15.123456] INFO 100% complete\n",
	},
	"no args no percent": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"plain message\"}\n",
		"{\"level\":\"INFO\",\"message\":\"plain message\"}\n",
		"[2026-09-24T08:30:15.123456] INFO plain message\n",
	},
	"empty format": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"\"}\n",
		"{\"level\":\"INFO\",\"message\":\"\"}\n",
		"[2026-09-24T08:30:15.123456] INFO \n",
	},
	"empty format with arg": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%!(EXTRA int=1)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%!(EXTRA int=1)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %!(EXTRA int=1)\n",
	},
	"empty string arg": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"[]\"}\n",
		"{\"level\":\"INFO\",\"message\":\"[]\"}\n",
		"[2026-09-24T08:30:15.123456] INFO []\n",
	},
	"percent literal around verb": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%d 1% %\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%d 1% %\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %d 1% %\n",
	},
	"only verb": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"7\"}\n",
		"{\"level\":\"INFO\",\"message\":\"7\"}\n",
		"[2026-09-24T08:30:15.123456] INFO 7\n",
	},
	"adjacent verbs": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"1a2\"}\n",
		"{\"level\":\"INFO\",\"message\":\"1a2\"}\n",
		"[2026-09-24T08:30:15.123456] INFO 1a2\n",
	},
	"escape needed in arg": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"quote \\\" backslash \\\\ newline \\n tab \\t cr \\r bell \\u0007 del \x7f end\"}\n",
		"{\"level\":\"INFO\",\"message\":\"quote \\\" backslash \\\\ newline \\n tab \\t cr \\r bell \\u0007 del \x7f end\"}\n",
		"[2026-09-24T08:30:15.123456] INFO quote \" backslash \\ newline \\n tab \t cr \\r bell \\x07 del \\x7f end\n",
	},
	"escape needed in format": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"say \\\"hi\\\" and\\n3\"}\n",
		"{\"level\":\"INFO\",\"message\":\"say \\\"hi\\\" and\\n3\"}\n",
		"[2026-09-24T08:30:15.123456] INFO say \"hi\" and\\n3\n",
	},
	"escape needed at end": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"trailing quote\\\"\"}\n",
		"{\"level\":\"INFO\",\"message\":\"trailing quote\\\"\"}\n",
		"[2026-09-24T08:30:15.123456] INFO trailing quote\"\n",
	},
	"escape needed at start": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"\\\"5\"}\n",
		"{\"level\":\"INFO\",\"message\":\"\\\"5\"}\n",
		"[2026-09-24T08:30:15.123456] INFO \"5\n",
	},
	"utf8": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"héllo wörld ✓ 日本 and -42\"}\n",
		"{\"level\":\"INFO\",\"message\":\"héllo wörld ✓ 日本 and -42\"}\n",
		"[2026-09-24T08:30:15.123456] INFO héllo wörld ✓ 日本 and -42\n",
	},
	"invalid utf8": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"bad\\ufffdutf8|\\ufffd\"}\n",
		"{\"level\":\"INFO\",\"message\":\"bad\\ufffdutf8|\\ufffd\"}\n",
		"[2026-09-24T08:30:15.123456] INFO bad\xffutf8|\xc3\n",
	},
	"c1 control": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"c1 \u0085 next\"}\n",
		"{\"level\":\"INFO\",\"message\":\"c1 \u0085 next\"}\n",
		"[2026-09-24T08:30:15.123456] INFO c1 \\u0085 next\n",
	},
	"line separator": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"sep \u2028\u2029 end\"}\n",
		"{\"level\":\"INFO\",\"message\":\"sep \u2028\u2029 end\"}\n",
		"[2026-09-24T08:30:15.123456] INFO sep \u2028\u2029 end\n",
	},
	"v basic": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"str 7 true false <nil>\"}\n",
		"{\"level\":\"INFO\",\"message\":\"str 7 true false <nil>\"}\n",
		"[2026-09-24T08:30:15.123456] INFO str 7 true false <nil>\n",
	},
	"t bool": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"true false true\"}\n",
		"{\"level\":\"INFO\",\"message\":\"true false true\"}\n",
		"[2026-09-24T08:30:15.123456] INFO true false true\n",
	},
	"d all int types": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"-8 -16 -32 -64 1 8 16 32 64 9223372036854775807 -9223372036854775808\"}\n",
		"{\"level\":\"INFO\",\"message\":\"-8 -16 -32 -64 1 8 16 32 64 9223372036854775807 -9223372036854775808\"}\n",
		"[2026-09-24T08:30:15.123456] INFO -8 -16 -32 -64 1 8 16 32 64 9223372036854775807 -9223372036854775808\n",
	},
	"v all int types": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"-8 -16 -32 -64 1 8 16 32 18446744073709551615 0 0\"}\n",
		"{\"level\":\"INFO\",\"message\":\"-8 -16 -32 -64 1 8 16 32 18446744073709551615 0 0\"}\n",
		"[2026-09-24T08:30:15.123456] INFO -8 -16 -32 -64 1 8 16 32 18446744073709551615 0 0\n",
	},
	"d uintptr": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"7 8\"}\n",
		"{\"level\":\"INFO\",\"message\":\"7 8\"}\n",
		"[2026-09-24T08:30:15.123456] INFO 7 8\n",
	},
	"s bytes": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"bytes [98 50]\"}\n",
		"{\"level\":\"INFO\",\"message\":\"bytes [98 50]\"}\n",
		"[2026-09-24T08:30:15.123456] INFO bytes [98 50]\n",
	},
	"floats": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"3.141590 2.50 1e+21 1.234568e+05 0.1 +Inf NaN\"}\n",
		"{\"level\":\"INFO\",\"message\":\"3.141590 2.50 1e+21 1.234568e+05 0.1 +Inf NaN\"}\n",
		"[2026-09-24T08:30:15.123456] INFO 3.141590 2.50 1e+21 1.234568e+05 0.1 +Inf NaN\n",
	},
	"float verbs": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%!d(float64=1.5) %!s(float64=2.5) %!f(int=3)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%!d(float64=1.5) %!s(float64=2.5) %!f(int=3)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %!d(float64=1.5) %!s(float64=2.5) %!f(int=3)\n",
	},
	"flags width precision": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"   42|42   |00042|+42| 42|ff|FF|10|101|A|\\\"q\\\\\\\"\\\"|U+1F600|0003.142|abc|0xff|\\\"s\\\"\"}\n",
		"{\"level\":\"INFO\",\"message\":\"   42|42   |00042|+42| 42|ff|FF|10|101|A|\\\"q\\\\\\\"\\\"|U+1F600|0003.142|abc|0xff|\\\"s\\\"\"}\n",
		"[2026-09-24T08:30:15.123456] INFO    42|42   |00042|+42| 42|ff|FF|10|101|A|\"q\\\"\"|U+1F600|0003.142|abc|0xff|\"s\"\n",
	},
	"missing arg": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"only-one %!d(MISSING)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"only-one %!d(MISSING)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO only-one %!d(MISSING)\n",
	},
	"missing all args": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%!s(MISSING) %!d(MISSING)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%!s(MISSING) %!d(MISSING)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %!s(MISSING) %!d(MISSING)\n",
	},
	"extra arg": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"a%!(EXTRA string=b)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"a%!(EXTRA string=b)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO a%!(EXTRA string=b)\n",
	},
	"extra args no verbs": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"literal%!(EXTRA int=1, string=two, <nil>)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"literal%!(EXTRA int=1, string=two, <nil>)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO literal%!(EXTRA int=1, string=two, <nil>)\n",
	},
	"bad verb type": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%!d(string=notint)|%!s(int=5)|%!t(int=1)|{}\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%!d(string=notint)|%!s(int=5)|%!t(int=1)|{}\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %!d(string=notint)|%!s(int=5)|%!t(int=1)|{}\n",
	},
	"unknown verb": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%!z(int=1) %!y(string=a)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%!z(int=1) %!y(string=a)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %!z(int=1) %!y(string=a)\n",
	},
	"no verb trailing": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"trailing %!(NOVERB)%!(EXTRA int=1)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"trailing %!(NOVERB)%!(EXTRA int=1)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO trailing %!(NOVERB)%!(EXTRA int=1)\n",
	},
	"no verb trailing no args": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"trailing %!(NOVERB)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"trailing %!(NOVERB)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO trailing %!(NOVERB)\n",
	},
	"non-ascii verb": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%!é(int=1)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%!é(int=1)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %!é(int=1)\n",
	},
	"arg index": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"2 1 2\"}\n",
		"{\"level\":\"INFO\",\"message\":\"2 1 2\"}\n",
		"[2026-09-24T08:30:15.123456] INFO 2 1 2\n",
	},
	"bad arg index": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%!d(BADINDEX)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%!d(BADINDEX)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %!d(BADINDEX)\n",
	},
	"star width": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"   42|7   \"}\n",
		"{\"level\":\"INFO\",\"message\":\"   42|7   \"}\n",
		"[2026-09-24T08:30:15.123456] INFO    42|7   \n",
	},
	"nil s": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"%!s(<nil>) %!d(<nil>)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"%!s(<nil>) %!d(<nil>)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO %!s(<nil>) %!d(<nil>)\n",
	},
	"struct": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"{1 x} {A:1 B:x} struct { A int; B string }{A:1, B:\\\"x\\\"}\"}\n",
		"{\"level\":\"INFO\",\"message\":\"{1 x} {A:1 B:x} struct { A int; B string }{A:1, B:\\\"x\\\"}\"}\n",
		"[2026-09-24T08:30:15.123456] INFO {1 x} {A:1 B:x} struct { A int; B string }{A:1, B:\"x\"}\n",
	},
	"slice map": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"[1 2 3] map[a:1 b:2] [4 5]\"}\n",
		"{\"level\":\"INFO\",\"message\":\"[1 2 3] map[a:1 b:2] [4 5]\"}\n",
		"[2026-09-24T08:30:15.123456] INFO [1 2 3] map[a:1 b:2] [4 5]\n",
	},
	"pointer to struct": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"&{9} &{A:9}\"}\n",
		"{\"level\":\"INFO\",\"message\":\"&{9} &{A:9}\"}\n",
		"[2026-09-24T08:30:15.123456] INFO &{9} &{A:9}\n",
	},
	"error and stringer": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"err! stringer! 4 stringer!\"}\n",
		"{\"level\":\"INFO\",\"message\":\"err! stringer! 4 stringer!\"}\n",
		"[2026-09-24T08:30:15.123456] INFO err! stringer! 4 stringer!\n",
	},
	"wrapped error verb": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"wrapped: %!w(*errors.errorString=&{inner})\"}\n",
		"{\"level\":\"INFO\",\"message\":\"wrapped: %!w(*errors.errorString=&{inner})\"}\n",
		"[2026-09-24T08:30:15.123456] INFO wrapped: %!w(*errors.errorString=&{inner})\n",
	},
	"formatter": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"formatted:v formatted:s formatted:d\"}\n",
		"{\"level\":\"INFO\",\"message\":\"formatted:v formatted:s formatted:d\"}\n",
		"[2026-09-24T08:30:15.123456] INFO formatted:v formatted:s formatted:d\n",
	},
	"named types": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"5 6 ns ns2\"}\n",
		"{\"level\":\"INFO\",\"message\":\"5 6 ns ns2\"}\n",
		"[2026-09-24T08:30:15.123456] INFO 5 6 ns ns2\n",
	},
	"duration and time": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"1.5s 2s 3\"}\n",
		"{\"level\":\"INFO\",\"message\":\"1.5s 2s 3\"}\n",
		"[2026-09-24T08:30:15.123456] INFO 1.5s 2s 3\n",
	},
	"rune and byte": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"x 121 122 %!s(int32=119)\"}\n",
		"{\"level\":\"INFO\",\"message\":\"x 121 122 %!s(int32=119)\"}\n",
		"[2026-09-24T08:30:15.123456] INFO x 121 122 %!s(int32=119)\n",
	},
	"type verb": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"int string <nil>\"}\n",
		"{\"level\":\"INFO\",\"message\":\"int string <nil>\"}\n",
		"[2026-09-24T08:30:15.123456] INFO int string <nil>\n",
	},
	"long hostile": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"prefix a\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\c END\"}\n",
		"{\"level\":\"INFO\",\"message\":\"prefix a\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\ca\\\"b\\\\c END\"}\n",
		"[2026-09-24T08:30:15.123456] INFO prefix a\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\ca\"b\\c END\n",
	},
	"long clean": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\"prefix abcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefgh 1 END\"}\n",
		"{\"level\":\"INFO\",\"message\":\"prefix abcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefgh 1 END\"}\n",
		"[2026-09-24T08:30:15.123456] INFO prefix abcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefghabcdefgh 1 END\n",
	},
	"spaces": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"message\":\" 5 \"}\n",
		"{\"level\":\"INFO\",\"message\":\" 5 \"}\n",
		"[2026-09-24T08:30:15.123456] INFO  5 \n",
	},
	"with fields hostile": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"INFO\",\"k\":\"v\\\"q\",\"n\":-1,\"b\":true,\"message\":\"x\\ty=2\"}\n",
		"{\"level\":\"INFO\",\"k\":\"v\\\"q\",\"n\":-1,\"b\":true,\"message\":\"x\\ty=2\"}\n",
		"[2026-09-24T08:30:15.123456] INFO k=v\"q n=-1 b=true x\ty=2\n",
	},
	"warn level": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"WARN\",\"message\":\"warn 1\"}\n",
		"{\"level\":\"WARN\",\"message\":\"warn 1\"}\n",
		"[2026-09-24T08:30:15.123456] WARN warn 1\n",
	},
	"error level": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"ERROR\",\"message\":\"error e\"}\n",
		"{\"level\":\"ERROR\",\"message\":\"error e\"}\n",
		"[2026-09-24T08:30:15.123456] ERRR error e\n",
	},
	"below level": {
		"",
		"",
		"",
	},
	"panic simple": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"PANIC\",\"message\":\"boom 42\"}\n",
		"{\"level\":\"PANIC\",\"message\":\"boom 42\"}\n",
		"[2026-09-24T08:30:15.123456] PANC boom 42\n",
	},
	"panic complex": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"PANIC\",\"message\":\"boom    42 x %!d(string=y)\"}\n",
		"{\"level\":\"PANIC\",\"message\":\"boom    42 x %!d(string=y)\"}\n",
		"[2026-09-24T08:30:15.123456] PANC boom    42 x %!d(string=y)\n",
	},
	"panic no args": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"PANIC\",\"message\":\"boom\"}\n",
		"{\"level\":\"PANIC\",\"message\":\"boom\"}\n",
		"[2026-09-24T08:30:15.123456] PANC boom\n",
	},
	"fatal": {
		"{\"time\":\"2026-09-24T08:30:15.123456Z\",\"level\":\"FATAL\",\"message\":\"fatal bye\"}\n",
		"{\"level\":\"FATAL\",\"message\":\"fatal bye\"}\n",
		"[2026-09-24T08:30:15.123456] FATL fatal bye\n",
	},
}
