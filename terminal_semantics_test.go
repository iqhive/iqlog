package iqlog

import (
	"strings"
	"sync"
	"testing"
	"time"
)

type gatedWriter struct {
	entered chan struct{}
	release chan struct{}
	mu      sync.Mutex
	lines   []string
	once    sync.Once
}

func (g *gatedWriter) Write(p []byte) (int, error) {
	g.once.Do(func() { close(g.entered) })
	<-g.release
	g.mu.Lock()
	g.lines = append(g.lines, string(p))
	n := len(p)
	g.mu.Unlock()
	return n, nil
}

func (g *gatedWriter) count() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.lines)
}

func (g *gatedWriter) output() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return strings.Join(g.lines, "")
}

func newTerminalTestLogger(w *gatedWriter, exit func(int)) *Logger {
	return MustNew(Config{
		Format:         FormatJSON,
		Writer:         w,
		Level:          LevelTrace,
		WriterMode:     WriterAsync,
		BufferSize:     8,
		OverflowPolicy: OverflowBlock,
		ExitFunc:       exit,
	})
}

func enqueueAcceptedRecords(t *testing.T, l *Logger, w *gatedWriter) {
	t.Helper()
	for i := 0; i < 5; i++ {
		l.Info("accepted")
	}
	select {
	case <-w.entered:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("writer did not enter its gate")
	}
}

func waitTerminalCall(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("terminal call did not complete after releasing writer")
	}
}

func TestFatalDiscardFlushesAcceptedRecords(t *testing.T) {
	w := &gatedWriter{entered: make(chan struct{}), release: make(chan struct{})}
	var exitCode, writtenAtExit int
	l := newTerminalTestLogger(w, func(code int) {
		exitCode = code
		writtenAtExit = w.count()
	})
	enqueueAcceptedRecords(t, l, w)

	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		close(started)
		l.FatalEvent().Discard()
		close(done)
	}()
	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("terminal call goroutine did not start")
	}
	select {
	case <-done:
		t.Fatal("FatalEvent().Discard returned before flushing accepted records")
	case <-time.After(50 * time.Millisecond):
	}
	close(w.release)
	waitTerminalCall(t, done)

	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if writtenAtExit != 5 {
		t.Fatalf("written at exit = %d, want 5", writtenAtExit)
	}
	if strings.Contains(w.output(), "discarded") {
		t.Fatalf("discarded fatal record was written: %q", w.output())
	}
}

func TestPanicTerminationFlushesAcceptedRecords(t *testing.T) {
	cases := []struct {
		name string
		call func(*Logger)
		want int
	}{
		{name: "PanicEvent Msg", call: func(l *Logger) { l.PanicEvent().Msg("boom") }, want: 6},
		{name: "Panicf", call: func(l *Logger) { l.Panicf("boom %d", 1) }, want: 6},
		{name: "Panic", call: func(l *Logger) { l.Panic("boom") }, want: 6},
		{name: "Log panic", call: func(l *Logger) { l.Log(LevelPanic, "boom") }, want: 6},
		{name: "PanicEvent Discard", call: func(l *Logger) { l.PanicEvent().Discard() }, want: 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := &gatedWriter{entered: make(chan struct{}), release: make(chan struct{})}
			l := newTerminalTestLogger(w, func(int) {})
			enqueueAcceptedRecords(t, l, w)

			started := make(chan struct{})
			done := make(chan struct{})
			var recovered any
			go func() {
				defer func() {
					recovered = recover()
					close(done)
				}()
				close(started)
				tc.call(l)
			}()
			select {
			case <-started:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("terminal call goroutine did not start")
			}
			select {
			case <-done:
				t.Fatal("panic termination returned before flushing accepted records")
			case <-time.After(50 * time.Millisecond):
			}
			close(w.release)
			waitTerminalCall(t, done)
			if w.count() != tc.want {
				t.Fatalf("written records = %d, want %d", w.count(), tc.want)
			}
			if recovered == nil {
				t.Fatal("panic termination did not propagate a panic")
			}
		})
	}
}

func TestSuppressedPanicFlushesAcceptedRecords(t *testing.T) {
	w := &gatedWriter{entered: make(chan struct{}), release: make(chan struct{})}
	l := newTerminalTestLogger(w, func(int) {})
	enqueueAcceptedRecords(t, l, w)
	l.SetLevel(LevelFatal + 1)

	started := make(chan struct{})
	done := make(chan struct{})
	var recovered any
	go func() {
		defer func() {
			recovered = recover()
			close(done)
		}()
		close(started)
		l.PanicEvent().Msg("suppressed panic")
	}()
	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("terminal call goroutine did not start")
	}
	select {
	case <-done:
		t.Fatal("suppressed panic returned before flushing accepted records")
	case <-time.After(50 * time.Millisecond):
	}
	close(w.release)
	waitTerminalCall(t, done)
	if w.count() != 5 {
		t.Fatalf("written records = %d, want 5", w.count())
	}
	if recovered == nil {
		t.Fatal("suppressed panic did not propagate a panic")
	}
}

func TestOverflowDropKeepsTerminalRecords(t *testing.T) {
	cases := []struct {
		name string
		call func(*Logger)
		want string
	}{
		{name: "fatal", call: func(l *Logger) { l.Fatal("fatal-terminal") }, want: "fatal-terminal"},
		{name: "panic", call: func(l *Logger) { l.Panic("panic-terminal") }, want: "panic-terminal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := &gatedWriter{entered: make(chan struct{}), release: make(chan struct{})}
			var exits int
			l := MustNew(Config{
				Format:         FormatJSON,
				Writer:         w,
				Level:          LevelTrace,
				WriterMode:     WriterAsync,
				BufferSize:     1,
				OverflowPolicy: OverflowDrop,
				ExitFunc:       func(int) { exits++ },
			})
			l.Info("queued")
			select {
			case <-w.entered:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("writer did not enter its gate")
			}
			l.Info("queued-second")
			l.Info("dropped")
			if l.Dropped() == 0 {
				t.Fatal("overflow setup did not drop an ordinary record")
			}
			droppedBeforeTerminal := l.Dropped()

			started := make(chan struct{})
			done := make(chan struct{})
			var recovered any
			go func() {
				defer func() {
					recovered = recover()
					close(done)
				}()
				close(started)
				tc.call(l)
			}()
			select {
			case <-started:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("terminal call goroutine did not start")
			}
			select {
			case <-done:
				t.Fatal("terminal record was dropped instead of drained synchronously")
			case <-time.After(50 * time.Millisecond):
			}
			close(w.release)
			waitTerminalCall(t, done)

			if tc.name == "fatal" && exits != 1 {
				t.Fatalf("fatal exit count = %d, want 1", exits)
			}
			if tc.name == "panic" && recovered == nil {
				t.Fatal("panic record did not propagate a panic")
			}
			if l.Dropped() != droppedBeforeTerminal {
				t.Fatalf("terminal record changed dropped count from %d to %d", droppedBeforeTerminal, l.Dropped())
			}
			if !strings.Contains(w.output(), tc.want) {
				t.Fatalf("terminal record missing from output: %q", w.output())
			}
		})
	}
}
