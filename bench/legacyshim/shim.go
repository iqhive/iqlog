package iqlog

import iqlog "github.com/iqhive/iqlog"

// Logger is the legacy logger wrapper.
type Logger struct{ *iqlog.Logger }

// NewGlobalIQLogger returns a legacy logger configured like the package default.
func NewGlobalIQLogger() *Logger { return &Logger{iqlog.NewGlobalIQLogger()} }

// GlobalLogger is the package-level legacy logger.
var GlobalLogger = NewGlobalIQLogger()

// Warnf logs a formatted warning through GlobalLogger.
func Warnf(format string, args ...any) { GlobalLogger.Warnf(format, args...) }

// Warn logs a warning through GlobalLogger.
func Warn(msg string, args ...any) { GlobalLogger.Warn(msg, args...) }

// Infof logs a formatted informational message through GlobalLogger.
func Infof(format string, args ...any) { GlobalLogger.Infof(format, args...) }

// Info logs an informational message through GlobalLogger.
func Info(msg string, args ...any) { GlobalLogger.Info(msg, args...) }
