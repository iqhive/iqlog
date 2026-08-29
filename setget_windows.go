//go:build windows

package iqlog

import (
	"errors"
	"io"
	"os"
)

func prepareSyslog(host, appName string) (io.Writer, error) { return os.Stderr, nil }

func (l *Logger) replaceWriter(w io.Writer) error {
	if nilWriter(w) {
		return errNilWriter
	}
	l.writer.mu.Lock()
	if l.writer.closed.Load() {
		l.writer.mu.Unlock()
		return ErrClosed
	}
	old := l.writer.active.Load()
	l.writer.active.Store(&outputState{out: w, configured: w})
	l.writer.mu.Unlock()
	if old.out == w {
		return nil
	}
	if async, ok := old.out.(*asyncWriter); ok {
		return async.Close()
	}
	return nil
}

func (l *Logger) setWriter(w io.Writer) error       { return l.replaceWriter(w) }
func (l *Logger) setWriterLegacy(w io.Writer) error { return l.setWriter(w) }
func (l *Logger) setAsyncWriter(w io.Writer, capacity int, policy OverflowPolicy) error {
	if nilWriter(w) {
		return errNilWriter
	}
	return l.replaceWriter(newAsyncWriter(w, capacity, policy, func(err error) {
		if errors.Is(err, ErrWriteDropped) {
			l.writer.dropped.Add(1)
		}
		l.recordWriteErr(err)
	}))
}
func (l *Logger) setAsyncWriterLegacy(w io.Writer) error {
	return l.setAsyncWriter(w, 1000, OverflowBlock)
}
func (l *Logger) setRingBufferWriter(w io.Writer, capacity int, policy OverflowPolicy) error {
	return l.setAsyncWriter(w, capacity, policy)
}
func (l *Logger) setRingBufferWriterLegacy(w io.Writer) error {
	return l.setRingBufferWriter(w, 10000, OverflowSync)
}
func (l *Logger) getWriter() io.Writer {
	return l.writer.active.Load().out
}
func (l *Logger) setApplicationName(name string) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.applicationName = name })
}
func (l *Logger) setCallerDepth(depth int) {
	if depth < 0 {
		depth = 0
	}
	l.updateConfig(func(cfg *loggerConfig) { cfg.callerDepth = depth })
}
func (l *Logger) setUseColor(enabled bool) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.color = enabled })
}
func (l *Logger) setIncludeTime(enabled bool) {
	l.updateConfig(func(cfg *loggerConfig) {
		cfg.includeTime = enabled
		if cfg.format == FormatJSON {
			if enabled {
				cfg.jsonTimeMode = JSONTimeUTC
			} else {
				cfg.jsonTimeMode = JSONTimeDisabled
			}
		}
	})
}
func (l *Logger) setTimestampLayout(layout string) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.timestampLayout = normalizeTimestampLayout(layout) })
}
func (l *Logger) setFormat(format Format) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.format = format })
}
func (l *Logger) setSyslogHostLegacy(host string) { _ = l.setSyslogHost(host) }
func (l *Logger) setSyslogHost(host string) error {
	l.updateConfig(func(cfg *loggerConfig) { cfg.syslogHost = host })
	if host != "" {
		l.Warnf("Syslog is not supported on Windows. Host %s will be ignored.", host)
	}
	return l.replaceWriter(os.Stderr)
}
