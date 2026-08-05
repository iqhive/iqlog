//go:build !windows

package iqlog

import (
	"errors"
	"io"
	"log/syslog"
	"os"
	"strings"
)

func prepareSyslog(host, appName string) (io.Writer, error) {
	if host != "" && !strings.Contains(host, ":") {
		host += ":514"
	}
	return syslog.Dial("udp", host, syslog.LOG_DAEMON|syslog.LOG_INFO, appName)
}

// replaceWriter swaps the logger output under the logger mutex and closes
// the previous writer when it is one of our wrapper types, so its goroutine
// exits and queued lines are drained instead of leaking.
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
	switch ow := old.out.(type) {
	case *asyncWriter:
		return ow.Close()
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
func (l *Logger) setCallerDepth(d int) {
	if d < 0 {
		d = 0
	}
	l.updateConfig(func(cfg *loggerConfig) { cfg.callerDepth = d })
}
func (l *Logger) setUseColor(d bool) { l.updateConfig(func(cfg *loggerConfig) { cfg.color = d }) }
func (l *Logger) setIncludeTime(d bool) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.includeTime = d })
}
func (l *Logger) setTimestampLayout(layout string) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.timestampLayout = normalizeTimestampLayout(layout) })
}
func (l *Logger) setFormat(format Format) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.format = format })
}

func (l *Logger) setSyslogHostLegacy(newhost string) {
	_ = l.setSyslogHost(newhost)
}
func (l *Logger) setSyslogHost(newhost string) error {
	if newhost != "" && !strings.Contains(newhost, ":") {
		// make sure we have a (UDP) port in the host definition
		newhost = newhost + ":514"
	}
	cfg := l.config.Load()
	unchanged := cfg.syslogHost == newhost
	appName := cfg.applicationName
	if unchanged {
		// no change
		l.Debugf("Syslog host not changed to (%s) - already set to that", newhost)
		return nil
	}
	if newhost == "" {
		l.Info("Log output changed to StdErr")
		l.updateConfig(func(cfg *loggerConfig) { cfg.syslogHost = newhost })
		return l.replaceWriter(os.Stderr)
	}
	newSyslog, syslogErr := syslog.Dial("udp", newhost, syslog.LOG_DAEMON|syslog.LOG_INFO, appName)
	if syslogErr == nil && newSyslog != nil {
		l.updateConfig(func(cfg *loggerConfig) { cfg.syslogHost = newhost })
		return l.replaceWriter(newSyslog)
	} else {
		// keep the current writer rather than terminating the host process;
		// a logging library must not exit the application
		l.Errorf("ERROR: Unable to init syslog to (%s): %v", newhost, syslogErr)
		return syslogErr
	}
}
