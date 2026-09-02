package iqlog

import (
	"errors"
	"io"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type nativeRecord struct {
	level Level
	msg   string
}

// fakeNativeWriter implements nativeLogWriter for plumbing tests on
// platforms without a native system log.
type fakeNativeWriter struct {
	mu       sync.Mutex
	records  []nativeRecord
	released bool
}

func (f *fakeNativeWriter) Write(p []byte) (int, error) {
	return f.writeLevel(LevelUnknown, p)
}

func (f *fakeNativeWriter) writeLevel(level Level, p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.records = append(f.records, nativeRecord{level: level, msg: string(trimRecordNewline(p))})
	return len(p), nil
}

func (f *fakeNativeWriter) release() {
	f.mu.Lock()
	f.released = true
	f.mu.Unlock()
}

func (f *fakeNativeWriter) snapshot() []nativeRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]nativeRecord(nil), f.records...)
}

func (f *fakeNativeWriter) isReleased() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.released
}

// installFakeNative points the logger's synchronous output at fake, the
// state setConfig produces for Config{NativeLog: true, WriterMode:
// WriterSync} on supported platforms.
func installFakeNative(l *Logger, fake *fakeNativeWriter) {
	l.writer.active.Store(&outputState{out: fake, configured: fake, native: fake, concurrent: true})
}

func TestNativeSeverityFor(t *testing.T) {
	want := map[Level]nativeSeverity{
		LevelTrace:   nativeSeverityDebug,
		LevelDebug:   nativeSeverityDebug,
		LevelInfo:    nativeSeverityInfo,
		LevelWarn:    nativeSeverityNotice,
		LevelError:   nativeSeverityError,
		LevelPanic:   nativeSeverityFault,
		LevelFatal:   nativeSeverityFault,
		LevelUnknown: nativeSeverityNotice,
	}
	for level, severity := range want {
		if got := nativeSeverityFor(level); got != severity {
			t.Errorf("nativeSeverityFor(%v) = %d, want %d", level, got, severity)
		}
	}
}

func TestTrimRecordNewline(t *testing.T) {
	cases := map[string]string{
		"":          "",
		"\n":        "",
		"abc":       "abc",
		"abc\n":     "abc",
		"abc\n\n":   "abc\n",
		"abc\r\n":   "abc\r",
		"line\nx\n": "line\nx",
	}
	for in, want := range cases {
		if got := string(trimRecordNewline([]byte(in))); got != want {
			t.Errorf("trimRecordNewline(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNativeLogUnsupportedPlatform(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("native system log is supported on this platform")
	}
	if _, err := New(Config{NativeLog: true}); !errors.Is(err, errNativeLogUnsupported) {
		t.Fatalf("New(Config{NativeLog: true}) error = %v, want %v", err, errNativeLogUnsupported)
	}
	l := MustNew(Config{Writer: io.Discard})
	if err := l.SetConfig(Config{NativeLog: true}); !errors.Is(err, errNativeLogUnsupported) {
		t.Fatalf("SetConfig error = %v, want %v", err, errNativeLogUnsupported)
	}
	// The failed SetConfig must leave the previous configuration intact.
	if cfg := l.Config(); cfg.NativeLog || cfg.Writer != io.Discard {
		t.Fatalf("config changed after failed SetConfig: %+v", cfg)
	}
}

func TestNativeLogSyncReceivesLevels(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(l, fake)

	l.Info("hello")
	l.WarnEvent().Str("key", "value").Msg("warn")
	l.Error("boom")

	records := fake.snapshot()
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}
	if records[0].level != LevelInfo || !strings.HasSuffix(records[0].msg, "hello") {
		t.Errorf("record[0] = %+v", records[0])
	}
	if records[1].level != LevelWarn || !strings.HasSuffix(records[1].msg, "warn") {
		t.Errorf("record[1] = %+v", records[1])
	}
	if records[2].level != LevelError || !strings.HasSuffix(records[2].msg, "boom") {
		t.Errorf("record[2] = %+v", records[2])
	}
	for i, r := range records {
		if strings.HasSuffix(r.msg, "\n") {
			t.Errorf("record[%d] kept its trailing newline: %q", i, r.msg)
		}
	}
}

func TestNativeLogSyncReceivesJSON(t *testing.T) {
	l := MustNew(Config{Format: FormatJSON, Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(l, fake)

	l.InfoEvent().Str("key", "value").Msg("hello")

	records := fake.snapshot()
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	want := `{"level":"INFO","key":"value","message":"hello"}`
	if records[0].level != LevelInfo || records[0].msg != want {
		t.Errorf("record = %+v, want level info and %q", records[0], want)
	}
}

func TestNativeLogAsyncReceivesLevels(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	if err := l.setAsyncWriter(fake, 100, OverflowBlock); err != nil {
		t.Fatal(err)
	}

	l.Info("async-info")
	l.Error("async-error")
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}

	records := fake.snapshot()
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0].level != LevelInfo || !strings.HasSuffix(records[0].msg, "async-info") {
		t.Errorf("record[0] = %+v", records[0])
	}
	if records[1].level != LevelError || !strings.HasSuffix(records[1].msg, "async-error") {
		t.Errorf("record[1] = %+v", records[1])
	}
}

func TestNativeLogWriterReleasedOnClose(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(l, fake)
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if !fake.isReleased() {
		t.Fatal("native writer was not released on Close")
	}
}

func TestNativeLogWriterReleasedOnReplace(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(l, fake)
	if err := l.setWriter(io.Discard); err != nil {
		t.Fatal(err)
	}
	if !fake.isReleased() {
		t.Fatal("native writer was not released on writer replacement")
	}
}

func TestNativeLogWriterReleasedOnAsyncClose(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	if err := l.setAsyncWriter(fake, 100, OverflowBlock); err != nil {
		t.Fatal(err)
	}
	l.Info("queued")
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if !fake.isReleased() {
		t.Fatal("native writer was not released on async Close")
	}
	// The queued record must have been drained before release.
	if got := len(fake.snapshot()); got != 1 {
		t.Fatalf("got %d records after close, want 1", got)
	}
}

func TestNativeLogConcurrentWrites(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(l, fake)

	const goroutines = 8
	const perGoroutine = 50
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				l.Info("concurrent")
			}
		}()
	}
	wg.Wait()
	if got := len(fake.snapshot()); got != goroutines*perGoroutine {
		t.Fatalf("got %d records, want %d", got, goroutines*perGoroutine)
	}
}

func TestNativeMessageEscapesNUL(t *testing.T) {
	cases := map[string]string{
		"plain\n":        "plain",
		"a\x00b\n":       `a\0b`,
		"\x00\n":         `\0`,
		"\x00\x00":       `\0\0`,
		"trailing\x00\n": `trailing\0`,
	}
	for in, want := range cases {
		if got := string(nativeMessage([]byte(in))); got != want {
			t.Errorf("nativeMessage(%q) = %q, want %q", in, got, want)
		}
	}
	// The common NUL-free case must not copy.
	in := []byte("no copy\n")
	if out := nativeMessage(in); &out[0] != &in[0] {
		t.Error("nativeMessage copied a NUL-free record")
	}
}

// A Config() round-trip that turns NativeLog off hands the native writer
// back in as Config.Writer. Retiring the old state must not release it.
func TestNativeLogReusedAsWriterIsNotReleased(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(l, fake)
	if err := l.SetConfig(Config{Writer: fake}); err != nil {
		t.Fatal(err)
	}
	if fake.isReleased() {
		t.Fatal("native writer released while still configured as Writer")
	}
	l.Info("after round-trip")
	if got := len(fake.snapshot()); got != 1 {
		t.Fatalf("got %d records after round-trip, want 1", got)
	}
}

// SetAsyncWriter(GetWriter()) wraps the live native writer in a queue. The
// wrapped writer must survive retirement of the old synchronous state and
// keep receiving levels through the queue.
func TestNativeLogWrappedInAsyncIsNotReleased(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(l, fake)
	if err := l.setAsyncWriter(l.getWriter(), 100, OverflowBlock); err != nil {
		t.Fatal(err)
	}
	if fake.isReleased() {
		t.Fatal("native writer released while wrapped in the new async writer")
	}
	l.Warn("wrapped")
	if err := l.Flush(); err != nil {
		t.Fatal(err)
	}
	records := fake.snapshot()
	if len(records) != 1 || records[0].level != LevelWarn {
		t.Fatalf("records = %+v, want one warn record", records)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if !fake.isReleased() {
		t.Fatal("native writer not released once nothing references it")
	}
}

func TestNativeLogReleasedOnSetConfigReplacement(t *testing.T) {
	sync := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(sync, fake)
	if err := sync.SetConfig(Config{Writer: io.Discard}); err != nil {
		t.Fatal(err)
	}
	if !fake.isReleased() {
		t.Fatal("sync native writer not released by SetConfig")
	}

	async := MustNew(Config{Writer: io.Discard})
	queued := &fakeNativeWriter{}
	if err := async.setAsyncWriter(queued, 100, OverflowBlock); err != nil {
		t.Fatal(err)
	}
	async.Info("drain me")
	if err := async.SetConfig(Config{Writer: io.Discard}); err != nil {
		t.Fatal(err)
	}
	if !queued.isReleased() {
		t.Fatal("queued native writer not released by SetConfig")
	}
	if got := len(queued.snapshot()); got != 1 {
		t.Fatalf("got %d records, want the queue drained before release", got)
	}
}

// Writer methods install an explicit destination, so the NativeLog flag must
// not linger and re-enable the system log on a later Config() round-trip.
func TestWriterMethodsClearNativeLogFlag(t *testing.T) {
	for name, install := range map[string]func(*Logger) error{
		"SetWriter":      func(l *Logger) error { return l.setWriter(io.Discard) },
		"SetAsyncWriter": func(l *Logger) error { return l.setAsyncWriter(io.Discard, 10, OverflowBlock) },
	} {
		l := MustNew(Config{Writer: &syncBuffer{}})
		l.updateConfig(func(cfg *loggerConfig) { cfg.nativeLog = true })
		if !l.Config().NativeLog {
			t.Fatalf("%s: precondition failed", name)
		}
		if err := install(l); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if l.Config().NativeLog {
			t.Errorf("%s left NativeLog set in the configuration", name)
		}
	}
}

func TestNativeLogPlainWriteUsesUnknownLevel(t *testing.T) {
	fake := &fakeNativeWriter{}
	if n, err := fake.Write([]byte("raw\n")); err != nil || n != 4 {
		t.Fatalf("Write = %d, %v", n, err)
	}
	records := fake.snapshot()
	if len(records) != 1 || records[0].level != LevelUnknown || records[0].msg != "raw" {
		t.Fatalf("records = %+v", records)
	}
}
