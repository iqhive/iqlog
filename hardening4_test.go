package iqlog

// Regression tests for the writer-ownership, log-injection, pooling,
// disabled-path allocation, and level-gate defects fixed alongside them.

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func countFDs(t *testing.T) int {
	ents, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Skip("no /proc")
	}
	return len(ents)
}

// 1. syslog connections iqlog dials must be closed when retired
func TestSyslogWriterLeak(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skip(err)
	}
	defer pc.Close()
	host := pc.LocalAddr().String()
	l := MustNew(Config{SyslogHost: host, ApplicationName: "t", Level: LevelInfo})
	before := countFDs(t)
	for i := 0; i < 20; i++ {
		if err := l.SetConfig(Config{SyslogHost: host, ApplicationName: "t", Level: LevelInfo}); err != nil {
			t.Fatal(err)
		}
	}
	after := countFDs(t)
	t.Logf("fds before=%d after=%d over 20 reconfigurations", before, after)
	if after > before+1 {
		t.Errorf("syslog sockets still leaking: +%d", after-before)
	}
	// the live writer must still work
	l.Info("still alive")
	if err := l.LastWriteError(); err != nil {
		t.Errorf("live syslog writer broken: %v", err)
	}
}

// a caller-supplied writer must never be closed by the logger
func TestDoesNotCloseCallerWriter(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	l := MustNew(Config{Writer: f, Level: LevelInfo})
	if err := l.SetWriter(io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("still open\n"); err != nil {
		t.Errorf("logger closed a caller-supplied writer: %v", err)
	}
}

// 2. hostile field names must not forge JSON fields
func TestFieldNameInjection(t *testing.T) {
	for _, key := range []string{`x","level":"CRITICAL","admin`, "tab\tkey", "ctl\x00key", `back\slash`, "uni\xffcode"} {
		var buf bytes.Buffer
		l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo})
		l.WithFields(map[string]any{key: "v"}).Info("hello")
		out := bytes.TrimSpace(buf.Bytes())
		var m map[string]any
		if err := json.Unmarshal(out, &m); err != nil {
			t.Errorf("key %q produced invalid JSON: %s (%v)", key, out, err)
			continue
		}
		if m["level"] != "INFO" || len(m) != 3 {
			t.Errorf("key %q forged fields: %v", key, m)
		}
	}
}

// 3. baPool must not retain oversized buffers
func TestBaPoolBounded(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard, Level: LevelInfo})
	big := strings.Repeat("x", 4<<20)
	l.Infof("%s", big)
	ba := acquireBytesAppender()
	t.Logf("pooled bytesAppender capacity: %d (ceiling %d)", cap(ba.Bytes), maxPooledCapacity)
	if cap(ba.Bytes) > maxPooledCapacity {
		t.Errorf("baPool still retains %d bytes", cap(ba.Bytes))
	}
}

// 4. disabled logging must not format its arguments
func TestDisabledAllocFree(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard, Level: LevelError})
	cases := map[string]func(){
		"Debug(msg, args...)":  func() { l.Debug("m", "arg1", 42, true) },
		"Debugf":               func() { l.Debugf("m %s %d", "arg1", 42) },
		"Debugln":              func() { l.Debugln("m", "arg1", 42) },
		"Debug(msg)":           func() { l.Debug("m") },
		"Print":                func() { l.Print("m", 42) },
		"Println":              func() { l.Println("m", 42) },
		"Printf":               func() { l.Printf("m %d", 42) },
		"Logf(debug)":          func() { l.Logf(LevelDebug, "m %d", 42) },
		"Logln(debug)":         func() { l.Logln(LevelDebug, "m", 42) },
		"Log(debug, msg, arg)": func() { l.Log(LevelDebug, "m", 42) },
	}
	for name, fn := range cases {
		if n := testing.AllocsPerRun(200, fn); n != 0 {
			t.Errorf("%s while disabled: %v allocs/op, want 0", name, n)
		}
	}
	// Print/Println at an enabled level must still emit
	var buf bytes.Buffer
	e := MustNew(Config{Writer: &buf, Level: LevelInfo})
	e.Print("a", 1)
	e.Println("b", 2)
	e.Logln(LevelInfo, "c", 3)
	e.Logf(LevelInfo, "d %d", 4)
	for _, want := range []string{"a1", "b 2", "c 3", "d 4"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("missing %q in %q", want, buf.String())
		}
	}
}

// 5. Fatal and Panic must terminate even when the snapshot gate suppresses them
func TestTerminalLevelsNeverSwallowed(t *testing.T) {
	exited := false
	l := MustNew(Config{Writer: io.Discard, Level: LevelInfo, ExitFunc: func(int) { exited = true }})
	l.updateConfig(func(cfg *loggerConfig) { cfg.level = LevelFatal + 1 })
	l.Fatal("boom")
	if !exited {
		t.Errorf("Fatal did not terminate")
	}
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Panic did not panic")
			} else if r != "boom two" {
				t.Errorf("panic value = %v, want %q", r, "boom two")
			}
		}()
		l.Panic("boom", "two")
	}()
}

// Level() and the record gate must agree
func TestLevelGatesAgree(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard, Level: LevelInfo})
	l.SetUseColor(true) // an unrelated updateConfig must not desync the level
	if got, want := Level(l.level.Load()), l.config.Load().level; got != want {
		t.Errorf("gates diverged: atomic=%v snapshot=%v", got, want)
	}
	if e := l.InfoEvent(); e == nil {
		t.Errorf("Enabled says info but InfoEvent returned nil")
	} else {
		e.Discard()
	}
}

// 6. control bytes from values must not reach the terminal
func TestConsoleControlBytes(t *testing.T) {
	hostile := "bob\x1b[2K\x1b[1;31mROOT\x1b[0m\nFAKE line\rmore\x08\x07"
	for _, color := range []bool{false, true} {
		var buf bytes.Buffer
		l := MustNew(Config{Format: FormatConsole, Writer: &buf, Level: LevelInfo, Color: color})
		l.InfoEvent().Str("user", hostile).Msg("login")
		out := buf.String()
		body := out[strings.Index(out, "user="):]
		if strings.ContainsAny(body, "\x1b\x08\x07\r") {
			t.Errorf("color=%v: control byte survived in %q", color, body)
		}
		if strings.Count(out, "\n") != 1 {
			t.Errorf("color=%v: line forged: %q", color, out)
		}
		if color && !strings.Contains(out, "\x1b[34m") {
			t.Errorf("color=%v: logger's own color prefix was mangled: %q", color, out)
		}
	}
}

// 7. invalid UTF-8 must not reach JSON output
func TestInvalidUTF8(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatJSON, Writer: &buf, Level: LevelInfo})
	l.InfoEvent().Str("k", "bad\xff\xfebytes").Msg("msg \xff end")
	out := bytes.TrimSpace(buf.Bytes())
	if !utf8.Valid(out) {
		t.Errorf("output is not valid UTF-8: %q", out)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid JSON: %s", out)
	}
	if m["k"] != "bad��bytes" {
		t.Errorf("k = %q, want U+FFFD substitution", m["k"])
	}
	if m["message"] != "msg � end" {
		t.Errorf("message = %q, want U+FFFD substitution", m["message"])
	}
}

// 8. Config() must round-trip without disabling color
func TestConfigRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatConsole, Writer: &buf, Level: LevelInfo, Color: true})
	for i := 0; i < 3; i++ {
		if err := l.SetConfig(l.Config()); err != nil {
			t.Fatal(err)
		}
	}
	if !l.Config().Color {
		t.Errorf("Config() round-trip lost Color")
	}
	l.Info("hi")
	if !strings.Contains(buf.String(), "\x1b[") {
		t.Errorf("color lost after round-trip: %q", buf.String())
	}
}

// SetWriter must keep the configured ConcurrentWriter setting
func TestSetWriterKeepsConcurrent(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard, ConcurrentWriter: true, Level: LevelInfo})
	if err := l.SetWriter(&syncBuffer{}); err != nil {
		t.Fatal(err)
	}
	if !l.writer.active.Load().concurrent {
		t.Errorf("SetWriter dropped ConcurrentWriter")
	}
}
