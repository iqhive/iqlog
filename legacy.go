package iqlog

import (
	"context"
	"fmt"
)

// LogInterface support from viper
func (l *logger) Print(msg string, args ...any) {
	l.LogWithFields(context.Background(), LevelPrint, nil, msg, args...)
}
func (l *logger) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.LogWithFields(context.Background(), LevelPrint, nil, msg)
}
func (l *logger) Println(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	l.LogWithFields(context.Background(), LevelPrint, nil, msg)
}

func (l *logger) Log(ctx context.Context, level Level, msg string, args ...any) {
	l.LogWithFields(ctx, level, nil, msg, args...)
}
func Log(ctx context.Context, level Level, msg string) {
	GlobalLogger.Log(ctx, level, msg)
}

func (l *logger) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Log(ctx, level, msg)
}
func Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	GlobalLogger.Logf(ctx, level, format, args...)
}

func (l *logger) LogFWithFields(ctx context.Context, level Level, fields map[string]any, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.LogWithFields(ctx, level, fields, msg)
}

func (l *logger) LogWithFields(ctx context.Context, level Level, fields map[string]any, msg string, args ...any) {
	var logger *bytesliceLine
	switch level {
	case LevelTrace:
		logger = l.WithByteSliceLineTrace()
	case LevelDebug:
		logger = l.WithByteSliceLineDebug()
	case LevelInfo:
		logger = l.WithByteSliceLineInfo()
	case LevelWarn:
		logger = l.WithByteSliceLineWarn()
	case LevelError:
		logger = l.WithByteSliceLineError()
	case LevelPanic:
		logger = l.WithByteSliceLinePanic()
	case LevelFatal:
		logger = l.WithByteSliceLineFatal()
	default:
		// unknown level: log at error level rather than silently dropping
		// the record (or printing to stdout)
		logger = l.WithByteSliceLineError()
	}

	for k, v := range fields {
		// TODO: handle different types
		logger = logger.Str(k, fmt.Sprintf("%v", v))
	}
	logger.Msgs(msg, args...)
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
	GlobalLogger.LogWithFields(context.Background(), level, h.fields, msg)
}

func (h *legacyHandler) Logf(level Level, format string, args ...interface{}) {
	GlobalLogger.LogFWithFields(context.Background(), level, h.fields, format, args...)
}

func (l *logger) WithFields(fields map[string]any) *legacyHandler {
	return &legacyHandler{
		fields: fields,
	}
}
