package iqlog

// Regression coverage for the defects found by the September 2026 review:
// writer lifecycle after Close, syslog connection reuse, Flush racing a
// reconfiguration, queue-over-queue deadlock, Config() round-trips through
// the writer setters, colour detection, timestamp layouts and extreme years,
// message-argument encoding, typed-nil values, the console scan fast path,
// pooled buffer recycling, and caller name and line formatting.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- writer lifecycle ------------------------------------------------------

func TestWriterSettersAfterCloseDoNotLeak(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	before := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		if err := l.SetAsyncWriter(io.Discard); !errors.Is(err, ErrClosed) {
			t.Fatalf("SetAsyncWriter after Close: %v", err)
		}
		if err := l.SetRingBufferWriter(io.Discard); !errors.Is(err, ErrClosed) {
			t.Fatalf("SetRingBufferWriter after Close: %v", err)
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && runtime.NumGoroutine() > before {
		time.Sleep(5 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after > before {
		t.Fatalf("async writers built after Close leaked goroutines: before=%d after=%d", before, after)
	}
}

func TestSetSyslogHostAfterCloseReleasesConnection(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skip(err)
	}
	defer pc.Close()
	l := MustNew(Config{Writer: io.Discard})
	_ = l.Close()
	before := countFDs(t)
	for i := 0; i < 10; i++ {
		// alternate hosts so the unchanged short-circuit does not apply
		host := pc.LocalAddr().String()
		if i%2 == 1 {
			host = "127.0.0.1:1"
		}
		l.SetSyslogHost(host)
	}
	if after := countFDs(t); after > before {
		t.Fatalf("syslog connections dialled after Close leaked: fds before=%d after=%d", before, after)
	}
	if cfg := l.Config(); cfg.SyslogHost != "" {
		t.Fatalf("closed logger recorded SyslogHost %q", cfg.SyslogHost)
	}
}

type recordingCloser struct {
	bytes.Buffer
	closed int
}

func (r *recordingCloser) Close() error { r.closed++; return nil }

func TestOwnedWriterRefusesWritesAfterRelease(t *testing.T) {
	rc := &recordingCloser{}
	w := newOwnedWriter(rc)
	if _, err := w.Write([]byte("before")); err != nil {
		t.Fatal(err)
	}
	w.(*ownedWriter).release()
	w.(*ownedWriter).release()
	if rc.closed != 1 {
		t.Fatalf("closed %d times, want once", rc.closed)
	}
	n, err := w.Write([]byte("after"))
	if n != 0 || !errors.Is(err, ErrClosed) {
		t.Fatalf("write after release: n=%d err=%v, want ErrClosed", n, err)
	}
	if rc.String() != "before" {
		t.Fatalf("released writer still wrote: %q", rc.String())
	}
}

func TestReleasedSyslogWriterDoesNotRedial(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skip(err)
	}
	defer pc.Close()
	l := MustNew(Config{SyslogHost: pc.LocalAddr().String(), ApplicationName: "t", Level: LevelInfo})
	stale := l.Config().Writer
	if err := l.SetWriter(io.Discard); err != nil {
		t.Fatal(err)
	}
	before := countFDs(t)
	for i := 0; i < 5; i++ {
		if _, err := stale.Write([]byte("late\n")); !errors.Is(err, ErrClosed) {
			t.Fatalf("write through the released syslog writer: %v, want ErrClosed", err)
		}
	}
	if after := countFDs(t); after > before {
		t.Fatalf("released syslog writer re-dialled: fds before=%d after=%d", before, after)
	}
	_ = l.Close()
}

// gateWriter blocks its first Write until released, so a test can hold a
// queue drain open and observe what waits for it.
type gateWriter struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	mu      sync.Mutex
	n       int
}

func (g *gateWriter) Write(p []byte) (int, error) {
	g.once.Do(func() {
		close(g.entered)
		<-g.release
	})
	g.mu.Lock()
	g.n++
	g.mu.Unlock()
	return len(p), nil
}

func (g *gateWriter) count() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.n
}

func TestFlushWaitsForWriterRetiredByReconfiguration(t *testing.T) {
	for name, reconfigure := range map[string]func(*Logger, io.Writer) error{
		"SetConfig":      func(l *Logger, w io.Writer) error { return l.SetConfig(Config{Format: FormatJSON, Writer: w}) },
		"SetWriter":      func(l *Logger, w io.Writer) error { return l.SetWriter(w) },
		"SetAsyncWriter": func(l *Logger, w io.Writer) error { return l.SetAsyncWriter(w) },
		"Close":          func(l *Logger, _ io.Writer) error { return l.Close() },
	} {
		t.Run(name, func(t *testing.T) {
			g := &gateWriter{entered: make(chan struct{}), release: make(chan struct{})}
			l := MustNew(Config{Format: FormatJSON, Writer: g, WriterMode: WriterAsync, BufferSize: 16})
			const records = 5
			for i := 0; i < records; i++ {
				l.Info("queued")
			}
			<-g.entered // the drain has started and is now blocked
			reconfigured := make(chan error, 1)
			go func() { reconfigured <- reconfigure(l, g) }()
			// give the reconfiguration time to swap the writer and start retiring the old one
			time.Sleep(20 * time.Millisecond)
			flushed := make(chan struct{})
			go func() { _ = l.Flush(); close(flushed) }()
			select {
			case <-flushed:
				t.Fatalf("Flush returned with %d/%d records written while the retired queue was still draining", g.count(), records)
			case <-time.After(50 * time.Millisecond):
			}
			close(g.release)
			<-flushed
			if err := <-reconfigured; err != nil {
				t.Fatal(err)
			}
			if got := g.count(); got != records {
				t.Fatalf("Flush returned with %d/%d records written", got, records)
			}
			_ = l.Close()
		})
	}
}

func TestQueueOverQueueDoesNotDeadlock(t *testing.T) {
	for name, wrap := range map[string]func(*Logger) error{
		"SetAsyncWriter(GetWriter())":      func(l *Logger) error { return l.SetAsyncWriter(l.GetWriter()) },
		"SetRingBufferWriter(GetWriter())": func(l *Logger) error { return l.SetRingBufferWriter(l.GetWriter()) },
		"SetConfig(Writer: GetWriter())": func(l *Logger) error {
			return l.SetConfig(Config{Format: FormatJSON, Writer: l.GetWriter(), WriterMode: WriterAsync})
		},
	} {
		t.Run(name, func(t *testing.T) {
			sb := &syncBuffer{}
			l := MustNew(Config{Format: FormatJSON, Writer: sb, WriterMode: WriterAsync, BufferSize: 4})
			if err := wrap(l); err != nil {
				t.Fatal(err)
			}
			if _, nested := l.GetWriter().(*asyncWriter).out.(*asyncWriter); nested {
				t.Fatal("queue installed in front of another queue")
			}
			if got := l.Config().Writer; got != sb {
				t.Fatalf("Config().Writer = %T, want the destination", got)
			}
			done := make(chan struct{})
			go func() {
				l.Info("through the queue")
				_ = l.Flush()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("Flush deadlocked after wrapping the live queue")
			}
			if !strings.Contains(sb.String(), "through the queue") {
				t.Fatalf("record lost: %q", sb.String())
			}
			_ = l.Close()
		})
	}
}

func TestWriterSettersRoundTripThroughConfig(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: io.Discard})
	if err := l.SetAsyncWriter(sb); err != nil {
		t.Fatal(err)
	}
	cfg := l.Config()
	if cfg.Writer != sb || cfg.WriterMode != WriterAsync || cfg.BufferSize != 1000 || cfg.OverflowPolicy != OverflowBlock {
		t.Fatalf("Config() after SetAsyncWriter = writer %T mode %d size %d policy %d", cfg.Writer, cfg.WriterMode, cfg.BufferSize, cfg.OverflowPolicy)
	}
	if err := l.SetConfig(cfg); err != nil {
		t.Fatal(err)
	}
	aw, ok := l.GetWriter().(*asyncWriter)
	if !ok || aw.closed.Load() {
		t.Fatalf("round-trip lost the queue: %T closed=%v", l.GetWriter(), ok && aw.closed.Load())
	}
	l.Info("after round trip")
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), "after round trip") {
		t.Fatalf("record lost: %q", sb.String())
	}

	if err := l.SetRingBufferWriter(sb); err != nil {
		t.Fatal(err)
	}
	if cfg := l.Config(); cfg.WriterMode != WriterRing || cfg.BufferSize != 10000 || cfg.OverflowPolicy != OverflowSync {
		t.Fatalf("Config() after SetRingBufferWriter = mode %d size %d policy %d", cfg.WriterMode, cfg.BufferSize, cfg.OverflowPolicy)
	}
	if err := l.SetWriter(sb); err != nil {
		t.Fatal(err)
	}
	if cfg := l.Config(); cfg.WriterMode != WriterSync {
		t.Fatalf("Config() after SetWriter = mode %d", cfg.WriterMode)
	}
	_ = l.Close()
}

func TestSetWriterClearsSyslogHost(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skip(err)
	}
	defer pc.Close()
	var buf bytes.Buffer
	l := MustNew(Config{SyslogHost: pc.LocalAddr().String(), ApplicationName: "t", Level: LevelInfo})
	if err := l.SetWriter(&buf); err != nil {
		t.Fatal(err)
	}
	if host := l.Config().SyslogHost; host != "" {
		t.Fatalf("SetWriter left SyslogHost %q in the configuration", host)
	}
	if err := l.SetConfig(l.Config()); err != nil {
		t.Fatal(err)
	}
	l.Info("where")
	if !strings.Contains(buf.String(), "where") {
		t.Fatalf("Config() round-trip after SetWriter dropped the explicit writer: %q", buf.String())
	}
	_ = l.Close()
}

func TestSyslogAddr(t *testing.T) {
	for in, want := range map[string]string{
		"":                "",
		"127.0.0.1":       "127.0.0.1:514",
		"localhost":       "localhost:514",
		"localhost:601":   "localhost:601",
		"::1":             "[::1]:514",
		"[::1]":           "[::1]:514",
		"[::1]:514":       "[::1]:514",
		"2001:db8::1":     "[2001:db8::1]:514",
		"[2001:db8::1]:6": "[2001:db8::1]:6",
		"fe80::1%lo":      "[fe80::1%lo]:514",
	} {
		if got := syslogAddr(in); got != want {
			t.Errorf("syslogAddr(%q) = %q, want %q", in, got, want)
		}
	}
}

// --- colour ----------------------------------------------------------------

// openPTY returns a terminal file descriptor, or skips.
func openPTY(t *testing.T) *os.File {
	t.Helper()
	f, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Skip("no pty:", err)
	}
	t.Cleanup(func() { f.Close() })
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm")
	if !terminalWriter(f) {
		t.Skip("pty master not detected as a terminal")
	}
	return f
}

func TestDetectedColorIsNotForced(t *testing.T) {
	pty := openPTY(t)
	l := MustNew(Config{Writer: pty, Level: LevelInfo})
	if l.Config().Color {
		t.Fatal("Config() reports detected colour as forced")
	}
	if !l.config.Load().color {
		t.Fatal("terminal writer not coloured")
	}
	var buf bytes.Buffer
	if err := l.SetWriter(&buf); err != nil {
		t.Fatal(err)
	}
	l.Info("redirected")
	if strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("colour kept after SetWriter to a buffer: %q", buf.String())
	}
	buf.Reset()
	cfg := l.Config()
	cfg.Writer = &buf
	if err := l.SetConfig(cfg); err != nil {
		t.Fatal(err)
	}
	l.Info("round trip")
	if strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("Config() round-trip forced colour onto a buffer: %q", buf.String())
	}
	// and the other way round: a buffer logger moved onto the terminal
	if err := l.SetWriter(pty); err != nil {
		t.Fatal(err)
	}
	if !l.config.Load().color {
		t.Fatal("SetWriter to a terminal did not detect colour")
	}
}

func TestSyslogOutputIsNotColoredByTerminalStderr(t *testing.T) {
	pty := openPTY(t)
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skip(err)
	}
	defer pc.Close()
	Default() // built now so it never binds to the swapped stderr
	oldStderr := os.Stderr
	os.Stderr = pty
	defer func() { os.Stderr = oldStderr }()

	l := MustNew(Config{SyslogHost: pc.LocalAddr().String(), ApplicationName: "t", Level: LevelInfo})
	l.Info("hello syslog")
	_ = pc.SetReadDeadline(time.Now().Add(2 * time.Second))
	datagram := make([]byte, 4096)
	n, _, err := pc.ReadFrom(datagram)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(datagram[:n], []byte("\x1b[")) {
		t.Fatalf("ANSI colour in syslog datagram: %q", datagram[:n])
	}
	_ = l.Close()

	// SetSyslogHost from a terminal logger switches colour off as well
	l = MustNew(Config{Writer: pty, Level: LevelInfo})
	l.SetSyslogHost(pc.LocalAddr().String())
	l.Info("hello again")
	n, _, err = pc.ReadFrom(datagram)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(datagram[:n], []byte("\x1b[")) {
		t.Fatalf("ANSI colour in syslog datagram after SetSyslogHost: %q", datagram[:n])
	}
	_ = l.Close()
}

func TestSetUseColorHonoursDisableColor(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Writer: &buf, DisableColor: true, Level: LevelInfo})
	l.SetUseColor(true)
	l.Info("x")
	if strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("SetUseColor(true) overrode DisableColor: %q", buf.String())
	}
	if cfg := l.Config(); !cfg.Color || !cfg.DisableColor {
		t.Fatalf("Config() = Color %v DisableColor %v, want both true", cfg.Color, cfg.DisableColor)
	}
	buf.Reset()
	l = MustNew(Config{Writer: &buf, Level: LevelInfo})
	l.SetUseColor(true)
	l.Info("y")
	if !strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("SetUseColor(true) did not force colour: %q", buf.String())
	}
	l.SetUseColor(false)
	buf.Reset()
	l.Info("z")
	if strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("SetUseColor(false) left colour on: %q", buf.String())
	}
}

// --- timestamps ------------------------------------------------------------

func TestTimestampLayoutNamesResolveForCustomJSON(t *testing.T) {
	now := time.Date(2026, 8, 29, 14, 3, 12, 481947000, time.UTC)
	for name, want := range map[string]string{
		"seconds":      "2026-08-29T14:03:12Z",
		"milliseconds": "2026-08-29T14:03:12.481Z",
		"microseconds": "2026-08-29T14:03:12.481947Z",
		"nanoseconds":  "2026-08-29T14:03:12.481947Z",
		"Seconds":      "2026-08-29T14:03:12Z",
	} {
		var buf bytes.Buffer
		l := MustNew(Config{Format: FormatJSON, Writer: &buf, JSONTimeMode: JSONTimeCustom, TimestampLayout: name, Now: func() time.Time { return now }})
		l.Info("m")
		var rec struct{ Time string }
		if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
			t.Fatalf("%s: %v in %q", name, err, buf.String())
		}
		if rec.Time != want {
			t.Errorf("layout %q emitted time %q, want %q", name, rec.Time, want)
		}
	}
}

func TestFixedWidthTimestampsOutsideFourDigitYears(t *testing.T) {
	for _, year := range []int{-500, -1, 0, 1, 9999, 10000, 12345, 99999, 2026} {
		now := time.Date(year, 6, 7, 8, 9, 10, 123456000, time.UTC)
		wantConsole := now.Format("[2006-01-02T15:04:05.000000] ")
		wantJSON := now.Format("2006-01-02T15:04:05.000000Z07:00")

		var buf bytes.Buffer
		l := MustNew(Config{Writer: &buf, IncludeTime: true, DisableColor: true, Now: func() time.Time { return now }})
		l.Info("m")
		if !strings.HasPrefix(buf.String(), wantConsole) {
			t.Errorf("year %d console: %q, want prefix %q", year, buf.String(), wantConsole)
		}

		for _, depth := range []int{0, 1} { // cached path and direct path
			buf.Reset()
			l = MustNew(Config{Format: FormatJSON, Writer: &buf, JSONTimeMode: JSONTimeUTC, CallerDepth: depth, Now: func() time.Time { return now }})
			l.Info("first")
			l.Info("second") // the cached path must not have cached an unrepresentable year
			for _, line := range strings.SplitAfter(strings.TrimSpace(buf.String()), "\n") {
				var rec struct{ Time string }
				if err := json.Unmarshal([]byte(line), &rec); err != nil {
					t.Fatalf("year %d depth %d: %v in %q", year, depth, err, line)
				}
				if rec.Time != wantJSON {
					t.Errorf("year %d depth %d JSON time %q, want %q", year, depth, rec.Time, wantJSON)
				}
			}
		}
	}
}

// --- messages --------------------------------------------------------------

func TestMessageArgumentsMatchTypedFields(t *testing.T) {
	args := []any{1e-7, 123456789.123456789, float32(1e21), 0.1, 200, int8(-3), uint16(7), uint(9), true, math.Inf(1), math.NaN(), []int{1}}
	var jsonBuf, consoleBuf bytes.Buffer
	MustNew(Config{Format: FormatJSON, Writer: &jsonBuf}).Info("m", args...)
	MustNew(Config{Writer: &consoleBuf, DisableColor: true}).Info("m", args...)
	var rec struct{ Message string }
	if err := json.Unmarshal(jsonBuf.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	want := "m 0.0000001 123456789.12345679 1000000000000000000000 0.1 200 -3 7 9 true +Inf NaN [1]"
	if rec.Message != want {
		t.Errorf("JSON message %q\nwant %q", rec.Message, want)
	}
	if got := strings.TrimSuffix(strings.TrimPrefix(consoleBuf.String(), "INFO "), "\n"); got != want {
		t.Errorf("console message %q\nwant %q", got, want)
	}

	defer func() {
		if r := recover(); r != want {
			t.Errorf("panic value %q, want the console message %q", r, want)
		}
	}()
	MustNew(Config{Format: FormatJSON, Writer: io.Discard}).Panic("m", args...)
}

func TestMessageArgumentsDoNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	for name, l := range map[string]*Logger{
		"json":    MustNew(Config{Format: FormatJSON, Writer: io.Discard}),
		"console": MustNew(Config{Writer: io.Discard, DisableColor: true}),
	} {
		if n := testing.AllocsPerRun(1000, func() {
			l.Info("m", "s", 200, int64(-5), uint64(7), float32(1.5), 2.5, true)
		}); n != 0 {
			t.Errorf("%s: Info with scalar arguments allocated %.2f times", name, n)
		}
	}
}

// --- typed nil values ------------------------------------------------------

type ptrErr struct{ msg string }

func (p *ptrErr) Error() string { return p.msg }

type ptrStringer struct{ s string }

func (p *ptrStringer) String() string { return p.s }

func TestTypedNilPointersDoNotPanic(t *testing.T) {
	var pe *ptrErr
	var ps *ptrStringer
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatJSON, Writer: &buf})
	for name, log := range map[string]func(){
		"Err":         func() { l.InfoEvent().Err(pe).Msg("m") },
		"Any error":   func() { l.InfoEvent().Any("e", pe).Msg("m") },
		"Stringer":    func() { l.InfoEvent().Stringer("s", ps).Msg("m") },
		"WithError":   func() { l.WithError(pe).Info("m") },
		"WithFields":  func() { l.WithFields(map[string]any{"e": pe}).Info("m") },
		"slog":        func() { slog.New(l.SlogHandler()).Info("m", "e", pe) },
		"LogContext":  func() { l.LogWithFields(context.Background(), LevelInfo, map[string]any{"e": pe}, "m") },
		"real values": func() { l.InfoEvent().Err(&ptrErr{"boom"}).Stringer("s", &ptrStringer{"str"}).Msg("m") },
	} {
		buf.Reset()
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: panicked: %v", name, r)
				}
			}()
			log()
		}()
		if !json.Valid(bytes.TrimSpace(buf.Bytes())) {
			t.Errorf("%s: invalid record %q", name, buf.String())
		}
		if strings.Contains(name, "real") {
			if !strings.Contains(buf.String(), `"error":"boom"`) || !strings.Contains(buf.String(), `"s":"str"`) {
				t.Errorf("real values not encoded: %q", buf.String())
			}
		} else if strings.Contains(buf.String(), "boom") {
			t.Errorf("%s: %q", name, buf.String())
		}
	}
	if isNilPointer(nil) || isNilPointer(&ptrErr{}) || isNilPointer(ptrErr{}) || isNilPointer(map[string]int(nil)) || !isNilPointer(pe) {
		t.Fatal("isNilPointer misclassifies")
	}
}

// --- console scan fast path --------------------------------------------------

func TestConsoleWordSuspectIsExactForSingleBytes(t *testing.T) {
	for lane := 0; lane < 8; lane++ {
		for c := 0; c < 256; c++ {
			w := uint64(0x6161616161616161) // "aaaaaaaa"
			w &^= 0xff << (8 * lane)
			w |= uint64(c) << (8 * lane)
			want := c < 0x20 || c == 0x7f || c == 0xc2 || (c >= 0x80 && c <= 0x9f)
			if got := consoleWordSuspect(w); got != want {
				t.Fatalf("byte %#x in lane %d: suspect=%v want %v", c, lane, got, want)
			}
		}
	}
	for _, text := range []string{"café résumé naïve señor", "Zürich Ärger Größe", "日本語のテキスト", "Привет мир"} {
		b := []byte(text)
		if idx := indexConsoleEscape(b); idx >= 0 {
			t.Fatalf("%q flagged at %d", text, idx)
		}
	}
}

// --- pooled buffers ----------------------------------------------------------

func TestSanitizedConsoleRecordKeepsPooledBuffer(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard, DisableColor: true, WriterMode: WriterAsync, BufferSize: 4})
	defer l.Close()
	clean := l.InfoEvent().Str("k", "v")
	clean.Msg("clean line")
	if clean.output != nil {
		t.Fatal("clean async record: buffer not handed to the queue")
	}
	dirty := l.InfoEvent().Str("k", "v")
	dirty.Msg("dirty\nline")
	if dirty.output == nil || cap(dirty.output) == 0 {
		t.Fatal("sanitized async record: the event's own buffer was dropped instead of kept for reuse")
	}
	sync := MustNew(Config{Writer: io.Discard, DisableColor: true})
	e := sync.InfoEvent()
	e.Msg("dirty\nline")
	if e.output == nil {
		t.Fatal("sanitized sync record: buffer not kept")
	}
}

// --- callers ----------------------------------------------------------------

func TestCallerLineNumberSurvivesLongBasename(t *testing.T) {
	for _, basename := range []int{50, 96, 97, 98, 99, 100, 140} {
		for _, line := range []int{7, 14, 1234, 987654} {
			file := "/src/" + strings.Repeat("f", basename-3) + ".go"
			var data callerData
			data.set("pkg.Func", file, line)
			got := string(data.callerFile[:data.callerFileLen])
			suffix := ":" + strconv.Itoa(line)
			if !strings.HasSuffix(got, suffix) {
				t.Errorf("basename %d line %d: %q lacks %q", basename, line, got, suffix)
			}
			if len(got) > callerDataMaxLen {
				t.Errorf("basename %d line %d: %d bytes", basename, line, len(got))
			}
			if basename+len(suffix) <= callerDataMaxLen && got != file[len("/src/"):]+suffix {
				t.Errorf("basename %d line %d: %q was truncated needlessly", basename, line, got)
			}
		}
	}
}

func TestTrimRepoPathCoversReposModulesAndRootPackages(t *testing.T) {
	savePrefix, saveRoot, saveTrim, saveModules := mainModulePrefix, mainModuleRootPrefix, mainModuleRootTrim, modulePaths
	defer func() {
		mainModulePrefix, mainModuleRootPrefix, mainModuleRootTrim, modulePaths = savePrefix, saveRoot, saveTrim, saveModules
	}()
	mainModulePrefix, mainModuleRootPrefix, mainModuleRootTrim = "example.com/app/", "example.com/app.", len("example.com/")
	modulePaths = map[string]struct{}{}
	for _, path := range []string{"example.com/app", "github.com/iqhive/netsplain", "github.com/iqhive/netsplain/tools", "bitbucket.org/iqhive/tool/v2", "go.example.org/lib", "go.example.org/lib/nested", "gopkg.in/yaml.v3", "local"} {
		addModulePath(path)
	}
	for in, want := range map[string]string{
		// main module
		"example.com/app.RootLog":           "app.RootLog",
		"example.com/app.(*T).Method":       "app.(*T).Method",
		"example.com/app/internal/x.SubLog": "internal/x.SubLog",
		"example.com/app/cmd/app.main":      "cmd/app.main",
		"example.com/apple.Other":           "example.com/apple.Other",
		"main.main":                         "main.main",
		"example.com/app.RootLog.func1":     "app.RootLog.func1",
		"example.com/app/internal/x.F[...]": "internal/x.F[...]",
		// code hosts lose host/owner/repository, whatever the module
		"github.com/iqhive/netsplain/pkg/client.(*Client).runQoSTestWithRequest": "pkg/client.(*Client).runQoSTestWithRequest",
		"github.com/iqhive/netsplain.Run":                                        "netsplain.Run",
		"github.com/iqhive/netsplain/pkg/client.F[...].func1":                    "pkg/client.F[...].func1",
		"github.com/iqhive/netsplain/tools/gen.F":                                "tools/gen.F",
		"bitbucket.org/iqhive/tool/v2/sub.F":                                     "v2/sub.F",
		"bitbucket.org/iqhive/tool/v2.F":                                         "v2.F",
		"github.com/other/repo/pkg/x.F":                                          "pkg/x.F",
		"github.com/other/repo.F":                                                "repo.F",
		"github.com/other/repo.js/x.F":                                           "x.F",
		"github.com/other/repo%2ejs.F":                                           "repo%2ejs.F",
		"gitlab.com/group/proj/a/b.F":                                            "a/b.F",
		"github.com/owneronly.F":                                                 "github.com/owneronly.F",
		// other dependency modules lose their module path
		"go.example.org/lib/http2.(*Framer).Write": "http2.(*Framer).Write",
		"go.example.org/lib/nested/deep.F":         "deep.F",
		"go.example.org/lib/nested.F":              "nested.F",
		"gopkg.in/yaml%2ev3.Marshal":               "yaml%2ev3.Marshal",
		"gopkg.in/yaml.v3/sub.F":                   "sub.F",
		"local/pkg.F":                              "pkg.F",
		// anything else is left alone
		"net/http.(*conn).serve":           "net/http.(*conn).serve",
		"fmt.Println":                      "fmt.Println",
		"example.org/unknown/pkg.F":        "example.org/unknown/pkg.F",
		"example.com/app-other/x.F":        "example.com/app-other/x.F",
		"github.com/iqhive/netsplainx/y.F": "y.F",
		"no/function/here":                 "no/function/here",
	} {
		if got := callerFuncName(in, CallerPathRelative); got != want {
			t.Errorf("callerFuncName(%q, CallerPathRelative) = %q, want %q", in, got, want)
		}
		for _, mode := range []CallerPathMode{CallerPathLong, CallerPathFile} {
			if got := callerFuncName(in, mode); got != in {
				t.Errorf("callerFuncName(%q, %d) = %q, want it unchanged", in, mode, got)
			}
		}
	}
	for in, want := range map[string]string{
		"github.com/iqhive/netsplain/pkg/client.(*Client).Run": "client.(*Client).Run",
		"github.com/iqhive/netsplain.Run":                      "netsplain.Run",
		"bitbucket.org/iqhive/tool/v2/sub.F":                   "sub.F",
		"example.com/app/internal/x.F[...].func1":              "x.F[...].func1",
		"gopkg.in/yaml%2ev3.Marshal":                           "yaml%2ev3.Marshal",
		"net/http.(*conn).serve":                               "http.(*conn).serve",
		"main.main":                                            "main.main",
	} {
		if got := callerFuncName(in, CallerPathShort); got != want {
			t.Errorf("callerFuncName(%q, CallerPathShort) = %q, want %q", in, got, want)
		}
	}
	// this module's own root package is trimmed to its package name
	mainModulePrefix, mainModuleRootPrefix, mainModuleRootTrim, modulePaths = savePrefix, saveRoot, saveTrim, saveModules
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	l.Info("here")
	fn, _ := lastJSONMap(t, buf)["func"].(string)
	if !strings.HasPrefix(fn, "iqlog.Test") {
		t.Fatalf("func %q, want the package-relative name", fn)
	}
}

// --- round two -------------------------------------------------------------

func TestAnyCyclicValueDoesNotOverflowTheStack(t *testing.T) {
	cyclicSlice := []any{nil}
	cyclicSlice[0] = cyclicSlice
	cyclicMap := map[string]any{}
	cyclicMap["self"] = cyclicMap
	for name, mk := range map[string]func() *Logger{
		"json":    func() *Logger { return MustNew(Config{Format: FormatJSON, Writer: &syncBuffer{}}) },
		"console": func() *Logger { return MustNew(Config{Writer: &syncBuffer{}, DisableColor: true}) },
	} {
		l := mk()
		buf := l.Config().Writer.(*syncBuffer)
		l.InfoEvent().Any("s", cyclicSlice).Any("m", cyclicMap).Msg("m")
		l.LogWithFields(context.Background(), LevelInfo, map[string]any{"s": cyclicSlice}, "m")
		slog.New(l.SlogHandler()).Info("m", "s", cyclicSlice)
		out := buf.String()
		if strings.Count(out, "\n") != 3 || strings.Count(out, "!ERROR:") < 3 {
			t.Fatalf("%s: %q", name, out)
		}
		if name == "json" {
			for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
				if !json.Valid([]byte(line)) {
					t.Fatalf("invalid JSON %q", line)
				}
			}
		}
		// ordinary generic values still render as before
		buf.buf.Reset()
		type point struct{ X, Y int }
		l.InfoEvent().Any("p", point{1, 2}).Msg("m")
		want := `"p":{"X":1,"Y":2}`
		if name == "console" {
			want = "p={1 2} "
		}
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("%s: %q lacks %q", name, buf.String(), want)
		}
	}
}

func TestLevelSettingsAreClampedConsistently(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	for _, tc := range []struct {
		set  Level
		want Level
	}{
		{LevelUnknown, LevelInfo}, {Level(-5), LevelTrace}, {LevelWarn, LevelWarn},
		{LevelFatal + 1, LevelFatal + 1}, {Level(1 << 32), LevelFatal + 1}, {Level(math.MaxInt), LevelFatal + 1}, {Level(1 << 31), LevelFatal + 1},
	} {
		l.SetLevel(tc.set)
		if got := l.Level(); got != tc.want {
			t.Errorf("SetLevel(%d): Level() = %d, want %d", tc.set, got, tc.want)
		}
		if got := l.Config().Level; got != tc.want {
			t.Errorf("SetLevel(%d): Config().Level = %d, want %d", tc.set, got, tc.want)
		}
		if l.config.Load().level != Level(l.level.Load()) {
			t.Errorf("SetLevel(%d): gates disagree: snapshot %d fast path %d", tc.set, l.config.Load().level, l.level.Load())
		}
		if err := l.SetConfig(l.Config()); err != nil || l.Level() != tc.want {
			t.Errorf("SetLevel(%d): round trip err=%v level=%d", tc.set, err, l.Level())
		}
	}
	// an out-of-range configured level suppresses everything, and says so
	var buf bytes.Buffer
	l = MustNew(Config{Writer: &buf, Level: Level(math.MaxInt)})
	l.Error("dropped")
	if buf.Len() != 0 || l.Enabled(LevelFatal) {
		t.Fatalf("huge level did not suppress: %q enabled=%v", buf.String(), l.Enabled(LevelFatal))
	}
	l = MustNew(Config{Writer: &buf, Level: Level(-1 << 40)})
	l.Trace("kept")
	if !strings.Contains(buf.String(), "kept") || !l.Enabled(LevelTrace) {
		t.Fatalf("negative level did not enable trace: %q", buf.String())
	}
}

func TestConfigAfterCloseKeepsWriter(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatJSON, Writer: &buf, WriterMode: WriterAsync, BufferSize: 8})
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	cfg := l.Config()
	if cfg.Writer != &buf || cfg.WriterMode != WriterAsync {
		t.Fatalf("Config() after Close = writer %T mode %d", cfg.Writer, cfg.WriterMode)
	}
	if l.GetWriter() != io.Discard {
		t.Fatal("closed logger still routes writes")
	}
	l.Info("dropped")
	reopened := MustNew(cfg)
	reopened.Info("reopened")
	_ = reopened.Close()
	if got := buf.String(); strings.Contains(got, "dropped") || !strings.Contains(got, "reopened") {
		t.Fatalf("reopened from a closed configuration: %q", got)
	}
}

func TestSetJSONTimeModeClampsInvalidModes(t *testing.T) {
	var buf bytes.Buffer
	l := MustNew(Config{Writer: &buf, Now: func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }})
	l.SetJSONTimeMode(JSONTimeMode(9))
	if got := l.Config().JSONTimeMode; got != JSONTimeCustom {
		t.Fatalf("JSONTimeMode = %d", got)
	}
	if err := l.SetConfig(l.Config()); err != nil {
		t.Fatalf("Config() round trip after legacy setter: %v", err)
	}
	l.Info("m")
	if !json.Valid(bytes.TrimSpace(buf.Bytes())) || !strings.Contains(buf.String(), `"time":"2026-01-02T03:04:05.000000"`) {
		t.Fatalf("record %q", buf.String())
	}
}

func TestLogWithFieldsDisabledDoesNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	l := MustNew(Config{Writer: io.Discard, Level: LevelError})
	fields := map[string]any{"k": "v"}
	ctx := context.Background()
	if n := testing.AllocsPerRun(200, func() {
		l.LogWithFields(ctx, LevelDebug, fields, "m", 1)
		l.LogFWithFields(ctx, LevelDebug, fields, "m %d", 1)
	}); n != 0 {
		t.Fatalf("disabled LogWithFields allocated %.2f times", n)
	}
	var buf bytes.Buffer
	l = MustNew(Config{Writer: &buf, Level: LevelInfo})
	l.LogWithFields(ctx, LevelInfo, fields, "m")
	l.LogFWithFields(ctx, LevelInfo, fields, "f %d", 1)
	if got := buf.String(); strings.Count(got, "k=v") != 2 || !strings.Contains(got, "f 1") {
		t.Fatalf("enabled LogWithFields: %q", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("suppressed Panic through LogWithFields did not panic")
		}
	}()
	MustNew(Config{Writer: io.Discard, Level: LevelFatal + 1}).LogWithFields(ctx, LevelPanic, fields, "still terminates")
}

func TestConcurrentCloseWaitsForDrain(t *testing.T) {
	g := &gateWriter{entered: make(chan struct{}), release: make(chan struct{})}
	l := MustNew(Config{Format: FormatJSON, Writer: g, WriterMode: WriterAsync, BufferSize: 16})
	const records = 4
	for i := 0; i < records; i++ {
		l.Info("queued")
	}
	<-g.entered
	first := make(chan error, 1)
	go func() { first <- l.Close() }()
	time.Sleep(20 * time.Millisecond)
	second := make(chan error, 1)
	go func() { second <- l.Close() }()
	select {
	case <-second:
		t.Fatalf("second Close returned with %d/%d records written while the first was still draining", g.count(), records)
	case <-time.After(50 * time.Millisecond):
	}
	close(g.release)
	if err := <-first; err != nil {
		t.Fatal(err)
	}
	if err := <-second; err != nil {
		t.Fatal(err)
	}
	if got := g.count(); got != records {
		t.Fatalf("%d/%d records written when Close returned", got, records)
	}
	if err := l.Flush(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Flush after Close: %v", err)
	}
}

// BenchmarkTrimRepoPath measures the default caller-name trim for a frame on
// a code host, in a main or dependency module elsewhere, and in no module.
func BenchmarkTrimRepoPath(b *testing.B) {
	savePrefix, saveRoot, saveTrim, saveModules := mainModulePrefix, mainModuleRootPrefix, mainModuleRootTrim, modulePaths
	b.Cleanup(func() {
		mainModulePrefix, mainModuleRootPrefix, mainModuleRootTrim, modulePaths = savePrefix, saveRoot, saveTrim, saveModules
	})
	mainModulePrefix, mainModuleRootPrefix, mainModuleRootTrim = "example.com/app/", "example.com/app.", len("example.com/")
	modulePaths = map[string]struct{}{}
	addModulePath("example.com/app")
	for i := 0; i < 100; i++ {
		addModulePath("github.com/vendor" + strconv.Itoa(i) + "/lib")
	}
	addModulePath("go.example.org/lib")
	for _, bc := range []struct{ name, fn string }{
		{"hosted", "github.com/iqhive/netsplain/pkg/client.(*Client).runQoSTestWithRequest"},
		{"main", "example.com/app/internal/server.(*Server).handle"},
		{"module", "go.example.org/lib/http2.(*Framer).Write"},
		{"stdlib", "net/http.(*conn).serve"},
		{"package-main", "main.main"},
	} {
		b.Run(bc.name, func(b *testing.B) {
			for b.Loop() {
				_ = trimRepoPath(bc.fn)
			}
		})
	}
}
