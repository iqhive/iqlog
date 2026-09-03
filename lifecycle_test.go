package iqlog

// Lifecycle coverage: the logger must survive a destination that panics, keep
// working afterwards, stay correct under concurrent use of every API at once,
// and fail safely when an event is used after it has been consumed.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

type panicOnceWriter struct {
	mu       sync.Mutex
	panicked bool
	out      strings.Builder
}

func (w *panicOnceWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.panicked {
		w.panicked = true
		panic("writer exploded")
	}
	return w.out.Write(p)
}

// A caller-supplied writer may panic. If the logger does not release its write
// mutex on the way out, every later record blocks forever.
func TestPanickingWriterDoesNotWedgeLogger(t *testing.T) {
	w := &panicOnceWriter{}
	l := MustNew(Config{Format: FormatJSON, Writer: w, Level: LevelInfo})

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected the writer panic to propagate")
			}
		}()
		l.Info("first")
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		l.Info("second")
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("BUG: logger deadlocked after a writer panic")
	}
	if !strings.Contains(w.out.String(), "second") {
		t.Errorf("record lost: %q", w.out.String())
	}
}

// same for the async writer's send path, which holds a read lock while queueing
func TestPanickingWriterDoesNotWedgeAsyncClose(t *testing.T) {
	w := &panicOnceWriter{}
	l := MustNew(Config{Format: FormatJSON, Writer: w, Level: LevelInfo,
		WriterMode: WriterAsync, BufferSize: 1, OverflowPolicy: OverflowSync})
	// fill the queue so OverflowSync takes its synchronous fallback
	for i := 0; i < 50; i++ {
		func() {
			defer func() { _ = recover() }()
			l.Info("x")
		}()
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		_ = l.Close()
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("BUG: Close deadlocked after a writer panic")
	}
}

type alwaysPanicWriter struct{}

func (alwaysPanicWriter) Write([]byte) (int, error) { panic("sink exploded") }

// A background write has no call site to propagate to; taking the host process
// down because a log sink misbehaved is the wrong outcome for a logger.
func TestAsyncWriterPanicBecomesError(t *testing.T) {
	l := MustNew(Config{Format: FormatJSON, Writer: alwaysPanicWriter{}, Level: LevelInfo,
		WriterMode: WriterAsync, BufferSize: 8})
	l.Info("this sink panics")
	_ = l.Flush()
	err := l.LastWriteError()
	if err == nil || !strings.Contains(err.Error(), "panicked") {
		t.Errorf("expected the panic reported as a write error, got %v", err)
	} else {
		t.Logf("reported: %v", err)
	}
	// the logger must still work after being pointed somewhere sane
	sb := &syncBuffer{}
	if err := l.SetConfig(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo}); err != nil {
		t.Fatal(err)
	}
	l.Info("recovered")
	if !strings.Contains(sb.String(), "recovered") {
		t.Errorf("logger unusable after sink panic: %q", sb.String())
	}
}

// a writer that exposes Fd() but is not a terminal must not confuse detection
type fakeFd struct{ strings.Builder }

func (fakeFd) Fd() uintptr { return ^uintptr(0) }

func TestTerminalDetectionOnFakeFd(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("BUG: terminal detection panicked on a fake Fd: %v", r)
		}
	}()
	l := MustNew(Config{Format: FormatConsole, Writer: &fakeFd{}, Level: LevelInfo})
	if l.Config().Color {
		t.Errorf("a non-terminal Fd was detected as a terminal")
	}
	l.Info("m")
}

// Everything at once: logging in every style, reconfiguration, level changes,
// writer replacement, flush, and derived loggers, all concurrently.
func TestSoakConcurrentEverything(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelTrace})
	derived := l.WithFields(map[string]any{"svc": "soak"}).WithContext(context.Background())

	var wg sync.WaitGroup
	stop := make(chan struct{})
	closed := func() bool {
		select {
		case <-stop:
			return true
		default:
			return false
		}
	}

	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(id)))
			for !closed() {
				switch r.Intn(12) {
				case 0:
					l.Info("plain", id)
				case 1:
					l.InfoEvent().Str("k", randString(r)).Int("n", id).Msg("event")
				case 2:
					l.Debugf("fmt %d %s", id, randString(r))
				case 3:
					derived.Warnln("derived", id)
				case 4:
					l.WithFields(map[string]any{randString(r): id}).Error("fields")
				case 5:
					l.EventAt(LevelInfo, randString(r), randString(r)).Msg("at")
				case 6:
					_ = l.Flush()
				case 7:
					l.LogContext(context.Background(), LevelInfo, "ctx", id)
				case 8:
					if e := l.TraceEvent(); e != nil {
						e.Bytes("b", []byte(randString(r))).Discard()
					}
				case 9:
					_ = l.Enabled(LevelInfo)
					_ = l.Level()
					_ = l.Dropped()
					_ = l.LastWriteError()
				case 10:
					l.InfoEvent().RawJSON("r", []byte(`{"a":1}`)).Msg("raw")
				default:
					l.Printf("%d", id)
				}
			}
		}(g)
	}

	// reconfiguration churn
	wg.Add(1)
	go func() {
		defer wg.Done()
		modes := []WriterMode{WriterSync, WriterAsync, WriterRing}
		for i := 0; !closed(); i++ {
			_ = l.SetConfig(Config{
				Format: Format(i % 2), Writer: sb, Level: Level(1 + i%5),
				WriterMode: modes[i%3], BufferSize: 4, OverflowPolicy: OverflowPolicy(i % 3),
				CallerDepth: i % 2, IncludeTime: i%2 == 0, JSONTimeMode: JSONTimeMode(i % 3),
			})
			l.SetLevel(Level(1 + i%5))
			l.SetUseColor(i%2 == 0)
			l.SetCallerDepth(i % 3)
			if i%17 == 0 {
				_ = l.SetWriter(sb)
			}
		}
	}()

	time.Sleep(1500 * time.Millisecond)
	close(stop)
	wg.Wait()
	if err := l.Close(); err != nil && !errors.Is(err, ErrWriteDropped) {
		t.Fatalf("close: %v", err)
	}
	// after Close everything must be inert, not panicking
	l.Info("after close")
	l.InfoEvent().Msg("after close")
	if err := l.Flush(); !errors.Is(err, ErrClosed) {
		t.Errorf("Flush after Close = %v, want ErrClosed", err)
	}
	t.Logf("wrote %d bytes", len(sb.String()))
	_ = io.Discard
}

// Msg/Msgs/Msgf consume the event. Using it afterwards is documented misuse;
// what matters is that it fails safely rather than corrupting another
// goroutine's record.
func TestEventUseAfterConsume(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo})

	e := l.InfoEvent().Str("a", "1")
	e.Msg("first")
	// everything below is misuse
	e.Str("b", "2")
	e.Msg("second")
	e.Discard()
	e.Msgf("third %d", 3)

	l.InfoEvent().Str("c", "3").Msg("clean")

	lines := strings.Split(strings.TrimSuffix(sb.String(), "\n"), "\n")
	t.Logf("%d lines emitted", len(lines))
	for i, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Errorf("line %d is not valid JSON: %q", i, line)
		}
		t.Logf("  %s", line)
	}
	if !strings.Contains(sb.String(), "clean") {
		t.Errorf("a later, correct record was lost")
	}
}
