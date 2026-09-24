package iqlog

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

// The timestamp formats are a public contract. These tests pin the exact
// bytes every timestamp path emits for the payload the comparative
// benchmarks use, over a clock sequence that exercises the per-second prefix
// cache (same second, next second, day and year rollover), the microsecond
// digits at their edges (0, 999999, truncated nanoseconds), a non-UTC input
// zone, a leap day, and the years the fixed-width encoders cannot represent.
// Any change to the timestamp path must leave every string here untouched.

var goldenClock = []time.Time{
	time.Date(2026, time.August, 29, 14, 3, 12, 481947000, time.UTC),
	time.Date(2026, time.August, 29, 14, 3, 12, 481947999, time.UTC), // same second: cache hit, nanoseconds truncated
	time.Date(2026, time.August, 29, 14, 3, 12, 0, time.UTC),
	time.Date(2026, time.August, 29, 14, 3, 12, 999999999, time.UTC),
	time.Date(2026, time.August, 29, 14, 3, 13, 1000, time.UTC), // next second: cache miss
	time.Date(2026, time.December, 31, 23, 59, 59, 999999000, time.UTC),
	time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC), // year rollover
	time.Date(1969, time.December, 31, 23, 59, 59, 123456000, time.UTC),
	time.Date(2026, time.August, 29, 14, 3, 12, 481947000, time.FixedZone("plus", 5*3600+30*60)),
	time.Date(2000, time.February, 29, 1, 2, 3, 40506000, time.UTC),
	time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC),
	time.Date(10000, time.January, 1, 0, 0, 0, 7000, time.UTC), // five-digit year: slow path
	time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
	time.Date(-1, time.June, 7, 8, 9, 10, 123456000, time.UTC), // negative year: slow path
}

// goldenSequence returns a Config.Now that walks goldenClock once per call.
func goldenSequence() func() time.Time {
	i := 0
	return func() time.Time {
		now := goldenClock[i%len(goldenClock)]
		i++
		return now
	}
}

func goldenTimestampPayload(l *Logger) {
	for range goldenClock {
		l.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(goldenMsg)
	}
}

func goldenEventAtPayload(l *Logger) {
	for range goldenClock {
		l.EventAt(LevelInfo, "main.run", "main.go:12").Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(goldenMsg)
	}
}

const goldenJSONTail = `,"level":"INFO","rate":"15","low":16,"high":123.2,"message":"The quick brown fox jumps over the lazy dog"}` + "\n"
const goldenJSONCallerTail = `,"level":"INFO","func":"main.run","file":"main.go:12","rate":"15","low":16,"high":123.2,"message":"The quick brown fox jumps over the lazy dog"}` + "\n"
const goldenConsoleTail = " INFO rate=15 low=16 high=123.2 The quick brown fox jumps over the lazy dog\n"

var goldenJSONUTCTimes = []string{
	"2026-08-29T14:03:12.481947Z",
	"2026-08-29T14:03:12.481947Z",
	"2026-08-29T14:03:12.000000Z",
	"2026-08-29T14:03:12.999999Z",
	"2026-08-29T14:03:13.000001Z",
	"2026-12-31T23:59:59.999999Z",
	"2027-01-01T00:00:00.000000Z",
	"1969-12-31T23:59:59.123456Z",
	"2026-08-29T08:33:12.481947Z",
	"2000-02-29T01:02:03.040506Z",
	"9999-12-31T23:59:59.999999Z",
	"10000-01-01T00:00:00.000007Z",
	"0000-01-01T00:00:00.000000Z",
	"-0001-06-07T08:09:10.123456Z",
}

func goldenLines(prefix string, times []string, tail string) string {
	var sb strings.Builder
	for _, t := range times {
		sb.WriteString(prefix)
		sb.WriteString(t)
		sb.WriteString(tail)
	}
	return sb.String()
}

func runGolden(t *testing.T, name string, cfg Config, emit func(*Logger), want string) {
	t.Helper()
	sb := &syncBuffer{}
	cfg.Writer = sb
	cfg.Now = goldenSequence()
	l := MustNew(cfg)
	emit(l)
	if got := sb.String(); got != want {
		t.Errorf("%s:\n got %q\nwant %q", name, got, want)
	}
}

func TestGoldenJSONUTCTimestampBenchmarkPayload(t *testing.T) {
	runGolden(t, "json utc", Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC}, goldenTimestampPayload,
		goldenLines(`{"time":"`, goldenJSONUTCTimes, `"`+goldenJSONTail))
}

func TestGoldenJSONUTCTimestampLevels(t *testing.T) {
	runGolden(t, "json utc levels", Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC, Level: LevelTrace}, func(l *Logger) {
		l.TraceEvent().Msg("t")
		l.DebugEvent().Msg("d")
		l.InfoEvent().Msg("i")
		l.WarnEvent().Msg("w")
		l.ErrorEvent().Msg("e")
	}, `{"time":"2026-08-29T14:03:12.481947Z","level":"TRACE","message":"t"}`+"\n"+
		`{"time":"2026-08-29T14:03:12.481947Z","level":"DEBUG","message":"d"}`+"\n"+
		`{"time":"2026-08-29T14:03:12.000000Z","level":"INFO","message":"i"}`+"\n"+
		`{"time":"2026-08-29T14:03:12.999999Z","level":"WARN","message":"w"}`+"\n"+
		`{"time":"2026-08-29T14:03:13.000001Z","level":"ERROR","message":"e"}`+"\n")
}

func TestGoldenJSONUTCTimestampWithCaller(t *testing.T) {
	// the explicit-caller path writes the caller scratch, which must leave
	// the timestamp cache alone, so it must produce the same timestamps
	runGolden(t, "json utc caller", Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC, CallerDepth: 1}, goldenEventAtPayload,
		goldenLines(`{"time":"`, goldenJSONUTCTimes, `"`+goldenJSONCallerTail))
	// alternating between the cached path and the caller path on one logger
	runGolden(t, "json utc mixed caller", Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC}, func(l *Logger) {
		l.InfoEvent().Msg("a")
		l.EventAt(LevelInfo, "main.run", "main.go:12").Msg("b")
		l.InfoEvent().Msg("c")
	}, `{"time":"2026-08-29T14:03:12.481947Z","level":"INFO","message":"a"}`+"\n"+
		`{"time":"2026-08-29T14:03:12.481947Z","level":"INFO","func":"main.run","file":"main.go:12","message":"b"}`+"\n"+
		`{"time":"2026-08-29T14:03:12.000000Z","level":"INFO","message":"c"}`+"\n")
}

func TestGoldenJSONCustomTimestampLayouts(t *testing.T) {
	runGolden(t, "json custom layout", Config{Format: FormatJSON, JSONTimeMode: JSONTimeCustom, TimestampLayout: "2006/01/02 15:04:05.000 -07:00"}, goldenTimestampPayload,
		goldenLines(`{"time":"`, []string{
			"2026/08/29 14:03:12.481 +00:00",
			"2026/08/29 14:03:12.481 +00:00",
			"2026/08/29 14:03:12.000 +00:00",
			"2026/08/29 14:03:12.999 +00:00",
			"2026/08/29 14:03:13.000 +00:00",
			"2026/12/31 23:59:59.999 +00:00",
			"2027/01/01 00:00:00.000 +00:00",
			"1969/12/31 23:59:59.123 +00:00",
			"2026/08/29 14:03:12.481 +05:30",
			"2000/02/29 01:02:03.040 +00:00",
			"9999/12/31 23:59:59.999 +00:00",
			"10000/01/01 00:00:00.000 +00:00",
			"0000/01/01 00:00:00.000 +00:00",
			"-0001/06/07 08:09:10.123 +00:00",
		}, `"`+goldenJSONTail))
	runGolden(t, "json custom nanoseconds", Config{Format: FormatJSON, JSONTimeMode: JSONTimeCustom, TimestampLayout: "nanoseconds"}, goldenTimestampPayload,
		goldenLines(`{"time":"`, []string{
			"2026-08-29T14:03:12.481947Z",
			"2026-08-29T14:03:12.481947999Z",
			"2026-08-29T14:03:12Z",
			"2026-08-29T14:03:12.999999999Z",
			"2026-08-29T14:03:13.000001Z",
			"2026-12-31T23:59:59.999999Z",
			"2027-01-01T00:00:00Z",
			"1969-12-31T23:59:59.123456Z",
			"2026-08-29T14:03:12.481947+05:30",
			"2000-02-29T01:02:03.040506Z",
			"9999-12-31T23:59:59.999999999Z",
			"10000-01-01T00:00:00.000007Z",
			"0000-01-01T00:00:00Z",
			"-0001-06-07T08:09:10.123456Z",
		}, `"`+goldenJSONTail))
}

func TestGoldenJSONTimestampDisabled(t *testing.T) {
	runGolden(t, "json disabled", Config{Format: FormatJSON, JSONTimeMode: JSONTimeDisabled}, goldenTimestampPayload,
		strings.Repeat(`{"level":"INFO","rate":"15","low":16,"high":123.2,"message":"The quick brown fox jumps over the lazy dog"}`+"\n", len(goldenClock)))
}

func TestGoldenConsoleTimestamps(t *testing.T) {
	// the console default keeps the clock's own zone
	runGolden(t, "console default", Config{IncludeTime: true, DisableColor: true}, goldenTimestampPayload,
		goldenLines("[", []string{
			"2026-08-29T14:03:12.481947]",
			"2026-08-29T14:03:12.481947]",
			"2026-08-29T14:03:12.000000]",
			"2026-08-29T14:03:12.999999]",
			"2026-08-29T14:03:13.000001]",
			"2026-12-31T23:59:59.999999]",
			"2027-01-01T00:00:00.000000]",
			"1969-12-31T23:59:59.123456]",
			"2026-08-29T14:03:12.481947]",
			"2000-02-29T01:02:03.040506]",
			"9999-12-31T23:59:59.999999]",
			"10000-01-01T00:00:00.000007]",
			"0000-01-01T00:00:00.000000]",
			"-0001-06-07T08:09:10.123456]",
		}, goldenConsoleTail))
	runGolden(t, "console custom layout", Config{IncludeTime: true, DisableColor: true, TimestampLayout: time.RFC1123}, goldenTimestampPayload,
		goldenLines("[", []string{
			"Sat, 29 Aug 2026 14:03:12 UTC]",
			"Sat, 29 Aug 2026 14:03:12 UTC]",
			"Sat, 29 Aug 2026 14:03:12 UTC]",
			"Sat, 29 Aug 2026 14:03:12 UTC]",
			"Sat, 29 Aug 2026 14:03:13 UTC]",
			"Thu, 31 Dec 2026 23:59:59 UTC]",
			"Fri, 01 Jan 2027 00:00:00 UTC]",
			"Wed, 31 Dec 1969 23:59:59 UTC]",
			"Sat, 29 Aug 2026 14:03:12 plus]",
			"Tue, 29 Feb 2000 01:02:03 UTC]",
			"Fri, 31 Dec 9999 23:59:59 UTC]",
			"Sat, 01 Jan 10000 00:00:00 UTC]",
			"Sat, 01 Jan 0000 00:00:00 UTC]",
			"Mon, 07 Jun -0001 08:09:10 UTC]",
		}, goldenConsoleTail))
	runGolden(t, "console color", Config{IncludeTime: true, Color: true}, func(l *Logger) {
		l.InfoEvent().Str("rate", "15").Msg(goldenMsg)
	}, "[2026-08-29T14:03:12.481947] \x1b[34mINFO\x1b[0m rate=15 The quick brown fox jumps over the lazy dog\n")
	runGolden(t, "console caller", Config{IncludeTime: true, DisableColor: true, CallerDepth: 1}, func(l *Logger) {
		l.EventAt(LevelWarn, "main.run", "main.go:12").Str("rate", "15").Msg(goldenMsg)
	}, "[2026-08-29T14:03:12.481947] WARN [main.run main.go:12] rate=15 The quick brown fox jumps over the lazy dog\n")
	runGolden(t, "console no time", Config{DisableColor: true}, func(l *Logger) {
		l.InfoEvent().Str("rate", "15").Msg(goldenMsg)
	}, "INFO rate=15 The quick brown fox jumps over the lazy dog\n")
}

func TestGoldenLegacyInfoTimestamp(t *testing.T) {
	runGolden(t, "legacy info", Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC}, func(l *Logger) {
		l.Info("m", 1, "x")
	}, `{"time":"2026-08-29T14:03:12.481947Z","level":"INFO","message":"m 1 x"}`+"\n")
}

// The default clock has no fixed sequence, so the contract it can be held to
// is the shape of the field and that the value is exactly the wall clock:
// parseable as the fixed-width UTC layout, never earlier than a time.Now
// taken before the record (at the microsecond the field carries) and never
// later than one taken after it. TestDefaultClockStampsAreExact holds both
// fixed-width encoders to the same window over many more records.
var goldenDefaultClockPattern = regexp.MustCompile(`^\{"time":"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z)","level":"INFO","rate":"15","low":16,"high":123.2,"message":"The quick brown fox jumps over the lazy dog"\}\n$`)

func TestGoldenJSONUTCTimestampDefaultClock(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC, Writer: sb})
	// spread the records over several milliseconds so the clock is sampled
	// across more than one instant
	for i := 0; i < 20; i++ {
		sb.Reset()
		before := time.Now()
		l.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(goldenMsg)
		after := time.Now()
		m := goldenDefaultClockPattern.FindStringSubmatch(sb.String())
		if m == nil {
			t.Fatalf("record %d: unexpected shape %q", i, sb.String())
		}
		stamp, err := time.Parse("2006-01-02T15:04:05.000000Z07:00", m[1])
		if err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
		if stamp.Before(before.Truncate(time.Microsecond)) || stamp.After(after) {
			t.Fatalf("record %d: time %s outside [%s, %s]", i, stamp.Format(time.RFC3339Nano), before.Format(time.RFC3339Nano), after.Format(time.RFC3339Nano))
		}
		time.Sleep(200 * time.Microsecond)
	}
}
