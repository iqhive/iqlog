//go:build windows

package iqlog

import (
	"io"
	"os"
)

func prepareSyslog(host, appName string) (io.Writer, error) { return os.Stderr, nil }

// replaceWriter swaps the logger output under the logger mutex and closes
// the previous writer when it is one of our wrapper types, so its goroutine
// exits and queued lines are drained instead of leaking. mode is the writer
// mode the installed writer represents, recorded so a Config() round-trip
// reproduces it. Callers that built w (a queue) retire it themselves when
// this returns an error.
func (l *Logger) replaceWriter(w io.Writer, mode WriterMode) error {
	if nilWriter(w) {
		return errNilWriter
	}
	l.writer.mu.Lock()
	if l.writer.closed.Load() {
		l.writer.mu.Unlock()
		return ErrClosed
	}
	old := l.writer.active.Load()
	l.writer.active.Store(l.newOutputState(w))
	// Every writer replacement stops native logging and syslog: the new state
	// carries an explicit destination, so the flags must be cleared even when
	// the writer itself is unchanged. Done under writer.mu so the snapshot
	// cannot be observed disagreeing with the routing.
	l.installedDestination(w, mode)
	var err error
	if !sameWriter(old.out, w) {
		// drained under writer.mu so a concurrent Flush waits for it
		err = retireWriter(old.out, w)
	}
	l.writer.mu.Unlock()
	return err
}

func (l *Logger) setWriter(w io.Writer) error { return l.replaceWriter(w, WriterSync) }

// installAsyncWriter queues w behind a fresh async writer. The queue is
// retired again when it cannot be installed, so a closed logger does not
// leak its goroutine.
func (l *Logger) installAsyncWriter(w io.Writer, capacity int, policy OverflowPolicy, mode WriterMode) error {
	if nilWriter(w) {
		return errNilWriter
	}
	aw := newAsyncWriter(w, capacity, policy, l.queueError, &l.writer.writeMu)
	if err := l.replaceWriter(aw, mode); err != nil {
		_ = retireWriter(aw, unwrapAsync(w))
		return err
	}
	return nil
}

func (l *Logger) setAsyncWriter(w io.Writer, capacity int, policy OverflowPolicy) error {
	return l.installAsyncWriter(w, capacity, policy, WriterAsync)
}
func (l *Logger) setAsyncWriterLegacy(w io.Writer) error {
	return l.setAsyncWriter(w, 1000, OverflowBlock)
}
func (l *Logger) setRingBufferWriter(w io.Writer, capacity int, policy OverflowPolicy) error {
	return l.installAsyncWriter(w, capacity, policy, WriterRing)
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

// setUseColor forces colour on or off. DisableColor still wins, as it does
// for Config.Color, so the snapshot never carries the contradictory pair the
// encoder would honour but a Config() round-trip would not.
func (l *Logger) setUseColor(enabled bool) {
	l.updateConfig(func(cfg *loggerConfig) {
		cfg.colorForced = enabled
		cfg.color = enabled && !cfg.disableColor
	})
}
func (l *Logger) setFormat(format Format) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.format = format })
}
func (l *Logger) setSyslogHostLegacy(host string) { _ = l.setSyslogHost(host) }
func (l *Logger) setSyslogHost(host string) error {
	if l.config.Load().syslogHost == host {
		// no change: keep the current writer, as the unix implementation does
		return nil
	}
	if host != "" {
		l.Warnf("Syslog is not supported on Windows. Host %s will be ignored.", host)
	}
	if err := l.replaceWriter(os.Stderr, WriterSync); err != nil {
		return err
	}
	l.updateConfig(func(cfg *loggerConfig) { cfg.syslogHost = host })
	return nil
}
