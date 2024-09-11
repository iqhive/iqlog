package iqlog

// LogLevel is our own logging level, follows slog for now
type Level int

const (
	LevelTrace Level = -3
	LevelDebug Level = 0
	LevelInfo  Level = 1
	LevelPrint Level = 1
	LevelWarn  Level = 2
	LevelError Level = 3
	LevelPanic Level = 9
	LevelFatal Level = 10
)
