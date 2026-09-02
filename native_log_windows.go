//go:build windows

package iqlog

import (
	"sync"
	"unicode/utf8"

	"golang.org/x/sys/windows/svc/eventlog"
)

// maxEventLogMessageBytes bounds a record before UTF-16 conversion.
// ReportEvent limits each string to 31,839 UTF-16 code units and the whole
// record, header included, to about 60 KiB. 16 KiB of UTF-8 yields at most
// 16K code units (32 KiB), which stays well inside both limits.
const maxEventLogMessageBytes = 16 << 10

// eventLogWriter writes records to the Windows Event Log (Application log).
// ReportEvent is safe for concurrent use; mu only guards release racing
// writes.
type eventLogWriter struct {
	mu  sync.RWMutex
	log *eventlog.Log
}

// prepareNativeLog opens the Application event log with ApplicationName as
// the event source ("iqlog" when empty). Register the source once with
// InstallEventLogSource so records render without a "description cannot be
// found" notice.
func prepareNativeLog(cfg Config) (nativeLogWriter, error) {
	source := cfg.ApplicationName
	if source == "" {
		source = "iqlog"
	}
	log, err := eventlog.Open(source)
	if err != nil {
		return nil, err
	}
	return &eventLogWriter{log: log}, nil
}

// Write implements io.Writer. Records written through it are reported as
// warning events, the lowest severity the Event Log persists by default.
func (w *eventLogWriter) Write(p []byte) (int, error) {
	return w.writeLevel(LevelUnknown, p)
}

func (w *eventLogWriter) writeLevel(level Level, p []byte) (int, error) {
	msg := nativeMessage(p)
	if len(msg) > maxEventLogMessageBytes {
		msg = msg[:maxEventLogMessageBytes]
		// Do not split a multi-byte UTF-8 rune at the cut point.
		last := len(msg) - 1
		for last > 0 && !utf8.RuneStart(msg[last]) {
			last--
		}
		if !utf8.Valid(msg[last:]) {
			msg = msg[:last]
		}
	}
	w.mu.RLock()
	log := w.log
	var err error
	if log != nil {
		// Event ID 1 selects the generic "%1" message template provided by
		// EventCreate.exe; see InstallEventLogSource. Severity travels in
		// the event type.
		switch nativeSeverityFor(level) {
		case nativeSeverityNotice:
			err = log.Warning(1, unsafeString(msg))
		case nativeSeverityError, nativeSeverityFault:
			err = log.Error(1, unsafeString(msg))
		default:
			err = log.Info(1, unsafeString(msg))
		}
	}
	w.mu.RUnlock()
	if log == nil {
		return 0, ErrClosed
	}
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *eventLogWriter) release() {
	w.mu.Lock()
	if w.log != nil {
		_ = w.log.Close()
		w.log = nil
	}
	w.mu.Unlock()
}

// InstallEventLogSource registers source as an event source in the
// Application log using the generic EventCreate.exe message template, so
// records render cleanly in Event Viewer. It modifies the registry and
// requires administrator privileges; run it once at install time. Remove
// with RemoveEventLogSource.
func InstallEventLogSource(source string) error {
	return eventlog.InstallAsEventCreate(source, eventlog.Info|eventlog.Warning|eventlog.Error)
}

// RemoveEventLogSource removes a source registered by InstallEventLogSource.
func RemoveEventLogSource(source string) error {
	return eventlog.Remove(source)
}
