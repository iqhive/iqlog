package iqlog

import (
	"context"
	"fmt"
)

// LogInterface support from viper
func (l *logger) Print(msg string) {
	l.LogWithFields(LevelPrint, nil, msg)
}
func (l *logger) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.LogWithFields(LevelPrint, nil, msg)
}
func (l *logger) Println(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	l.LogWithFields(LevelPrint, nil, msg)
}

func (l *logger) Log(level Level, msg string) {
	l.LogWithFields(level, nil, msg)
}
func Log(level Level, msg string) {
	GlobalLogger.Log(level, msg)
}

func (l *logger) Logf(level Level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Log(level, msg)
}
func Logf(level Level, format string, args ...interface{}) {
	GlobalLogger.Logf(level, format, args...)
}

func (l *logger) LogFWithFields(level Level, fields map[string]any, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.LogWithFields(level, fields, msg)
}

func (l *logger) LogWithFields(level Level, fields map[string]any, msg string) {
	var logger *bytesliceLine
	switch level {
	case LevelTrace:
		logger = GlobalLogger.WithByteSliceLineTrace()
	case LevelDebug:
		logger = GlobalLogger.WithByteSliceLineDebug()
	case LevelInfo:
		logger = GlobalLogger.WithByteSliceLineInfo()
	case LevelWarn:
		logger = GlobalLogger.WithByteSliceLineWarn()
	case LevelError:
		logger = GlobalLogger.WithByteSliceLineError()
	case LevelPanic:
		logger = GlobalLogger.WithByteSliceLinePanic()
	case LevelFatal:
		logger = GlobalLogger.WithByteSliceLineFatal()
	default:
		fmt.Println("Invalid log level")
		return
	}

	for k, v := range fields {
		// TODO: handle different types
		logger = logger.Str(k, fmt.Sprintf("%v", v))
	}
	logger.Msg(msg)
}

type legacyHandler struct {
	fields map[string]any
	err    error
	ctx    context.Context
}

func (h *legacyHandler) Trace(msg string) {
	h.Log(LevelTrace, msg)
}
func (h *legacyHandler) Tracef(format string, args ...interface{}) {
	h.Logf(LevelTrace, format, args...)
}

func (h *legacyHandler) Debug(msg string) {
	h.Log(LevelDebug, msg)
}
func (h *legacyHandler) Debugf(format string, args ...interface{}) {
	h.Logf(LevelDebug, format, args...)
}

func (h *legacyHandler) Info(msg string) {
	h.Log(LevelInfo, msg)
}
func (h *legacyHandler) Infof(format string, args ...interface{}) {
	h.Logf(LevelInfo, format, args...)
}

func (h *legacyHandler) Warn(msg string) {
	h.Log(LevelWarn, msg)
}
func (h *legacyHandler) Warnf(format string, args ...interface{}) {
	h.Logf(LevelWarn, format, args...)
}

func (h *legacyHandler) Error(msg string) {
	h.Log(LevelError, msg)
}
func (h *legacyHandler) Errorf(format string, args ...interface{}) {
	h.Logf(LevelError, format, args...)
}

func (h *legacyHandler) Panic(msg string) {
	h.Log(LevelPanic, msg)
}
func (h *legacyHandler) Panicf(format string, args ...interface{}) {
	h.Logf(LevelPanic, format, args...)
}

func (h *legacyHandler) Fatal(msg string) {
	h.Log(LevelFatal, msg)
}
func (h *legacyHandler) Fatalf(format string, args ...interface{}) {
	h.Logf(LevelFatal, format, args...)
}

func (h *legacyHandler) Log(level Level, msg string) {
	GlobalLogger.LogWithFields(level, h.fields, msg)
}

func (h *legacyHandler) Logf(level Level, format string, args ...interface{}) {
	GlobalLogger.LogFWithFields(level, h.fields, format, args...)
}

func (l *logger) WithFields(fields map[string]any) *legacyHandler {
	return &legacyHandler{
		fields: fields,
	}
}
