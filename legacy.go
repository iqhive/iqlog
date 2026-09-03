package iqlog

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func sprintln(args ...any) string {
	return strings.TrimSuffix(fmt.Sprintln(args...), "\n")
}

func Print(args ...any) {
	Default().Print(args...)
}

func Printf(format string, args ...any) {
	Default().Printf(format, args...)
}

func Println(args ...any) {
	Default().Println(args...)
}

func (l *Logger) Print(args ...any) {
	if l.skipRecord(LevelInfo) {
		return
	}
	l.Log(LevelInfo, fmt.Sprint(args...))
}

func (l *Logger) Printf(format string, args ...any) {
	l.Logf(LevelInfo, format, args...)
}

func (l *Logger) Println(args ...any) {
	if l.skipRecord(LevelInfo) {
		return
	}
	l.Log(LevelInfo, sprintln(args...))
}

func (l *Logger) Log(level Level, msg string, args ...any) {
	l.LogContext(l.ctx, level, msg, args...)
}

func Log(level Level, msg string, args ...any) {
	Default().Log(level, msg, args...)
}

func (l *Logger) Logf(level Level, format string, args ...any) {
	l.LogContextf(l.ctx, level, format, args...)
}

func Logf(level Level, format string, args ...any) {
	Default().Logf(level, format, args...)
}

func (l *Logger) Logln(level Level, args ...any) {
	level = normalizeLevel(level)
	if l.skipRecord(level) {
		return
	}
	l.Log(level, sprintln(args...))
}

func Logln(level Level, args ...any) {
	Default().Logln(level, args...)
}

func (l *Logger) LogContext(ctx context.Context, level Level, msg string, args ...any) {
	l.logContext(ctx, level, nil, msg, args...)
}

func LogContext(ctx context.Context, level Level, msg string, args ...any) {
	Default().LogContext(ctx, level, msg, args...)
}

func (l *Logger) LogContextf(ctx context.Context, level Level, format string, args ...any) {
	l.logContextf(ctx, level, format, args...)
}

func LogContextf(ctx context.Context, level Level, format string, args ...any) {
	Default().LogContextf(ctx, level, format, args...)
}

// WithFields returns a copy of the default logger with persistent fields.
func WithFields(fields map[string]any) *Logger {
	return Default().WithFields(fields)
}

// WithError returns a copy of the default logger with err attached.
func WithError(err error) *Logger {
	return Default().WithError(err)
}

func (l *Logger) logContext(ctx context.Context, level Level, fields map[string]any, msg string, args ...any) {
	line := l.newEventContext(ctx, normalizeLevel(level), 2)
	for k, v := range fields {
		addField(line, k, v)
	}
	line.Msgs(msg, args...)
}

// logContextf keeps the same call depth as logContext so caller attribution
// is unchanged, and hands the format to Msgf rather than formatting up front:
// a record the level gate drops must not pay for fmt.Sprintf.
func (l *Logger) logContextf(ctx context.Context, level Level, format string, args ...any) {
	l.newEventContext(ctx, normalizeLevel(level), 2).Msgf(format, args...)
}

// skipRecord reports whether an event at level would neither be written nor
// terminate the process, so callers can skip formatting their arguments.
func (l *Logger) skipRecord(level Level) bool {
	if level == LevelFatal || level == LevelPanic {
		return false
	}
	return Level(l.level.Load()) > level
}

func normalizeLevel(level Level) Level {
	switch level {
	case LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelPanic, LevelFatal:
		return level
	default:
		return LevelError
	}
}

func addField(line *Event, key string, value any) {
	switch value := value.(type) {
	case string:
		line.Str(key, value)
	case int:
		line.Int(key, value)
	case int64:
		line.Int64(key, value)
	case uint:
		line.Uint(key, value)
	case uint64:
		line.Uint64(key, value)
	case float32:
		line.Float32(key, value)
	case float64:
		line.Float64(key, value)
	case bool:
		line.Bool(key, value)
	case error:
		line.Str(key, value.Error())
	case time.Time:
		line.Time(key, value)
	case time.Duration:
		line.Duration(key, value)
	case []byte:
		line.Bytes(key, value)
	case json.RawMessage:
		line.RawJSON(key, value)
	default:
		line.Any(key, value)
	}
}

// WithContext returns a logger copy carrying ctx. The copy retains fields and
// shares writer synchronization and error state with the receiver.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	nl := l.copy()
	nl.ctx = ctx
	return nl
}

// WithFields returns a logger copy with fields added to every subsequent line.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	nl := l.copy()
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		replaced := false
		for i := range nl.fields {
			if nl.fields[i].key == key {
				nl.fields[i].value = fields[key]
				replaced = true
				break
			}
		}
		if !replaced {
			nl.fields = append(nl.fields, field{key: key, value: fields[key]})
		}
	}
	return nl
}

// WithError returns a logger copy that includes err under the "error" field.
func (l *Logger) WithError(err error) *Logger {
	nl := l.WithFields(map[string]any{"error": err})
	nl.err = err
	return nl
}
