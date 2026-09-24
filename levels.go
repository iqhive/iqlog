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

// terminal reports whether a level terminates the process (panics or exits).
func (l Level) terminal() bool { return l >= LevelPanic }

// Level returns the logger's current minimum level.
func (l *Logger) Level() Level {
	return Level(l.level.Load())
}

// defaultLevel is the minimum level an unset Config.Level selects, and so the
// level of the logger Default builds. slogHandler.Enabled relies on it to
// answer for a default logger that has not been built yet.
const defaultLevel = LevelInfo

// clampLevel maps a configured minimum level onto what the gates represent,
// the same way for Config.Level and for SetLevel: the zero value means unset
// and selects Info, anything below Trace enables everything Trace does, and
// anything above Fatal suppresses every record. Logger.level is a 32-bit
// atomic, so an unbounded value would otherwise be truncated in the fast
// gate while the snapshot kept it, and the two gates would disagree.
func clampLevel(level Level) Level {
	switch {
	case level == LevelUnknown:
		return defaultLevel
	case level < LevelUnknown:
		return LevelTrace
	case level > LevelFatal+1:
		return LevelFatal + 1
	}
	return level
}

// SetLevel sets the logger's minimum level. Safe to call while other
// goroutines are logging. A level above LevelFatal suppresses every record
// including Fatal and Panic, which still terminate. LevelUnknown selects
// the default, LevelInfo, as it does in Config.
func (l *Logger) setLevel(level Level) {
	level = clampLevel(level)
	// updateConfig republishes the level fast path from the same snapshot, so
	// the two gates in newEventContextAt cannot be left disagreeing
	l.updateConfig(func(cfg *loggerConfig) { cfg.level = level })
}

// Enabled reports whether level would be emitted.
//
// The level is tested first: it lives on the logger itself, whereas the
// closed flag is a further pointer away on the shared writer state. A level
// that is already too low answers without that dependent load, which is the
// common case for the slog handler's Enabled on a disabled level. Both flags
// are independent atomics, so the order does not change the result.
func (l *Logger) Enabled(level Level) bool {
	return l.Level() <= level && !l.writer.closed.Load()
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
