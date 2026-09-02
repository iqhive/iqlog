//go:build darwin && cgo

package iqlog

/*
#include <stdint.h>
#include <stdlib.h>
#include <os/log.h>
#include <os/object.h>

static os_log_t iqlog_os_log_create(const char *subsystem, const char *category) {
	return os_log_create(subsystem, category);
}

// The type arrives as a plain uint8_t so the Go side does not depend on cgo
// enum constant generation. "%{public}s" keeps the message visible in
// Console.app and log stream instead of redacted as <private>.
static void iqlog_os_log_write(os_log_t log, uint8_t type, const char *msg) {
	os_log_with_type(log, (os_log_type_t)type, "%{public}s", msg);
}

static void iqlog_os_log_release(os_log_t log) {
	os_release(log);
}
*/
import "C"

import (
	"sync"
	"unsafe"
)

// os_log_type_t values from <os/log.h>, mirrored as plain constants for the
// reason described above.
const (
	osLogTypeDefault = 0
	osLogTypeInfo    = 1
	osLogTypeDebug   = 2
	osLogTypeError   = 3
	osLogTypeFault   = 4
)

func (s nativeSeverity) osLogType() uint8 {
	switch s {
	case nativeSeverityDebug:
		return osLogTypeDebug
	case nativeSeverityInfo:
		return osLogTypeInfo
	case nativeSeverityError:
		return osLogTypeError
	case nativeSeverityFault:
		return osLogTypeFault
	default:
		return osLogTypeDefault
	}
}

// osLogWriter writes records to Apple's unified logging system (os_log).
// os_log is safe for concurrent use; mu only guards release racing writes.
type osLogWriter struct {
	mu     sync.RWMutex
	handle C.os_log_t
}

// prepareNativeLog creates an os_log writer. ApplicationName becomes the
// os_log subsystem ("iqlog" when empty), so records are filterable with
// `log stream --predicate 'subsystem == "<ApplicationName>"'`.
func prepareNativeLog(cfg Config) (nativeLogWriter, error) {
	subsystem := cfg.ApplicationName
	if subsystem == "" {
		subsystem = "iqlog"
	}
	// os_log_create copies both strings.
	cSubsystem := C.CString(subsystem)
	cCategory := C.CString("default")
	handle := C.iqlog_os_log_create(cSubsystem, cCategory)
	C.free(unsafe.Pointer(cSubsystem))
	C.free(unsafe.Pointer(cCategory))
	if handle == nil {
		return nil, errNativeLogUnavailable
	}
	return &osLogWriter{handle: handle}, nil
}

// Write implements io.Writer. Records written through it land at the
// default (notice) os_log type.
func (w *osLogWriter) Write(p []byte) (int, error) {
	return w.writeLevel(LevelUnknown, p)
}

func (w *osLogWriter) writeLevel(level Level, p []byte) (int, error) {
	msg := nativeMessage(p)
	// C.CString copies into C memory and NUL-terminates; os_log copies the
	// message into its own buffer during the call, so cmsg can be freed as
	// soon as it returns.
	cmsg := C.CString(unsafeString(msg))
	w.mu.RLock()
	handle := w.handle
	if handle != nil {
		C.iqlog_os_log_write(handle, C.uint8_t(nativeSeverityFor(level).osLogType()), cmsg)
	}
	w.mu.RUnlock()
	C.free(unsafe.Pointer(cmsg))
	if handle == nil {
		return 0, ErrClosed
	}
	return len(p), nil
}

func (w *osLogWriter) release() {
	w.mu.Lock()
	if w.handle != nil {
		C.iqlog_os_log_release(w.handle)
		w.handle = nil
	}
	w.mu.Unlock()
}
