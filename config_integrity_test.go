package iqlog

// Configuration and lifecycle coverage: Config must survive a round-trip,
// reconfiguration must not leak goroutines, values that reach a transport
// header must not forge records there, and no public API path may panic on
// hostile input.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Config must survive a round-trip through SetConfig unchanged.
func TestAuditConfigRoundTripFidelity(t *testing.T) {
	sb := &syncBuffer{}
	cfgs := []Config{
		{Format: FormatJSON, Writer: sb, Level: LevelWarn, JSONTimeMode: JSONTimeUTC},
		{Format: FormatConsole, Writer: sb, Level: LevelTrace, IncludeTime: true, Color: true},
		{Format: FormatConsole, Writer: sb, Level: LevelError, DisableColor: true, CallerDepth: 3},
		{Format: FormatJSON, Writer: sb, Level: LevelDebug, JSONTimeMode: JSONTimeCustom, TimestampLayout: time.RFC3339},
		{Format: FormatJSON, Writer: sb, WriterMode: WriterAsync, BufferSize: 64, OverflowPolicy: OverflowDrop},
		{Format: FormatJSON, Writer: sb, WriterMode: WriterRing, BufferSize: 32, OverflowPolicy: OverflowSync},
		{Format: FormatJSON, Writer: sb, ConcurrentWriter: true, EscapeFieldNames: true, ApplicationName: "app"},
	}
	for i, want := range cfgs {
		l, err := New(want)
		if err != nil {
			t.Fatalf("cfg %d: %v", i, err)
		}
		first := l.Config()
		for r := 0; r < 3; r++ {
			if err := l.SetConfig(l.Config()); err != nil {
				t.Fatalf("cfg %d round %d: %v", i, r, err)
			}
		}
		got := l.Config()
		for _, f := range []struct {
			name string
			a, b any
		}{
			{"Format", first.Format, got.Format},
			{"Level", first.Level, got.Level},
			{"ConcurrentWriter", first.ConcurrentWriter, got.ConcurrentWriter},
			{"EscapeFieldNames", first.EscapeFieldNames, got.EscapeFieldNames},
			{"IncludeTime", first.IncludeTime, got.IncludeTime},
			{"TimestampLayout", first.TimestampLayout, got.TimestampLayout},
			{"JSONTimeMode", first.JSONTimeMode, got.JSONTimeMode},
			{"CallerDepth", first.CallerDepth, got.CallerDepth},
			{"Color", first.Color, got.Color},
			{"DisableColor", first.DisableColor, got.DisableColor},
			{"ApplicationName", first.ApplicationName, got.ApplicationName},
			{"NativeLog", first.NativeLog, got.NativeLog},
			{"WriterMode", first.WriterMode, got.WriterMode},
			{"BufferSize", first.BufferSize, got.BufferSize},
			{"OverflowPolicy", first.OverflowPolicy, got.OverflowPolicy},
		} {
			if f.a != f.b {
				t.Errorf("cfg %d: %s drifted %v -> %v", i, f.name, f.a, f.b)
			}
		}
		_ = l.Close()
	}
}

// repeated reconfiguration must not leak goroutines
func TestAuditNoGoroutineLeakAcrossReconfigure(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb})
	before := runtime.NumGoroutine()
	for i := 0; i < 300; i++ {
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: sb, WriterMode: WriterAsync, BufferSize: 4})
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: sb, WriterMode: WriterRing, BufferSize: 4})
		_ = l.SetConfig(Config{Format: FormatJSON, Writer: sb})
	}
	_ = l.Close()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("goroutines leaked: before=%d after=%d", before, runtime.NumGoroutine())
}

// a hostile syslog tag must not be able to forge a syslog record
func TestAuditSyslogTagInjection(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skip(err)
	}
	defer pc.Close()
	l, err := New(Config{SyslogHost: pc.LocalAddr().String(), ApplicationName: "app\nFAKE", Level: LevelInfo})
	if err != nil {
		t.Skip(err)
	}
	defer l.Close()
	l.Info("real")
	buf := make([]byte, 4096)
	_ = pc.SetReadDeadline(time.Now().Add(time.Second))
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		t.Skip(err)
	}
	got := string(buf[:n])
	t.Logf("syslog datagram: %q", got)
	if strings.Count(got, "\n") > 1 {
		t.Errorf("NOTE: syslog tag introduced %d newlines into one datagram", strings.Count(got, "\n"))
	}
}

// no public API path may panic on hostile or degenerate input
func TestAuditNoPanicFuzz(t *testing.T) {
	r := rand.New(rand.NewSource(29))
	for i := 0; i < 3000; i++ {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					t.Fatalf("iter %d panicked: %v", i, rec)
				}
			}()
			l := MustNew(Config{
				Format: Format(r.Intn(2)), Writer: io.Discard, Level: Level(1 + r.Intn(5)),
				CallerDepth: r.Intn(4), IncludeTime: r.Intn(2) == 0,
				JSONTimeMode: JSONTimeMode(r.Intn(3)), TimestampLayout: time.RFC3339,
				ContextExtractor: func(context.Context) map[string]any {
					return map[string]any{randString(r): randString(r)}
				},
			})
			s := randString(r)
			l.WithFields(map[string]any{s: s}).WithError(errors.New(s)).
				WithContext(context.Background()).Info(s, s, r.Int(), math.NaN())
			l.EventAt(Level(1+r.Intn(5)), s, s).Str(s, s).Any(s, nil).
				RawJSON(s, []byte(s)).Bytes(s, []byte(s)).
				Float64(s, math.Inf(-1)).Msgf("%s %d %v", s, r.Int(), nil)
			l.Event(Level(1 + r.Intn(5))).Discard()
			l.Logln(Level(1+r.Intn(5)), s, r.Int())
			l.Printf("%s", s)
			_ = l.Flush()
			_ = l.Close()
		}()
	}
}

// a writer whose dynamic type is a non-comparable value type
type sliceWriter struct {
	mu  *sync.Mutex
	buf *[]byte
	pad []byte // makes the struct non-comparable
}

func (w sliceWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	*w.buf = append(*w.buf, p...)
	w.mu.Unlock()
	return len(p), nil
}

func TestProbeNonComparableWriter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("BUG: replacing a non-comparable writer panicked: %v", r)
		}
	}()
	var b []byte
	w := sliceWriter{mu: &sync.Mutex{}, buf: &b}
	l := MustNew(Config{Format: FormatJSON, Writer: w, Level: LevelInfo})
	if err := l.SetWriter(w); err != nil {
		t.Fatal(err)
	}
	if err := l.SetWriter(sliceWriter{mu: &sync.Mutex{}, buf: &b}); err != nil {
		t.Fatal(err)
	}
	l.Info("still alive")
	if !strings.Contains(string(b), "still alive") {
		t.Errorf("record lost: %q", b)
	}
}

// retireWriter documents feeding GetWriter() back into SetConfig as legitimate
func TestProbeGetWriterRoundTripKeepsAsync(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo,
		WriterMode: WriterAsync, BufferSize: 16})
	w := l.GetWriter()
	if err := l.SetConfig(Config{Format: FormatJSON, Writer: w, Level: LevelInfo, WriterMode: WriterSync}); err != nil {
		t.Fatal(err)
	}
	l.Info("after round trip")
	_ = l.Flush()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && !strings.Contains(sb.String(), "after round trip") {
		time.Sleep(5 * time.Millisecond)
	}
	if !strings.Contains(sb.String(), "after round trip") {
		t.Errorf("BUG: record lost after GetWriter round trip: %q", sb.String())
	}
}

// Level text encoding must round-trip
func TestLevelTextRoundTrip(t *testing.T) {
	for _, lv := range []Level{LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelPanic, LevelFatal} {
		b, err := lv.MarshalText()
		if err != nil {
			t.Errorf("%v: marshal: %v", lv, err)
			continue
		}
		var got Level
		if err := got.UnmarshalText(b); err != nil {
			t.Errorf("%v: unmarshal %q: %v", lv, b, err)
			continue
		}
		if got != lv {
			t.Errorf("round trip %v -> %q -> %v", lv, b, got)
		}
	}
	// JSON round trip through a struct, the usual way configs are loaded
	type cfg struct{ Level Level }
	for _, s := range []string{`{"Level":"debug"}`, `{"Level":"WARNING"}`, `{"Level":"Fatal"}`} {
		var c cfg
		if err := json.Unmarshal([]byte(s), &c); err != nil {
			t.Errorf("%s: %v", s, err)
			continue
		}
		b, err := json.Marshal(c)
		if err != nil {
			t.Errorf("%s: remarshal: %v", s, err)
			continue
		}
		t.Logf("%-22s -> %v -> %s", s, c.Level, b)
	}
	// an unset Level in a struct must not make the whole config unmarshalable
	var c cfg
	if b, err := json.Marshal(c); err != nil {
		t.Logf("NOTE: marshalling a zero Level fails: %v", err)
	} else {
		t.Logf("zero Level marshals as %s", b)
	}
}
