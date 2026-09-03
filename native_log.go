package iqlog

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"sync"
)

var (
	// errNativeLogUnsupported is returned when Config.NativeLog is set on a
	// platform without a native system log writer.
	errNativeLogUnsupported = errors.New("iqlog: native system log is not supported on this platform")
	// errNativeLogUnavailable is returned when the platform native system log
	// cannot be opened.
	errNativeLogUnavailable = errors.New("iqlog: native system log is unavailable")
	// errNativeLogCGO is returned on macOS builds without cgo; os_log is a C
	// API and cannot be reached with CGO_ENABLED=0.
	errNativeLogCGO = errors.New("iqlog: native system log on macOS requires cgo (CGO_ENABLED=1)")
)

// nativeLogWriter is implemented by the platform-native system log writers
// (os_log on macOS, the Event Log on Windows). It is level-aware so records
// keep their severity in the system log, and releasable so writer
// replacement and Close free the platform handle.
type nativeLogWriter interface {
	io.Writer
	writeLevel(level Level, p []byte) (int, error)
	release()
}

// ownedWriter wraps a writer this package opened itself, currently the syslog
// connection dialled by prepareSyslog and setSyslogHost. Caller-supplied
// writers must never be closed by the logger, so ownership is recorded here
// rather than inferred from an io.Closer assertion: retireWriter closes only
// what iqlog opened.
type ownedWriter struct {
	io.Writer
	closer io.Closer
	once   sync.Once
}

func newOwnedWriter(w io.Writer) io.Writer {
	closer, ok := w.(io.Closer)
	if !ok {
		return w
	}
	return &ownedWriter{Writer: w, closer: closer}
}

// release satisfies the releasable interface retireWriter looks for.
func (w *ownedWriter) release() {
	w.once.Do(func() { _ = w.closer.Close() })
}

// sameWriter reports whether a and b are the same writer.
//
// Comparing two io.Writer values directly panics when their dynamic type is
// identical and not comparable, which a caller-supplied writer may well be:
// any struct with a slice, map, or func field qualifies. Writer identity is
// only ever used to skip work that is harmless to repeat, so a writer whose
// type cannot be compared is simply reported as different.
func sameWriter(a, b io.Writer) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ta := reflect.TypeOf(a)
	if ta != reflect.TypeOf(b) || !ta.Comparable() {
		return false
	}
	return a == b
}

// retireWriter shuts down a writer that is no longer active: async writers
// are drained and stopped, and native system log handles and connections
// iqlog opened itself are released unless keep still uses the same
// underlying writer. That reuse happens legitimately when Config().Writer or
// GetWriter() is fed back into SetConfig or the Set*Writer methods, and
// releasing the handle then would leave the logger silently dead.
//
// Writer identity goes through sameWriter because keep may be a
// caller-supplied value whose type cannot be compared.
func retireWriter(old, keep io.Writer) error {
	inner := old
	var err error
	if async, ok := old.(*asyncWriter); ok {
		err = async.Close()
		inner = async.out
	}
	if async, ok := keep.(*asyncWriter); ok {
		keep = async.out
	}
	if native, ok := inner.(interface{ release() }); ok && !sameWriter(inner, keep) {
		native.release()
	}
	return err
}

// clearNativeLog drops the NativeLog flag from the configuration snapshot
// after a writer method installed an explicit destination. Without this a
// Config() round-trip through SetConfig would silently re-enable the system
// log the caller had just replaced.
func (l *Logger) clearNativeLog() {
	if l.config.Load().nativeLog {
		l.updateConfig(func(cfg *loggerConfig) { cfg.nativeLog = false })
	}
}

// nativeSeverity abstracts platform-native severity levels so the level
// mapping is testable on every platform.
type nativeSeverity uint8

const (
	nativeSeverityDebug nativeSeverity = iota
	nativeSeverityInfo
	nativeSeverityNotice
	nativeSeverityError
	nativeSeverityFault
)

// nativeSeverityFor maps iqlog levels onto native severities. Unknown levels
// (including writes through the plain io.Writer interface) map to notice,
// the lowest severity both platforms persist by default.
func nativeSeverityFor(level Level) nativeSeverity {
	switch level {
	case LevelTrace, LevelDebug:
		return nativeSeverityDebug
	case LevelInfo:
		return nativeSeverityInfo
	case LevelWarn:
		return nativeSeverityNotice
	case LevelError:
		return nativeSeverityError
	case LevelPanic, LevelFatal:
		return nativeSeverityFault
	default:
		return nativeSeverityNotice
	}
}

// sanitizeSyslogTag replaces control bytes in a syslog tag. The tag is
// written into the record header ahead of the message, so a newline in it
// splits one datagram into what a line-oriented collector reads as two
// records. ApplicationName often comes from a config file or the environment,
// so it is not trusted to be a bare identifier.
func sanitizeSyslogTag(tag string) string {
	needs := false
	for i := 0; i < len(tag); i++ {
		if c := tag[i]; c < 0x20 || c == 0x7f {
			needs = true
			break
		}
	}
	if !needs {
		return tag
	}
	out := []byte(tag)
	for i, c := range out {
		if c < 0x20 || c == 0x7f {
			out[i] = '_'
		}
	}
	return string(out)
}

// trimRecordNewline drops the single trailing newline the record encoders
// append; system log APIs frame records themselves.
func trimRecordNewline(p []byte) []byte {
	if n := len(p); n > 0 && p[n-1] == '\n' {
		return p[:n-1]
	}
	return p
}

// nativeMessage prepares a record for a system log API: it trims the
// trailing newline and escapes NUL bytes, which would otherwise truncate the
// C string handed to os_log or make the Event Log UTF-16 conversion reject
// the whole record. The console sanitizer only escapes CR and LF, so NUL can
// reach this point through logged values. Returns p's storage unchanged in
// the common NUL-free case.
func nativeMessage(p []byte) []byte {
	msg := trimRecordNewline(p)
	if bytes.IndexByte(msg, 0) < 0 {
		return msg
	}
	out := make([]byte, 0, len(msg)+8)
	for _, b := range msg {
		if b == 0 {
			out = append(out, '\\', '0')
		} else {
			out = append(out, b)
		}
	}
	return out
}
