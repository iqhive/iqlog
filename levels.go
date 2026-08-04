package iqlog

// LogLevel is our own logging level, follows slog for now
type Level int

const (
	LevelUnknown Level = -255
	LevelTrace   Level = -3
	LevelDebug   Level = 0
	LevelInfo    Level = 1
	LevelPrint   Level = 1
	LevelWarn    Level = 2
	LevelError   Level = 3
	LevelPanic   Level = 9
	LevelFatal   Level = 10
)

// Level returns the logger's current minimum level.
func (l *logger) Level() Level {
	return Level(l.level.Load())
}

// SetLevel sets the logger's minimum level. Safe to call while other
// goroutines are logging.
func (l *logger) SetLevel(level Level) {
	l.level.Store(int32(level))
}

// SetLevel sets the GlobalLogger's minimum level.
func SetLevel(level Level) {
	GlobalLogger.SetLevel(level)
}
