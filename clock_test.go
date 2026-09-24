package iqlog

import (
	"io"
	"regexp"
	"strings"
	"testing"
	"time"
)

func withClock(cfg Config, now func() time.Time) Config {
	cfg.Now = now
	return cfg
}

// The fast clock is taken only for the two fixed-width encoders and only
// when the clock is time.Now itself, as configured or as reported back by
// Config(); every other layout or clock calls Config.Now for every record.
func TestFastClockDetection(t *testing.T) {
	closure := func() time.Time { return time.Now() }
	fixed := func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }
	jsonUTC := Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC, Writer: io.Discard}
	console := Config{IncludeTime: true, DisableColor: true, Writer: io.Discard}
	for name, tc := range map[string]struct {
		cfg  Config
		want bool
	}{
		"json utc, unset":                {jsonUTC, true},
		"json utc, time.Now itself":      {withClock(jsonUTC, time.Now), true},
		"json utc, closure around Now":   {withClock(jsonUTC, closure), false},
		"json utc, fixed clock":          {withClock(jsonUTC, fixed), false},
		"json utc, round trip of unset":  {withClock(jsonUTC, MustNew(jsonUTC).Config().Now), true},
		"json utc, round trip of custom": {withClock(jsonUTC, MustNew(withClock(jsonUTC, closure)).Config().Now), false},
		"json custom layout":             {Config{Format: FormatJSON, JSONTimeMode: JSONTimeCustom, TimestampLayout: "microseconds", Writer: io.Discard}, false},
		"json time disabled":             {Config{Format: FormatJSON, Writer: io.Discard}, false},
		"console default layout":         {console, true},
		"console default, closure":       {withClock(console, closure), false},
		"console custom layout":          {Config{IncludeTime: true, TimestampLayout: time.RFC3339Nano, Writer: io.Discard}, false},
		"console named layout":           {Config{IncludeTime: true, TimestampLayout: "microseconds", Writer: io.Discard}, false},
		"console without time":           {Config{Writer: io.Discard}, false},
	} {
		l := MustNew(tc.cfg)
		if got := l.config.Load().fastClock; got != tc.want {
			t.Errorf("%s: fastClock = %v, want %v", name, got, tc.want)
		}
	}
}

// SetJSONTimeMode rewrites the format and time mode of the live snapshot, so
// the fast-clock decision has to follow the change rather than stay as it
// was when the logger was built; the other setters go through the same
// update path and must leave it alone.
func TestFastClockFollowsConfigUpdates(t *testing.T) {
	l := MustNew(Config{IncludeTime: true, TimestampLayout: time.RFC3339Nano, Writer: io.Discard})
	if l.config.Load().fastClock {
		t.Fatal("console with a custom layout: fastClock set")
	}
	for i, step := range []struct {
		mode JSONTimeMode
		want bool
	}{
		{JSONTimeUTC, true},
		{JSONTimeDisabled, false},
		{JSONTimeCustom, false},
		{JSONTimeUTC, true},
	} {
		l.SetJSONTimeMode(step.mode)
		if got := l.config.Load().fastClock; got != step.want {
			t.Fatalf("step %d: after SetJSONTimeMode(%d) fastClock = %v, want %v", i, step.mode, got, step.want)
		}
	}
	l.SetLevel(LevelDebug)
	l.SetCallerDepth(1)
	l.SetUseColor(true)
	if !l.config.Load().fastClock {
		t.Fatal("an unrelated setter cleared fastClock")
	}
	custom := MustNew(withClock(Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC, Writer: io.Discard}, func() time.Time { return time.Now() }))
	custom.SetJSONTimeMode(JSONTimeUTC)
	if custom.config.Load().fastClock {
		t.Fatal("a caller's clock was replaced by a time mode change")
	}
}

// wallClock must be exactly CLOCK_REALTIME truncated to the microsecond:
// never before a time.Now taken just before it and never after one taken
// just after it. There is no slack here; a derived or cached reading that
// could land a microsecond early fails this.
func TestWallClockIsExact(t *testing.T) {
	for i := 0; i < 100000; i++ {
		before := time.Now()
		sec, usec := wallClock()
		after := time.Now()
		if usec < 0 || usec >= 1e6 {
			t.Fatalf("sample %d: microseconds %d out of range", i, usec)
		}
		got := time.Unix(sec, usec*1000)
		if got.Before(before.Truncate(time.Microsecond)) || got.After(after) {
			t.Fatalf("sample %d: wallClock %s outside [%s, %s]", i, got.Format(time.RFC3339Nano), before.Format(time.RFC3339Nano), after.Format(time.RFC3339Nano))
		}
	}
}

var (
	exactJSONRecord    = regexp.MustCompile(`^\{"time":"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z)","level":"INFO","message":"m"\}\n$`)
	exactConsoleRecord = regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6})\] INFO m\n$`)
)

// Every record stamped by the default clock carries the wall clock at the
// instant it was read, truncated to the microsecond the field holds: never
// earlier than a time.Now taken before the record (at that resolution) and
// never later than one taken after it. Both fixed-width encoders take the
// fast clock and both are held to the same window.
func TestDefaultClockStampsAreExact(t *testing.T) {
	const records = 5000
	for _, tc := range []struct {
		name    string
		cfg     Config
		pattern *regexp.Regexp
		// parse turns the stamped text back into an instant. ref is a
		// time.Now taken just before the record: the console layout carries
		// no zone, and resolving it through ref's offset rather than
		// time.Local keeps the hour around a DST transition unambiguous.
		parse func(s string, ref time.Time) (time.Time, error)
	}{
		{"json utc", Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC}, exactJSONRecord,
			func(s string, _ time.Time) (time.Time, error) {
				return time.Parse("2006-01-02T15:04:05.000000Z07:00", s)
			}},
		{"console default", Config{IncludeTime: true, DisableColor: true}, exactConsoleRecord,
			func(s string, ref time.Time) (time.Time, error) {
				_, offset := ref.Zone()
				return time.ParseInLocation(defaultConsoleTimestampLayout, s, time.FixedZone("", offset))
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sb := &syncBuffer{}
			cfg := tc.cfg
			cfg.Writer = sb
			l := MustNew(cfg)
			if !l.config.Load().fastClock {
				t.Fatal("configuration did not select the fast clock")
			}
			for i := 0; i < records; i++ {
				sb.Reset()
				before := time.Now()
				l.InfoEvent().Msg("m")
				after := time.Now()
				m := tc.pattern.FindStringSubmatch(sb.String())
				if m == nil {
					t.Fatalf("record %d: unexpected shape %q", i, sb.String())
				}
				stamp, err := tc.parse(m[1], before)
				if err != nil {
					t.Fatalf("record %d: %v", i, err)
				}
				if stamp.Before(before.Truncate(time.Microsecond)) || stamp.After(after) {
					t.Fatalf("record %d: stamped %s outside [%s, %s]", i, stamp.Format(time.RFC3339Nano), before.Format(time.RFC3339Nano), after.Format(time.RFC3339Nano))
				}
			}
		})
	}
}

// A custom layout keeps calling the configured clock even when that clock is
// time.Now, so a nanosecond layout is never quietly truncated to the
// microsecond the fast clock reads.
func TestCustomLayoutKeepsNanosecondsWithDefaultClock(t *testing.T) {
	fine := false
	for i := 0; i < 1000 && !fine; i++ {
		fine = time.Now().Nanosecond()%1000 != 0
	}
	if !fine {
		t.Skip("time.Now resolves to whole microseconds on this host")
	}
	fraction := regexp.MustCompile(`\.(\d+)`)
	for name, cfg := range map[string]Config{
		"json":    {Format: FormatJSON, JSONTimeMode: JSONTimeCustom, TimestampLayout: "nanoseconds"},
		"console": {IncludeTime: true, DisableColor: true, TimestampLayout: "nanoseconds"},
	} {
		sb := &syncBuffer{}
		cfg.Writer = sb
		l := MustNew(cfg)
		if l.config.Load().fastClock {
			t.Fatalf("%s: a nanosecond layout selected the fast clock", name)
		}
		seen := false
		for i := 0; i < 1000 && !seen; i++ {
			sb.Reset()
			l.InfoEvent().Msg("m")
			// RFC3339Nano drops trailing zeros, so more than six fractional
			// digits means sub-microsecond digits survived
			m := fraction.FindStringSubmatch(sb.String())
			seen = m != nil && len(m[1]) > 6
		}
		if !seen {
			t.Errorf("%s: 1000 records with the nanoseconds layout never carried sub-microsecond digits", name)
		}
	}
}

// Every timestamped default path has to stay allocation-free, the clock read
// included.
func TestFastClockPathsDoNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	jsonUTC := MustNew(Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC, Writer: io.Discard})
	jsonCaller := MustNew(Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC, CallerDepth: 1, Writer: io.Discard})
	jsonCustom := MustNew(Config{Format: FormatJSON, JSONTimeMode: JSONTimeCustom, TimestampLayout: "2006-01-02 15:04:05", Writer: io.Discard})
	console := MustNew(Config{IncludeTime: true, DisableColor: true, Writer: io.Discard})
	consoleCustom := MustNew(Config{IncludeTime: true, DisableColor: true, TimestampLayout: time.RFC1123, Writer: io.Discard})
	// the caller stack scan is not covered here: its symbolization is a
	// separate cost that is not part of the timestamp contract
	for name, fn := range map[string]func(){
		"wall clock":     func() { wallClock() },
		"json utc":       func() { jsonUTC.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg("message") },
		"json event at":  func() { jsonCaller.EventAt(LevelInfo, "main.run", "main.go:12").Str("rate", "15").Msg("message") },
		"json custom":    func() { jsonCustom.InfoEvent().Str("rate", "15").Msg("message") },
		"console":        func() { console.InfoEvent().Str("rate", "15").Msg("message") },
		"console custom": func() { consoleCustom.InfoEvent().Str("rate", "15").Msg("message") },
	} {
		if got := testing.AllocsPerRun(1000, fn); got != 0 {
			t.Errorf("%s allocated %.2f times per record", name, got)
		}
	}
}

// The cached prefix must be rebuilt for a new second even when the caller
// scratch is in use, and must survive a record whose year it cannot hold.
func TestJSONTimePrefixCacheAcrossCallerAndUnrepresentableYears(t *testing.T) {
	times := []time.Time{
		time.Date(2026, 8, 29, 14, 3, 12, 481947000, time.UTC),
		time.Date(12345, 6, 7, 8, 9, 10, 123456000, time.UTC),
		time.Date(2026, 8, 29, 14, 3, 12, 5000, time.UTC),
		time.Date(2026, 8, 29, 14, 3, 13, 0, time.UTC),
	}
	i := 0
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, JSONTimeMode: JSONTimeUTC, Writer: sb, Now: func() time.Time { n := times[i]; i++; return n }})
	l.InfoEvent().Msg("a")
	l.EventAt(LevelInfo, "main.run", "main.go:12").Msg("b")
	l.InfoEvent().Msg("c")
	l.EventAt(LevelInfo, "main.run", "main.go:12").Msg("d")
	want := strings.Join([]string{
		`{"time":"2026-08-29T14:03:12.481947Z","level":"INFO","message":"a"}`,
		`{"time":"12345-06-07T08:09:10.123456Z","level":"INFO","func":"main.run","file":"main.go:12","message":"b"}`,
		`{"time":"2026-08-29T14:03:12.000005Z","level":"INFO","message":"c"}`,
		`{"time":"2026-08-29T14:03:13.000000Z","level":"INFO","func":"main.run","file":"main.go:12","message":"d"}`,
	}, "\n") + "\n"
	if got := sb.String(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

// The encoder must produce the same bytes from a cold cache, a warm cache,
// and a buffer too small to hold the timestamp, through both the nanosecond
// entry a time.Time takes and the microsecond entry the fast clock takes.
func TestDefaultJSONTimestampEncoderMatchesLayout(t *testing.T) {
	e := &Event{timeSecond: invalidTimestampSecond}
	for _, tm := range append(append([]time.Time{}, goldenClock...),
		time.Date(2026, 8, 29, 14, 3, 12, 999, time.UTC),
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1969, 12, 31, 23, 59, 59, 999999999, time.UTC),
	) {
		want := `{"time":"` + tm.UTC().Format("2006-01-02T15:04:05.000000Z07:00") + `",`
		for _, capacity := range []int{0, 10, 64} {
			e.output = make([]byte, 0, capacity)
			e.writeDefaultJSONTimestamp(tm.Unix(), int64(tm.Nanosecond()))
			if got := string(e.output); got != want {
				t.Errorf("%s (cap %d, nanoseconds): got %q, want %q", tm, capacity, got, want)
			}
			e.output = make([]byte, 0, capacity)
			e.writeJSONTimestampMicros(tm.Unix(), int64(tm.Nanosecond()/1000))
			if got := string(e.output); got != want {
				t.Errorf("%s (cap %d, microseconds): got %q, want %q", tm, capacity, got, want)
			}
		}
	}
}

func BenchmarkWallClock(b *testing.B) {
	var sink int64
	for i := 0; i < b.N; i++ {
		sec, usec := wallClock()
		sink += sec + usec
	}
	_ = sink
}

func BenchmarkTimeNow(b *testing.B) {
	var sink int64
	for i := 0; i < b.N; i++ {
		now := time.Now()
		sink += now.Unix() + int64(now.Nanosecond())
	}
	_ = sink
}

func BenchmarkDefaultJSONTimestamp(b *testing.B) {
	e := &Event{timeSecond: invalidTimestampSecond, output: make([]byte, 0, 512)}
	sec, usec := int64(1787680992), int64(481947)
	for i := 0; i < b.N; i++ {
		e.output = e.output[:0]
		e.writeJSONTimestampMicros(sec, usec+int64(i&1023))
	}
}
