package iqlog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type slogEmu struct {
	slog.Handler

	Logger *slog.Logger
}

// We allow a fixed maximum of fields so we can try avoid an alloc
const maxFields = 8
const maxStringLen = 256

// LogField is used to store each field in a (pre-alloc'd) array
type LogField struct {
	Key    [maxStringLen]byte // fixed-size key
	KeyLen int                // length of the key
	VStr   [maxStringLen]byte // fixed-size string representation
	VLen   int                // length of the "string" portion
	Quote  bool               // whether to quote the value
	Used   bool               // whether this slot is in use
}

type LogRecord struct {
	Time       time.Time
	Level      Level
	Message    [maxStringLen]byte // fixed-size message buffer
	MsgLen     int                // length of message
	Fields     [maxFields]LogField
	UsedFields int
}

func (l *logger) WriteJSONRecord(record LogRecord) error {
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = l.out.Write(b)
	return err
}

// WriteRecord writes out a log record
func (l *logger) WriteRecord(ctx context.Context, record LogRecord) error {
	if l.jsonMode {
		// If you need JSON logs, just call WriteRecordAsJSONNoAlloc or similar, but that
		// requires hand-rolling JSON without any library helpers to avoid allocations.
		return l.WriteJSONRecord(record)
	}

	// If the log level is disabled, short-circuit quickly:
	if !l.Enabled(ctx, record.Level) {
		return nil
	}

	// Create a stack buffer of fixed size
	var out [512]byte
	used := 0

	if l.IncludeTimePrefix {
		timeStr := time.Now().Format("2006-01-02T15:04:05.999Z07:00")
		used += copy(out[used:], timeStr)
		used += len(timeStr)
		out[used] = ' '
		used++
	}

	// Level prefix (with optional colour)
	if l.useColour {
		used += copy(out[used:], colourPrefix(record.Level))
	} else {
		used += copy(out[used:], levelPrefix(record.Level))
	}

	// Optional application name
	if l.applicationName != "" {
		out[used] = '['
		used++
		used += copy(out[used:], l.applicationName)
		out[used] = ']'
		used++
		out[used] = ' '
		used++
	}

	// The msg
	msgLen := record.MsgLen
	if msgLen > maxStringLen {
		msgLen = maxStringLen
	}
	copy(out[used:], record.Message[:msgLen])
	used += msgLen

	// newline
	out[used] = '\n'
	used++

	_, err := l.out.Write(out[:used])
	return err
}

func colourPrefix(level Level) []byte {
	switch level {
	case LevelInfo:
		return []byte("\x1b[34mINFO\x1b[0m ")
	case LevelDebug, LevelTrace:
		return []byte("\x1b[37mDBUG\x1b[0m ")
	case LevelWarn:
		return []byte("\x1b[33mWARN\x1b[0m ")
	case LevelError:
		return []byte("\x1b[31mERRR\x1b[0m ")
	case LevelFatal:
		return []byte("\x1b[31mFATL\x1b[0m ")
	case LevelPanic:
		return []byte("\x1b[31mPANC\x1b[0m ")
	default:
		return []byte("\x1b[34m????\x1b[0m ")
	}
}

// levelPrefix returns a constant []byte for each level (no allocations)
func levelPrefix(level Level) []byte {
	switch level {
	case LevelInfo:
		return []byte("INFO ")
	case LevelDebug, LevelTrace:
		return []byte("DBUG ")
	case LevelWarn:
		return []byte("WARN ")
	case LevelError:
		return []byte("ERRR ")
	case LevelFatal:
		return []byte("FATL ")
	case LevelPanic:
		return []byte("PANC ")
	default:
		return []byte("???? ")
	}
}

func (l *slogEmu) Handle(ctx context.Context, entry slog.Record) error {
	return l.Handler.Handle(ctx, entry)
}

func (l *slogEmu) Log(ctx context.Context, level slog.Level, msg string) {
	if l == nil {
		return
	}
	if l.Handler == nil {
		l.Handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})
	}
	l.Handler.Handle(ctx, slog.Record{
		Level:   level,
		Message: msg,
	})
}

func (l *slogEmu) DebugContext(ctx context.Context, msg string) {
	l.Log(ctx, slog.LevelDebug, msg)
}

func (l *slogEmu) Debug(msg string) {
	l.Log(context.Background(), slog.LevelDebug, msg)
}

func (l *slogEmu) InfoContext(ctx context.Context, msg string) {
	l.Log(ctx, slog.LevelInfo, msg)
}

func (l *slogEmu) Info(msg string) {
	l.Log(context.Background(), slog.LevelInfo, msg)
}

func (l *slogEmu) PrintContext(ctx context.Context, msg string) {
	l.Log(ctx, slog.LevelInfo, msg)
}

func (l *slogEmu) Print(msg string) {
	l.Log(context.Background(), slog.LevelInfo, msg)
}

func (l *slogEmu) Printf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (l *slogEmu) WarnContext(ctx context.Context, msg string) {
	l.Log(ctx, slog.LevelWarn, msg)
}

func (l *slogEmu) Warn(msg string) {
	l.Log(context.Background(), slog.LevelWarn, msg)
}

func (l *slogEmu) Warnf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...))
}

func (l *slogEmu) ErrorContext(ctx context.Context, msg string) {
	l.Log(ctx, slog.LevelError, msg)
}

func (l *slogEmu) Error(msg string) {
	l.Log(context.Background(), slog.LevelError, msg)
}

func (l *slogEmu) Errorf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

func (l *slogEmu) FatalContext(ctx context.Context, msg string) {
	l.Log(ctx, slog.LevelError, msg)
}

func (l *slogEmu) Fatal(msg string) {
	l.Log(context.Background(), slog.LevelError, msg)
}

func (l *slogEmu) Fatalf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

func (l *slogEmu) PanicContext(ctx context.Context, msg string) {
	l.Log(ctx, slog.LevelError, msg)
}

func (l *slogEmu) Panic(msg string) {
	l.Log(context.Background(), slog.LevelError, msg)
}

func (l *slogEmu) Panicf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

func (l *slogEmu) LogContext(ctx context.Context, level slog.Level, msg string) {
	l.Log(ctx, level, msg)
}

func (l *slogEmu) WithError(err error) *slogEmu {
	l.Handler = l.Handler.WithAttrs([]slog.Attr{{
		Key:   "error",
		Value: slog.AnyValue(err),
	}})
	return l
}

func (l *slogEmu) SetApplicationName(name string) *slogEmu {
	return l
}
