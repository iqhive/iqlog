//go:build !windows

package iqlog

import (
	"io"
	"log/syslog"
	"os"
)

func prepareSyslog(host, appName string) (io.Writer, error) {
	w, err := syslog.Dial("udp", syslogAddr(host), syslog.LOG_DAEMON|syslog.LOG_INFO, sanitizeSyslogTag(appName))
	if err != nil {
		return nil, err
	}
	// iqlog dialled this connection, so iqlog closes it when it is retired
	return newOwnedWriter(w), nil
}

// replaceWriter swaps the logger output under the logger mutex and closes
// the previous writer when it is one of our wrapper types, so its goroutine
// exits and queued lines are drained instead of leaking. mode is the writer
// mode the installed writer represents, recorded so a Config() round-trip
// reproduces it. Callers that built w (a queue or a dialled connection) retire
// it themselves when this returns an error.
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
func (l *Logger) setCallerDepth(d int) {
	if d < 0 {
		d = 0
	}
	l.updateConfig(func(cfg *loggerConfig) { cfg.callerDepth = d })
}

// setUseColor forces colour on or off. DisableColor still wins, as it does
// for Config.Color, so the snapshot never carries the contradictory pair the
// encoder would honour but a Config() round-trip would not.
func (l *Logger) setUseColor(d bool) {
	l.updateConfig(func(cfg *loggerConfig) {
		cfg.colorForced = d
		cfg.color = d && !cfg.disableColor
	})
}
func (l *Logger) setFormat(format Format) {
	l.updateConfig(func(cfg *loggerConfig) { cfg.format = format })
}

func (l *Logger) setSyslogHostLegacy(newhost string) {
	_ = l.setSyslogHost(newhost)
}
func (l *Logger) setSyslogHost(newhost string) error {
	newhost = syslogAddr(newhost)
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
		return l.replaceWriter(os.Stderr, WriterSync)
	}
	newSyslog, syslogErr := syslog.Dial("udp", newhost, syslog.LOG_DAEMON|syslog.LOG_INFO, sanitizeSyslogTag(appName))
	if syslogErr != nil || newSyslog == nil {
		// keep the current writer rather than terminating the host process;
		// a logging library must not exit the application
		l.Errorf("ERROR: Unable to init syslog to (%s): %v", newhost, syslogErr)
		return syslogErr
	}
	// iqlog dialled this connection, so iqlog closes it when it is retired,
	// including right here when the logger turns out to be closed already.
	owned := newOwnedWriter(newSyslog)
	if err := l.replaceWriter(owned, WriterSync); err != nil {
		_ = retireWriter(owned, nil)
		return err
	}
	// recorded after the install, which clears it, so a Config() round-trip
	// reproduces the syslog destination
	l.updateConfig(func(cfg *loggerConfig) { cfg.syslogHost = newhost })
	return nil
}
