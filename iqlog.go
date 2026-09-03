package iqlog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

var defaultLogger atomic.Pointer[Logger]

// maxLineLen is the initial capacity of the pooled format appender.
const maxLineLen = 1024

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

	// cfgMu serializes configuration updates so the config snapshot and the
	// level fast path can never be left disagreeing by interleaved writers.
	// Kept last: it is touched only by reconfiguration, never by logging, so
	// it must not push the level and config fields off the hot cache line.
	cfgMu sync.Mutex
}

type field struct {
	key   string
	value any
}

type writerState struct {
	// mu guards writer replacement.
	mu sync.Mutex
	// writeMu serializes writes to a writer that is not safe for concurrent
	// use. It lives here rather than in outputState because a reconfiguration
	// installs a new outputState: a per-state mutex would let a write holding
	// the old state run concurrently with one holding the new state, against
	// the same underlying writer. Logger copies share this writerState, so
	// they serialize against each other too.
	writeMu sync.Mutex
	active  atomic.Pointer[outputState]
	closed  atomic.Bool
	dropped atomic.Uint64
}

type outputState struct {
	out        io.Writer
	configured io.Writer
	concurrent bool
	// native is non-nil only when Config.NativeLog selected the platform
	// system log, keeping the level-aware branch off the default path.
	native nativeLogWriter
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

// maxPendingWriteErrors bounds how many distinct write errors a single Flush
// joins together. A destination that fails for every record would otherwise
// extend the pending chain without limit, turning a broken log sink into
// unbounded memory growth in the host process. Errors past the cap are
// counted instead of retained.
const maxPendingWriteErrors = 8

// writeErrBox holds the most recent writer error behind its own lock so
// background writer goroutines can record errors without touching the
// logger mutex.
type writeErrBox struct {
	mu sync.Mutex
	// err is the most recent error, reported by LastWriteError.
	err error
	// pending accumulates errors since the last Flush, up to
	// maxPendingWriteErrors; further ones only increment suppressed.
	pending      error
	pendingCount int
	suppressed   uint64
}

func (l *Logger) writeRecord(line []byte, level Level) bool {
	state := l.writer.active.Load()
	if async, ok := state.out.(*asyncWriter); ok {
		async.WriteOwned(line, level)
		return true
	}
	if state.native != nil {
		return l.writeNativeRecord(state.native, line, level)
	}
	var n int
	var err error
	if state.concurrent {
		n, err = state.out.Write(line)
	} else {
		n, err = l.writeSerialized(state.out, line)
	}
	if err == nil && n != len(line) {
		err = io.ErrShortWrite
	}
	l.recordWriteErr(err)
	return false
}

// writeSerialized writes through the logger's shared write mutex. The unlock
// is deferred because the destination is caller-supplied and may panic:
// releasing the lock only on the happy path would wedge every later record on
// this logger and all of its copies.
func (l *Logger) writeSerialized(w io.Writer, line []byte) (int, error) {
	l.writer.writeMu.Lock()
	defer l.writer.writeMu.Unlock()
	return w.Write(line)
}

// writeNativeRecord delivers the record to the platform system log. Native
// writers are safe for concurrent use and frame records themselves, so they
// bypass the serializing writer path. Kept out of writeRecord so the
// default sync path stays compact.
func (l *Logger) writeNativeRecord(native nativeLogWriter, line []byte, level Level) bool {
	n, err := native.writeLevel(level, line)
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
	box := l.writeErr
	box.mu.Lock()
	box.err = err
	switch {
	case box.pending == nil:
		box.pending = err
		box.pendingCount = 1
	case box.pendingCount < maxPendingWriteErrors:
		box.pending = errors.Join(box.pending, err)
		box.pendingCount++
	default:
		box.suppressed++
	}
	box.mu.Unlock()
}

func (l *Logger) takeWriteError() error {
	if l.writeErr == nil {
		return nil
	}
	box := l.writeErr
	box.mu.Lock()
	err, suppressed := box.pending, box.suppressed
	box.pending, box.pendingCount, box.suppressed = nil, 0, 0
	box.mu.Unlock()
	if suppressed > 0 {
		return fmt.Errorf("%w (and %d further write errors)", err, suppressed)
	}
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
	// derive the level from the same snapshot the copy carries so the two
	// gates in newEventContextAt cannot disagree in the copy
	cfg := l.config.Load()
	nl.config.Store(cfg)
	nl.level.Store(int32(cfg.level))
	nl.fields = append([]field(nil), l.fields...)
	return nl
}

// WithContext returns a copy of the default logger carrying ctx.
func WithContext(ctx context.Context) *Logger {
	nl := Default().copy()
	nl.ctx = ctx
	return nl
}
