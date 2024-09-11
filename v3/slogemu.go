package iqlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type slogEmu struct {
	slog.Handler

	Logger *slog.Logger
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

func (l *slogEmu) Debug(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
}

func (l *slogEmu) DebugLn(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
}

func (l *slogEmu) Debugf(format string, args ...interface{}) {
	l.Log(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
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

// func (l *slogEmu) WithError(err error) Logger {
// 	l.Handler = l.Handler.WithAttrs([]slog.Attr{{
// 		Key:   "error",
// 		Value: slog.AnyValue(err),
// 	}})
// 	return l.
// }

func (l *slogEmu) SetApplicationName(name string) *slogEmu {
	return l
}
