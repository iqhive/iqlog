package iqlog

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

func Debugf(format string, args ...interface{}) {
	GlobalLogger.Debugf(format, args...)
}
func Debug(format string, args ...interface{}) {
	GlobalLogger.Debug(format, args...)
}
func Debugln(args ...interface{}) {
	GlobalLogger.Debugln(args...)
}
func (l *logger) Debugf(format string, args ...interface{}) {
	if !l.debug {
		return
	}
	l.logMessage(LevelDebug, format, args...)
	return
}
func (l *logger) Debug(format string, args ...interface{}) {
	if !l.debug {
		return
	}
	l.logMessage(LevelDebug, format, args...)
	return
}
func (l *logger) Debugln(args ...interface{}) {
	if !l.debug {
		return
	}
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelDebug,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	return
}

func Infof(format string, args ...interface{}) {
	GlobalLogger.Infof(format, args...)
}
func Info(format string, args ...interface{}) {
	GlobalLogger.Info(format, args...)
}
func Infoln(args ...interface{}) {
	GlobalLogger.Infoln(args...)
}
func (l *logger) Infof(format string, args ...interface{}) {
	l.logMessage(LevelInfo, format, args...)
	return
}

func (l *logger) Info(format string, args ...interface{}) {
	if !l.Enabled(l.ctx, LevelInfo) {
		return
	}
	if len(args) == 0 {
		record := LogRecord{
			Time:  time.Now(),
			Level: LevelInfo,
		}
		record.MsgLen = safeStringCopy(&record.Message, format)
		_ = l.WriteRecord(l.ctx, record)
	} else {
		l.logMessage(LevelInfo, format, args...)
	}
	return
}
func (l *logger) Infoln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelInfo,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	return
}

func Printf(format string, args ...interface{}) {
	GlobalLogger.Printf(format, args...)
}
func Print(format string, args ...interface{}) {
	GlobalLogger.Print(format, args...)
}
func Println(args ...interface{}) {
	GlobalLogger.Println(args...)
}

func (l *logger) Printf(format string, args ...interface{}) {
	l.logMessage(LevelInfo, format, args...)
	return
}
func (l *logger) Print(format string, args ...interface{}) {
	if len(args) == 0 {
		record := LogRecord{
			Time:  time.Now(),
			Level: LevelInfo,
		}
		record.MsgLen = safeStringCopy(&record.Message, format)
		_ = l.WriteRecordNoAlloc(record)
	} else {
		l.logMessage(LevelInfo, format, args...)
	}
	return
}
func (l *logger) Println(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelInfo,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	return
}

func Warnf(format string, args ...interface{}) {
	GlobalLogger.Warnf(format, args...)
}
func Warn(format string, args ...interface{}) {
	GlobalLogger.Warn(format, args...)
}
func Warnln(args ...interface{}) {
	GlobalLogger.Warnln(args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.logMessage(LevelWarn, format, args...)
	return
}
func (l *logger) Warn(format string, args ...interface{}) {
	if len(args) == 0 {
		record := LogRecord{
			Time:  time.Now(),
			Level: LevelWarn,
		}
		record.MsgLen = safeStringCopy(&record.Message, format)
		_ = l.WriteRecordNoAlloc(record)
	} else {
		l.logMessage(LevelWarn, format, args...)
	}
	return
}
func (l *logger) Warnln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelWarn,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	return
}

func Warningf(format string, args ...interface{}) {
	GlobalLogger.Warningf(format, args...)
}
func Warning(format string, args ...interface{}) {
	GlobalLogger.Warning(format, args...)
}
func Warningln(args ...interface{}) {
	GlobalLogger.Warningln(args...)
}

func (l *logger) Warningf(format string, args ...interface{}) {
	l.logMessage(LevelWarn, format, args...)
	return
}
func (l *logger) Warning(format string, args ...interface{}) {
	if len(args) == 0 {
		record := LogRecord{
			Time:  time.Now(),
			Level: LevelWarn,
		}
		record.MsgLen = safeStringCopy(&record.Message, format)
		_ = l.WriteRecordNoAlloc(record)
	} else {
		l.logMessage(LevelWarn, format, args...)
	}
	return
}
func (l *logger) Warningln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelWarn,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	return
}

func Errorf(format string, args ...interface{}) {
	GlobalLogger.Errorf(format, args...)
}
func Error(format string, args ...interface{}) {
	GlobalLogger.Error(format, args...)
}
func Errorln(args ...interface{}) {
	GlobalLogger.Errorln(args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.logMessage(LevelError, format, args...)
	return
}
func (l *logger) Error(format string, args ...interface{}) {
	if len(args) == 0 {
		record := LogRecord{
			Time:  time.Now(),
			Level: LevelError,
		}
		record.MsgLen = safeStringCopy(&record.Message, format)
		_ = l.WriteRecordNoAlloc(record)
	} else {
		l.logMessage(LevelError, format, args...)
	}
	return
}

func (l *logger) Errorln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelError,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	return
}

func Fatalf(format string, args ...interface{}) {
	GlobalLogger.Fatalf(format, args...)
}
func Fatal(format string, args ...interface{}) {
	GlobalLogger.Fatal(format, args...)
}
func Fatalln(args ...interface{}) {
	GlobalLogger.Fatalln(args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.logMessage(LevelFatal, format, args...)
	os.Exit(1)
	// unreachable:
	return
}
func (l *logger) Fatal(format string, args ...interface{}) {
	if len(args) == 0 {
		record := LogRecord{
			Time:  time.Now(),
			Level: LevelFatal,
		}
		record.MsgLen = safeStringCopy(&record.Message, format)
		_ = l.WriteRecordNoAlloc(record)
	} else {
		l.logMessage(LevelFatal, format, args...)
	}
	os.Exit(1)
	// unreachable:
	return
}
func (l *logger) Fatalln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelFatal,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	os.Exit(1)
	// unreachable:
	return
}

func Panicf(format string, args ...interface{}) {
	GlobalLogger.Panicf(format, args...)
}
func Panic(format string, args ...interface{}) {
	GlobalLogger.Panic(format, args...)
}
func Panicln(args ...interface{}) {
	GlobalLogger.Panicln(args...)
}
func (l *logger) Panicf(format string, args ...interface{}) {
	l.logMessage(LevelPanic, format, args...)
	panic(fmt.Sprintf(format, args...))
	// unreachable:
	return
}
func (l *logger) Panic(format string, args ...interface{}) {
	if len(args) == 0 {
		record := LogRecord{
			Time:  time.Now(),
			Level: LevelPanic,
		}
		record.MsgLen = safeStringCopy(&record.Message, format)
		_ = l.WriteRecordNoAlloc(record)
		panic(format)
	}
	msg := fmt.Sprintf(format, args...)
	l.logMessage(LevelPanic, msg)
	panic(msg)
	// unreachable:
	return
}
func (l *logger) Panicln(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelPanic,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	panic(msg)
	// unreachable:
	return
}

func Trace(args ...interface{}) {
	GlobalLogger.Trace(args...)
}
func Tracef(format string, args ...interface{}) {
	GlobalLogger.Tracef(format, args...)
}
func Traceln(args ...interface{}) {
	GlobalLogger.Traceln(args...)
}

func (l *logger) Tracef(format string, args ...interface{}) {
	l.logMessage(LevelTrace, format, args...)
	return
}
func (l *logger) Trace(args ...interface{}) {
	if !l.Enabled(l.ctx, LevelTrace) {
		return
	}
	msg := fmt.Sprint(args...)
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelTrace,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	return
}
func (l *logger) Traceln(args ...interface{}) {
	if !l.Enabled(l.ctx, LevelTrace) {
		return
	}
	msg := fmt.Sprintln(args...)
	msg = strings.TrimSuffix(msg, "\n")
	record := LogRecord{
		Time:  time.Now(),
		Level: LevelTrace,
	}
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
	return
}

// Log is a global helper / convenience function for accessing the GlobalLogger object
func Log(level Level, args ...interface{}) { GlobalLogger.Log(level, args...) }

// Logln is a global helper / convenience function for accessing the GlobalLogger object
func Logln(level Level, args ...interface{}) { GlobalLogger.Logln(level, args) }

func (l logger) Log(level Level, args ...interface{}) {
	l.slog.Log(l.ctx, slog.Level(level), fmt.Sprint(args...))
}

func (l logger) Logln(level Level, args ...interface{}) {
	l.slog.Log(l.ctx, slog.Level(level), fmt.Sprint(args...))
}
