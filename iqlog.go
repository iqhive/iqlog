package iqlog

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

var defaultLogger atomic.Pointer[Logger]

const maxLineLen = 1024
const maxStringLen = 256

// Logger is an independently configurable logger. Its fields are private;
// construct one with New.
type Logger struct {
	ctx    context.Context
	err    error
	level  atomic.Int32
	config atomic.Pointer[loggerConfig]
	writer *writerState
	// writeErr records the most recent error returned by the output
	// writer; shared between logger copies and background writers
	writeErr *writeErrBox

	fields []field
}

type field struct {
	key   string
	value any
}

type writerState struct {
	mu      sync.Mutex
	active  atomic.Pointer[outputState]
	closed  atomic.Bool
	dropped atomic.Uint64
}

type outputState struct {
	out        io.Writer
	configured io.Writer
	concurrent bool
	mu         sync.Mutex
}

// New creates an independently configured logger.
func New(input Config) (*Logger, error) {
	base := normalizeConfig(Config{})
	ws := &writerState{}
	ws.active.Store(&outputState{out: io.Discard, configured: io.Discard, concurrent: true})
	l := &Logger{writer: ws, writeErr: &writeErrBox{}}
	l.config.Store(base.snapshot())
	l.level.Store(int32(base.Level))
	if err := l.setConfig(input); err != nil {
		return nil, err
	}
	return l, nil
}

// MustNew is New but panics if configuration is invalid.
func MustNew(input Config) *Logger {
	l, err := New(input)
	if err != nil {
		panic(err)
	}
	return l
}

// Default returns the logger used by package-level functions.
func Default() *Logger {
	if l := defaultLogger.Load(); l != nil {
		return l
	}
	l := MustNew(Config{CallerDepth: 1, IncludeTime: true})
	if defaultLogger.CompareAndSwap(nil, l) {
		return l
	}
	return defaultLogger.Load()
}

// SetDefault changes the logger used by package-level functions.
func SetDefault(l *Logger) {
	if l == nil {
		panic("iqlog: nil default logger")
	}
	defaultLogger.Store(l)
}

// writeErrBox holds the most recent writer error behind its own lock so
// background writer goroutines can record errors without touching the
// logger mutex.
type writeErrBox struct {
	mu      sync.Mutex
	err     error
	pending error
}

func (l *Logger) writeRecord(line []byte) bool {
	state := l.writer.active.Load()
	if async, ok := state.out.(*asyncWriter); ok {
		async.WriteOwned(line)
		return true
	}
	var n int
	var err error
	if state.concurrent {
		n, err = state.out.Write(line)
	} else {
		state.mu.Lock()
		n, err = state.out.Write(line)
		state.mu.Unlock()
	}
	if err == nil && n != len(line) {
		err = io.ErrShortWrite
	}
	l.recordWriteErr(err)
	return false
}

// recordWriteErr stores a writer error for later inspection via
// LastWriteError. Safe to call from background writer goroutines.
func (l *Logger) recordWriteErr(err error) {
	if err == nil || l.writeErr == nil {
		return
	}
	l.writeErr.mu.Lock()
	l.writeErr.err = err
	l.writeErr.pending = errors.Join(l.writeErr.pending, err)
	l.writeErr.mu.Unlock()
}

func (l *Logger) takeWriteError() error {
	if l.writeErr == nil {
		return nil
	}
	l.writeErr.mu.Lock()
	err := l.writeErr.pending
	l.writeErr.pending = nil
	l.writeErr.mu.Unlock()
	return err
}

// Dropped returns the number of records rejected by overflow-drop policy.
func (l *Logger) Dropped() uint64 { return l.writer.dropped.Load() }
func Dropped() uint64             { return Default().Dropped() }

// LastWriteError returns the most recent error returned by the logger's
// output writer, or nil if all writes have succeeded. Errors from async
// and ring-buffer background writes are also recorded here.
func (l *Logger) LastWriteError() error {
	if l.writeErr == nil {
		return nil
	}
	l.writeErr.mu.Lock()
	defer l.writeErr.mu.Unlock()
	return l.writeErr.err
}

// LastWriteError returns the most recent write error of the default logger.
func LastWriteError() error {
	return Default().LastWriteError()
}

func (l *Logger) copy() *Logger {
	// copy fields individually, sharing the mutex pointer so copies
	// serialize writes against the original logger
	nl := &Logger{
		ctx:      l.ctx,
		err:      l.err,
		writer:   l.writer,
		writeErr: l.writeErr,
	}
	nl.config.Store(l.config.Load())
	nl.level.Store(l.level.Load())
	nl.fields = append([]field(nil), l.fields...)
	return nl
}

// WithContext returns a copy of the default logger carrying ctx.
func WithContext(ctx context.Context) *Logger {
	nl := Default().copy()
	nl.ctx = ctx
	return nl
}
