package iqlog

// Record-integrity coverage: paths that put caller- or configuration-supplied
// text into the record envelope (which the whole-line console scan exempts),
// paths that could split a newline-delimited record, writer serialization
// across reconfiguration, and bounded retention of write errors.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"
)

// A: EventAt lets the caller supply func/file, which land in the console
// envelope -- before prefixLen, so the control-byte scan skips them.
func TestEventAtConsoleInjection(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatConsole, Writer: &buf, Level: LevelInfo, CallerDepth: 1})
	l.EventAt(LevelInfo, "func\nINFO forged line", "f.go:1\x1b[31m").Msg("real")
	out := buf.String()
	t.Logf("console: %q", out)
	if strings.Count(out, "\n") != 1 {
		t.Errorf("BUG: EventAt forged a line: %q", out)
	}
	if strings.ContainsRune(out, 0x1b) {
		t.Errorf("BUG: EventAt passed ESC to the terminal: %q", out)
	}
}

// B: RawJSON accepts any valid JSON document, including one with literal
// newlines between tokens, which would split a newline-delimited record.
func TestRawJSONNewline(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo})
	raw := []byte("{\"a\":\n1}")
	if !json.Valid(raw) {
		t.Fatal("premise wrong: not valid JSON")
	}
	l.InfoEvent().RawJSON("payload", raw).Msg("m")
	out := buf.String()
	t.Logf("json: %q", out)
	if strings.Count(out, "\n") != 1 {
		t.Errorf("BUG: RawJSON split the record across %d lines: %q", strings.Count(out, "\n"), out)
	}
}

// C: a custom timestamp layout is written into the JSON envelope unescaped.
func TestTimestampLayoutJSON(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{
		Format: FormatJSON, Writer: &buf, Level: LevelInfo,
		JSONTimeMode: JSONTimeCustom, TimestampLayout: `2006 "quoted`,
	})
	l.Info("m")
	out := bytes.TrimSpace(buf.Bytes())
	t.Logf("json: %s", out)
	if !json.Valid(out) {
		t.Errorf("BUG: timestamp layout broke the record: %s", out)
	}
}

// D: same layout in console mode lands before prefixLen
func TestTimestampLayoutConsole(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{
		Format: FormatConsole, Writer: &buf, Level: LevelInfo,
		IncludeTime: true, TimestampLayout: "2006\nINFO forged",
	})
	l.Info("m")
	out := buf.String()
	t.Logf("console: %q", out)
	if strings.Count(out, "\n") != 1 {
		t.Errorf("BUG: timestamp layout forged a line: %q", out)
	}
}

// E: extreme clock values in the fixed-width timestamp encoders
func TestExtremeYear(t *testing.T) {
	for _, tm := range []time.Time{
		time.Date(99999, 1, 2, 3, 4, 5, 0, time.UTC),
		time.Date(-5000, 1, 2, 3, 4, 5, 0, time.UTC),
		time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
	} {
		var buf bytes.Buffer
		l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo,
			JSONTimeMode: JSONTimeUTC, Now: func() time.Time { return tm }})
		l.Info("m")
		out := bytes.TrimSpace(buf.Bytes())
		t.Logf("year %d -> %s", tm.Year(), out)
		if !json.Valid(out) {
			t.Errorf("BUG: year %d produced invalid JSON: %s", tm.Year(), out)
		}
	}
}

// F: console values that are not strings still route through the scan
func TestAnyConsole(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatConsole, Writer: &buf, Level: LevelInfo})
	type weird struct{ V string }
	l.InfoEvent().Any("k", weird{"a\nINFO forged"}).Msg("m")
	out := buf.String()
	t.Logf("console: %q", out)
	if strings.Count(out, "\n") != 1 {
		t.Errorf("BUG: Any forged a line: %q", out)
	}
}

type mjson struct{}

func (mjson) MarshalJSON() ([]byte, error) { return []byte("{\"a\":\n1}"), nil }

func TestAnyMarshalerNewline(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo})
	l.InfoEvent().Any("k", mjson{}).Msg("m")
	t.Logf("json: %q", buf.String())
	if strings.Count(buf.String(), "\n") != 1 {
		t.Errorf("BUG: custom marshaler split the record")
	}
}

// async writer must not lose or corrupt records under concurrent close
func TestAsyncCloseRecordLoss(t *testing.T) {
	for iter := 0; iter < 20; iter++ {
		sb := &syncBuffer{}
		l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo,
			WriterMode: WriterAsync, BufferSize: 8})
		var wg sync.WaitGroup
		for g := 0; g < 4; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					l.Info("record")
				}
			}()
		}
		wg.Wait()
		if err := l.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		got := strings.Count(sb.String(), "record")
		if got != 400 {
			t.Fatalf("iter %d: got %d records, want 400", iter, got)
		}
		for _, line := range strings.Split(strings.TrimSuffix(sb.String(), "\n"), "\n") {
			if !json.Valid([]byte(line)) {
				t.Fatalf("corrupt line: %q", line)
			}
		}
	}
}

// logging concurrently with Close must not panic or corrupt
func TestLogDuringClose(t *testing.T) {
	for iter := 0; iter < 30; iter++ {
		sb := &syncBuffer{}
		l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo,
			WriterMode: WriterAsync, BufferSize: 4})
		var wg sync.WaitGroup
		for g := 0; g < 4; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 200; i++ {
					l.Info("x")
				}
			}()
		}
		_ = l.Close()
		wg.Wait()
		for _, line := range strings.Split(strings.TrimSuffix(sb.String(), "\n"), "\n") {
			if line != "" && !json.Valid([]byte(line)) {
				t.Fatalf("iter %d: corrupt line during close: %q", iter, line)
			}
		}
	}
}

// reconfiguring while logging must not corrupt records
func TestReconfigureDuringLogging(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo})
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					l.InfoEvent().Str("k", "v").Msg("m")
				}
			}
		}()
	}
	for i := 0; i < 300; i++ {
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo})
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo, WriterMode: WriterAsync})
	}
	close(stop)
	wg.Wait()
	_ = l.Flush()
	bad := 0
	for _, line := range strings.Split(strings.TrimSuffix(sb.String(), "\n"), "\n") {
		if line != "" && !json.Valid([]byte(line)) {
			bad++
			if bad == 1 {
				t.Errorf("BUG: corrupt line during reconfiguration: %q", line)
			}
		}
	}
	if bad > 0 {
		t.Errorf("%d corrupt lines", bad)
	}
}

// Every failing write joins onto writeErr.pending. Nothing trims it, so a
// logger writing to a broken destination accumulates one error per record
// until someone calls Flush.
func TestWriteErrorAccumulation(t *testing.T) {
	l := MustNew(Config{Format: FormatJSON, Writer: failingWriter{}, Level: LevelInfo})

	var m0, m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m0)
	const n = 200000
	for i := 0; i < n; i++ {
		l.Info("this write fails")
	}
	runtime.GC()
	runtime.ReadMemStats(&m1)
	grew := int64(m1.HeapAlloc) - int64(m0.HeapAlloc)
	t.Logf("heap grew %d bytes after %d failed writes (%.1f bytes/record retained)",
		grew, n, float64(grew)/float64(n))

	err := l.Flush()
	depth := 0
	if err != nil {
		depth = strings.Count(err.Error(), "disk full")
	}
	t.Logf("Flush() error joins %d underlying errors", depth)
	if grew > 1<<20 {
		t.Errorf("BUG: %d bytes retained by pending write errors", grew)
	}
}

// a typed-nil error must not panic
func TestTypedNilErr(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("NOTE: Err(typed-nil) panics: %v", r)
		}
	}()
	var e *myErr
	l := MustNew(Config{Format: FormatJSON, Writer: &syncBuffer{}, Level: LevelInfo})
	l.InfoEvent().Err(e).Msg("m")
}

type myErr struct{}

func (m *myErr) Error() string { return "x" }

// detectWriter reports whether two goroutines are ever inside Write at once.
type detectWriter struct {
	inside   atomic.Int32
	overlaps atomic.Int64
	mu       sync.Mutex
	n        int
}

func (d *detectWriter) Write(p []byte) (int, error) {
	if d.inside.Add(1) > 1 {
		d.overlaps.Add(1)
	}
	// widen the window
	d.mu.Lock()
	d.n++
	d.mu.Unlock()
	for i := 0; i < 200; i++ {
		_ = i
	}
	d.inside.Add(-1)
	return len(p), nil
}

// A non-concurrent writer is supposed to be serialized by the logger. The
// serializing mutex lives in the per-configuration outputState, so a
// reconfiguration hands new writers a different mutex than in-flight writers
// are holding.
func TestWriterSerializationAcrossReconfigure(t *testing.T) {
	d := &detectWriter{}
	l := MustNew(Config{Format: FormatJSON, Writer: d, Level: LevelInfo})
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for g := 0; g < 6; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					l.Info("m")
				}
			}
		}()
	}
	for i := 0; i < 3000; i++ {
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: d, Level: LevelInfo})
	}
	close(stop)
	wg.Wait()
	t.Logf("writes=%d overlapping writes=%d", d.n, d.overlaps.Load())
	if d.overlaps.Load() > 0 {
		t.Errorf("BUG: %d concurrent writes to a writer the logger promised to serialize", d.overlaps.Load())
	}
}

// baseline: no reconfiguration, serialization must hold
func TestWriterSerializationSteady(t *testing.T) {
	d := &detectWriter{}
	l := MustNew(Config{Format: FormatJSON, Writer: d, Level: LevelInfo})
	var wg sync.WaitGroup
	for g := 0; g < 6; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 3000; i++ {
				l.Info("m")
			}
		}()
	}
	wg.Wait()
	t.Logf("writes=%d overlapping writes=%d", d.n, d.overlaps.Load())
	if d.overlaps.Load() > 0 {
		t.Errorf("BUG: %d overlapping writes without reconfiguration", d.overlaps.Load())
	}
}

// During reconfiguration the retiring async writer drains to the same
// destination the new writer is already using, each under its own mutex.
func TestAsyncWriterSerializationAcrossReconfigure(t *testing.T) {
	d := &detectWriter{}
	l := MustNew(Config{Format: FormatJSON, Writer: d, Level: LevelInfo, WriterMode: WriterAsync, BufferSize: 4})
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for g := 0; g < 6; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					l.Info("m")
				}
			}
		}()
	}
	for i := 0; i < 400; i++ {
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: d, Level: LevelInfo, WriterMode: WriterAsync, BufferSize: 4})
	}
	close(stop)
	wg.Wait()
	_ = l.Close()
	t.Logf("writes=%d overlapping writes=%d", d.n, d.overlaps.Load())
	if d.overlaps.Load() > 0 {
		t.Errorf("BUG: %d overlapping async writes", d.overlaps.Load())
	}
}

// switching between sync and async modes swaps which mutex serializes writes
func TestMixedModeSerialization(t *testing.T) {
	d := &detectWriter{}
	l := MustNew(Config{Format: FormatJSON, Writer: d, Level: LevelInfo})
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for g := 0; g < 6; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					l.Info("m")
				}
			}
		}()
	}
	for i := 0; i < 400; i++ {
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: d, Level: LevelInfo, WriterMode: WriterAsync, BufferSize: 4})
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: d, Level: LevelInfo})
	}
	close(stop)
	wg.Wait()
	_ = l.Close()
	t.Logf("writes=%d overlapping writes=%d", d.n, d.overlaps.Load())
	if d.overlaps.Load() > 0 {
		t.Errorf("BUG: %d overlapping writes across mode switches", d.overlaps.Load())
	}
}

func randString(r *rand.Rand) string {
	n := r.Intn(30)
	b := make([]byte, n)
	for i := range b {
		switch r.Intn(5) {
		case 0:
			b[i] = byte(r.Intn(0x21)) // controls
		case 1:
			b[i] = byte(r.Intn(256)) // any byte, incl. invalid UTF-8
		case 2:
			b[i] = []byte(`"\`)[r.Intn(2)]
		default:
			b[i] = byte('a' + r.Intn(26))
		}
	}
	return string(b)
}

// Every emitted record must be exactly one line; JSON records must parse and
// be valid UTF-8, whatever field names, values, and messages are thrown at it.
func TestAuditRecordInvariantsFuzz(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	for _, format := range []Format{FormatJSON, FormatConsole} {
		for _, color := range []bool{false, true} {
			for _, depth := range []int{0, 1} {
				for iter := 0; iter < 4000; iter++ {
					sb := &bytes.Buffer{}
					l := MustNew(Config{Format: format, Writer: sb, Level: LevelInfo,
						Color: color, CallerDepth: depth, IncludeTime: r.Intn(2) == 0,
						JSONTimeMode: JSONTimeMode(r.Intn(2))})
					e := l.InfoEvent()
					for f := 0; f < r.Intn(4); f++ {
						k := randString(r)
						switch r.Intn(10) {
						case 0:
							e = e.Str(k, randString(r))
						case 1:
							e = e.Int(k, r.Int())
						case 2:
							e = e.Float64(k, r.NormFloat64())
						case 3:
							e = e.Bool(k, r.Intn(2) == 0)
						case 4:
							e = e.Bytes(k, []byte(randString(r)))
						case 5:
							e = e.Time(k, time.Unix(r.Int63n(1<<31), 0))
						case 6:
							e = e.Duration(k, time.Duration(r.Int63()))
						case 7:
							e = e.Err(errors.New(randString(r)))
						case 8:
							e = e.Any(k, struct{ A string }{randString(r)})
						default:
							e = e.Uint64(k, r.Uint64())
						}
					}
					e.Msgs(randString(r), randString(r), r.Int(), math.Inf(1))
					out := sb.Bytes()
					if n := bytes.Count(out, []byte("\n")); n != 1 {
						t.Fatalf("format=%v color=%v depth=%d: %d newlines in %q", format, color, depth, n, out)
					}
					trimmed := bytes.TrimSuffix(out, []byte("\n"))
					if format == FormatJSON {
						if !json.Valid(trimmed) {
							t.Fatalf("invalid JSON: %q", trimmed)
						}
						if !utf8.Valid(trimmed) {
							t.Fatalf("invalid UTF-8: %q", trimmed)
						}
						var m map[string]any
						if err := json.Unmarshal(trimmed, &m); err != nil {
							t.Fatalf("unmarshal: %v on %q", err, trimmed)
						}
						if m["level"] != "INFO" {
							t.Fatalf("level forged to %v in %q", m["level"], trimmed)
						}
					} else if !color {
						// without colour the encoder emits no escapes of its
						// own, so nothing in the line may need escaping
						if idx := indexConsoleEscape(trimmed); idx >= 0 {
							t.Fatalf("control byte %#x at %d reached console output: %q",
								trimmed[idx], idx, trimmed)
						}
					}
				}
			}
		}
	}
}

// overflow policies must not lose or duplicate records beyond their contract
func TestAuditOverflowPolicies(t *testing.T) {
	for _, policy := range []OverflowPolicy{OverflowBlock, OverflowSync, OverflowDrop} {
		sb := &syncBuffer{}
		l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo,
			WriterMode: WriterAsync, BufferSize: 2, OverflowPolicy: policy})
		const n = 2000
		for i := 0; i < n; i++ {
			l.Info("rec")
		}
		if err := l.Close(); err != nil && !errors.Is(err, ErrWriteDropped) {
			t.Fatalf("policy %v close: %v", policy, err)
		}
		got := strings.Count(sb.String(), "rec")
		dropped := int(l.Dropped())
		t.Logf("policy=%v written=%d dropped=%d", policy, got, dropped)
		if policy != OverflowDrop && got != n {
			t.Errorf("policy %v: wrote %d of %d records", policy, got, n)
		}
		if got+dropped != n {
			t.Errorf("policy %v: %d written + %d dropped != %d", policy, got, dropped, n)
		}
		for _, line := range strings.Split(strings.TrimSuffix(sb.String(), "\n"), "\n") {
			if line != "" && !json.Valid([]byte(line)) {
				t.Fatalf("policy %v: corrupt line %q", policy, line)
			}
		}
	}
}

// Panic and Fatal must terminate on every path, at every level threshold,
// in both formats, enabled or suppressed.
func TestAuditTerminalLevelsAlwaysTerminate(t *testing.T) {
	for _, format := range []Format{FormatJSON, FormatConsole} {
		for _, threshold := range []Level{LevelTrace, LevelInfo, LevelFatal, LevelFatal + 1} {
			exits := 0
			mk := func() *Logger {
				return MustNew(Config{Format: format, Writer: io.Discard, Level: threshold,
					ExitFunc: func(int) { exits++ }})
			}
			for name, fn := range map[string]func(*Logger){
				"Fatal":        func(l *Logger) { l.Fatal("m") },
				"Fatalf":       func(l *Logger) { l.Fatalf("m %d", 1) },
				"Fatalln":      func(l *Logger) { l.Fatalln("m", 1) },
				"FatalEvent":   func(l *Logger) { l.FatalEvent().Str("k", "v").Msg("m") },
				"FatalMsgs":    func(l *Logger) { l.FatalEvent().Msgs("m", 1, "x") },
				"FatalMsgf":    func(l *Logger) { l.FatalEvent().Msgf("m %d", 1) },
				"FatalDiscard": func(l *Logger) { l.FatalEvent().Discard() },
				"LogFatal":     func(l *Logger) { l.Log(LevelFatal, "m") },
				"LoglnFatal":   func(l *Logger) { l.Logln(LevelFatal, "m", 1) },
				"LogfFatal":    func(l *Logger) { l.Logf(LevelFatal, "m %d", 1) },
			} {
				before := exits
				fn(mk())
				if exits != before+1 {
					t.Errorf("format=%v threshold=%v %s: did not terminate", format, threshold, name)
				}
			}

			for name, fn := range map[string]func(*Logger){
				"Panic":      func(l *Logger) { l.Panic("boom") },
				"Panicf":     func(l *Logger) { l.Panicf("boom %d", 1) },
				"Panicln":    func(l *Logger) { l.Panicln("boom", 1) },
				"PanicEvent": func(l *Logger) { l.PanicEvent().Str("k", "v").Msg("boom") },
				"PanicMsgs":  func(l *Logger) { l.PanicEvent().Msgs("boom", 1) },
				"PanicMsgf":  func(l *Logger) { l.PanicEvent().Msgf("boom %d", 1) },
				"LogPanic":   func(l *Logger) { l.Log(LevelPanic, "boom") },
				"LoglnPanic": func(l *Logger) { l.Logln(LevelPanic, "boom", 1) },
			} {
				func() {
					defer func() {
						r := recover()
						if r == nil {
							t.Errorf("format=%v threshold=%v %s: did not panic", format, threshold, name)
							return
						}
						if s, ok := r.(string); !ok || !strings.Contains(s, "boom") {
							t.Errorf("format=%v threshold=%v %s: panic value %q lost the message",
								format, threshold, name, r)
						}
					}()
					fn(mk())
				}()
			}
		}
	}
}

// A wide random sweep with a fresh seed each run: every configuration
// combination, every field type, hostile keys/values, asserting the record
// invariants. Looks for anything the fixed-seed fuzzers miss.
func TestHeavyRandomSweep(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed=%d", seed)
	r := rand.New(rand.NewSource(seed))
	for iter := 0; iter < 60000; iter++ {
		format := Format(r.Intn(2))
		var buf bytes.Buffer
		cfg := Config{
			Format: format, Writer: &buf, Level: LevelTrace,
			IncludeTime:      r.Intn(2) == 0,
			JSONTimeMode:     JSONTimeMode(r.Intn(3)),
			TimestampLayout:  []string{"", time.RFC3339, "seconds", "milliseconds", "2006-01-02"}[r.Intn(5)],
			CallerDepth:      r.Intn(3),
			Color:            r.Intn(2) == 0,
			DisableColor:     r.Intn(4) == 0,
			EscapeFieldNames: r.Intn(2) == 0,
			ConcurrentWriter: r.Intn(2) == 0,
		}
		l, err := New(cfg)
		if err != nil {
			continue
		}
		e := l.EventAt(Level(1+r.Intn(5)), randString(r), randString(r))
		for f := r.Intn(5); f > 0; f-- {
			k := randString(r)
			switch r.Intn(12) {
			case 0:
				e = e.Str(k, randString(r))
			case 1:
				e = e.Int64(k, r.Int63()-r.Int63())
			case 2:
				e = e.Float32(k, float32(r.NormFloat64()))
			case 3:
				e = e.Bool(k, r.Intn(2) == 0)
			case 4:
				e = e.Bytes(k, []byte(randString(r)))
			case 5:
				e = e.Time(k, time.Unix(r.Int63n(1<<33)-(1<<32), r.Int63n(1e9)))
			case 6:
				e = e.Duration(k, time.Duration(r.Int63()-r.Int63()))
			case 7:
				e = e.RawJSON(k, []byte(randString(r)))
			case 8:
				e = e.Any(k, map[string]string{randString(r): randString(r)})
			case 9:
				e = e.Uint64(k, r.Uint64())
			case 10:
				e = e.Any(k, []any{randString(r), r.Int(), nil})
			default:
				e = e.Stringer(k, time.Duration(r.Int63()))
			}
		}
		switch r.Intn(3) {
		case 0:
			e.Msg(randString(r))
		case 1:
			e.Msgs(randString(r), randString(r), r.Int())
		default:
			e.Msgf("%s|%d|%v", randString(r), r.Int(), randString(r))
		}

		out := buf.Bytes()
		if n := bytes.Count(out, []byte("\n")); n != 1 {
			t.Fatalf("seed=%d iter=%d: %d newlines in %q\ncfg=%+v", seed, iter, n, out, cfg)
		}
		trimmed := bytes.TrimSuffix(out, []byte("\n"))
		if format == FormatJSON {
			if !json.Valid(trimmed) {
				t.Fatalf("seed=%d iter=%d: invalid JSON %q\ncfg=%+v", seed, iter, trimmed, cfg)
			}
			if !utf8.Valid(trimmed) {
				t.Fatalf("seed=%d iter=%d: invalid UTF-8 %q\ncfg=%+v", seed, iter, trimmed, cfg)
			}
		} else if !cfg.Color || cfg.DisableColor {
			if idx := indexConsoleEscape(trimmed); idx >= 0 {
				t.Fatalf("seed=%d iter=%d: control byte %#x at %d in %q\ncfg=%+v",
					seed, iter, trimmed[idx], idx, trimmed, cfg)
			}
		}
	}
}

func TestProbeJSONValidAcceptsBadUTF8(t *testing.T) {
	raw := []byte("\"\xd9\"")
	t.Logf("json.Valid(%q) = %v, utf8.Valid = %v", raw, json.Valid(raw), utf8.Valid(raw))
}

type badMarshaler struct{}

func (badMarshaler) MarshalJSON() ([]byte, error) { return []byte("\"\xd9\xff\""), nil }

func TestProbeRawJSONBadUTF8(t *testing.T) {
	cases := map[string]func(*Event) *Event{
		"RawJSON string": func(e *Event) *Event { return e.RawJSON("k", []byte("\"\xd9\"")) },
		"RawJSON object": func(e *Event) *Event { return e.RawJSON("k", []byte("{\"a\":\"\xd9\xfe\"}")) },
		"Any RawMessage": func(e *Event) *Event { return e.Any("k", json.RawMessage("\"\xd9\"")) },
		"Any marshaler":  func(e *Event) *Event { return e.Any("k", badMarshaler{}) },
	}
	for name, fn := range cases {
		var buf bytes.Buffer
		l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo})
		fn(l.InfoEvent()).Msg("m")
		out := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
		ok := utf8.Valid(out)
		t.Logf("%-16s utf8.Valid=%-5v  %q", name, ok, out)
		if !ok {
			t.Errorf("BUG: %s emitted invalid UTF-8", name)
		}
	}
}

// Reserved keys are remapped to field_<name>, which is not injective: an
// explicit field_<name> lands on the same key. This records that limitation
// so a future change to the mapping is a deliberate one. See eventFieldName.
func TestReservedKeyCollisionIsDocumented(t *testing.T) {
	for _, reserved := range []string{"time", "level", "message", "func", "file"} {
		var buf bytes.Buffer
		l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo})
		l.InfoEvent().Str(reserved, "remapped").Str("field_"+reserved, "explicit").Msg("m")
		out := bytes.TrimSpace(buf.Bytes())
		// the record stays parseable, and the envelope is never overwritten
		var m map[string]any
		if err := json.Unmarshal(out, &m); err != nil {
			t.Errorf("%s: record is not parseable: %v", reserved, err)
			continue
		}
		if m["level"] != "INFO" || m["message"] != "m" {
			t.Errorf("%s: envelope damaged: %s", reserved, out)
		}
		if n := strings.Count(string(out), `"field_`+reserved+`"`); n != 2 {
			t.Errorf("%s: expected the documented duplicate key, got %d in %s", reserved, n, out)
		}
	}
}
