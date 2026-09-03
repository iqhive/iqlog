package iqlog

import (
	"fmt"
	"strings"
)

// LogLevel is our own logging level, follows slog for now
type Level int

const (
	LevelUnknown Level = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
	// LevelPrint is the historical print level alias.
	LevelPrint = LevelInfo
)

// Level returns the logger's current minimum level.
func (l *Logger) Level() Level {
	return Level(l.level.Load())
}

// SetLevel sets the logger's minimum level. Safe to call while other
// goroutines are logging. A level above LevelFatal suppresses every record
// including Fatal and Panic, which still terminate.
func (l *Logger) setLevel(level Level) {
	// updateConfig republishes the level fast path from the same snapshot, so
	// the two gates in newEventContextAt cannot be left disagreeing
	l.updateConfig(func(cfg *loggerConfig) { cfg.level = level })
}

// Enabled reports whether level would be emitted.
func (l *Logger) Enabled(level Level) bool {
	return !l.writer.closed.Load() && l.Level() <= level
}

// Enabled reports whether the default logger would emit level.
func Enabled(level Level) bool { return Default().Enabled(level) }

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelPanic:
		return "panic"
	case LevelFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLevel parses a level name.
func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(value) {
	case "trace":
		return LevelTrace, nil
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	case "panic":
		return LevelPanic, nil
	case "fatal":
		return LevelFatal, nil
	case "unknown":
		return LevelUnknown, nil
	default:
		return LevelUnknown, fmt.Errorf("iqlog: invalid level %q", value)
	}
}

// MarshalText encodes the level as its name. The zero value encodes as
// "unknown" rather than failing: it is the natural state of a Level field in
// a configuration struct that has not been set, and returning an error there
// makes the whole surrounding struct unmarshalable.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *Level) UnmarshalText(text []byte) error {
	parsed, err := ParseLevel(string(text))
	if err != nil {
		return err
	}
	*l = parsed
	return nil
}
