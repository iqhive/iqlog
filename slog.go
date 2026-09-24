package iqlog

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"
)

var _ slog.Handler = (*slogHandler)(nil)

// SlogHandler returns a slog.Handler that writes through l. Use it to route
// libraries, or the standard log package after slog.SetDefault, into an
// iqlog logger:
//
//	slog.SetDefault(slog.New(log.SlogHandler()))
//
// Format, colour, level, timestamps, caller capture, context extraction,
// persistent fields, and the writer all come from l, so there are no handler
// options. The mapping is:
//
//   - slog levels below Debug become Trace; Debug up to Info become Debug;
//     Info up to Warn become Info; Warn up to Error become Warn; Error and
//     above become Error. No slog level terminates the process.
//   - The record message is written under "message", and groups become
//     dotted keys ("http.method"). Attributes named like the record envelope
//     (time, level, message, func, file) are prefixed with "field_" as for
//     any other event.
//   - The record's own timestamp is used, and only when the logger emits
//     timestamps (IncludeTime for console, JSONTimeMode for JSON). A record
//     with a zero time has no timestamp.
//   - Caller capture happens only when Config.CallerDepth is set, and that is
//     the only case that pays for it. When the record carries a program
//     counter, that call site (the direct caller of the slog call) is reported;
//     a nonzero program counter is reported as its own call site and does not
//     honour CallerDepth values above 1. When PC is zero, the handler scans the
//     live stack, skipping the Go runtime and the log and log/slog front ends,
//     so the caller is reported when it falls within the bounded scan window.
//     Caller output never includes frames
//     from iqlog or its legacy package family. A call site's resolved caller
//     is remembered on the pooled event, so a hot call site pays for the
//     symbol lookup once rather than on every record.
//   - LogValuer values are resolved, in WithAttrs as well as in records.
//     Attributes with an empty key and an empty value are dropped.
//
// Attributes added with WithAttrs are encoded once, so they cost one copy per
// record. Handle does not allocate for records made of typed attributes; the
// remaining cost is slog's own record construction. Application hot paths
// should still call the typed event methods directly.
//
// Handle returns nil after a closed logger, and returns the event's build
// error when an attribute carried invalid raw JSON.
func (l *Logger) SlogHandler() slog.Handler {
	if l == nil {
		panic("iqlog: SlogHandler: nil logger")
	}
	return &slogHandler{log: l}
}

// SlogHandler returns a slog.Handler bound to the package-level logger. It
// consults Default on every call, so a later SetDefault is honoured, unlike
// Default().SlogHandler(), which binds the logger installed at that moment.
//
// Caller capture happens only when Config.CallerDepth is set, and that is the
// only case that pays for it. A record with a program counter reports that call
// site, the direct caller of the slog call; the nonzero-PC path does not honour
// CallerDepth values above 1. A record with PC zero falls back to scanning the
// live stack while skipping the Go runtime and the log and log/slog front ends;
// the bounded scan may report no caller beyond its frame window.
// Caller output never includes iqlog or legacy iqlog-family frames. See
// Logger.SlogHandler for the other mapping rules.
func SlogHandler() slog.Handler { return &slogHandler{} }

type slogHandler struct {
	log *Logger // nil: Default() at call time
	// prefix is the open group path with a trailing dot ("http.req."), or
	// empty. prefixEscape records whether it needs JSON escaping so the key
	// scan covers only the attribute name per record.
	prefix       string
	prefixEscape bool
	// preJSON and preConsole hold the attributes added with WithAttrs,
	// encoded once in each format so Handle appends them with one copy. Each
	// WithAttrs builds fresh slices, so handlers never share spare capacity.
	preJSON    []byte
	preConsole []byte
	// preErr is the first build error met while pre-encoding; Handle reports
	// it when the record itself was clean.
	preErr error
}

func (h *slogHandler) logger() *Logger {
	if h.log != nil {
		return h.log
	}
	return Default()
}

func slogLevelToIQ(l slog.Level) Level {
	switch {
	case l < slog.LevelDebug:
		return LevelTrace
	case l < slog.LevelInfo:
		return LevelDebug
	case l < slog.LevelWarn:
		return LevelInfo
	case l < slog.LevelError:
		return LevelWarn
	default:
		return LevelError
	}
}

// Enabled is called by slog before every record, so it is kept free of calls:
// with nothing to call, the compiler emits it as a leaf without a stack check
// or frame. That is why the unbound handler reads the default logger directly
// rather than through Default(): when nothing has built the default yet,
// Default() would build it at the default minimum level and open, so the
// answer is known without building it here. Handle builds it on the first
// record that is enabled.
func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	l := h.log
	if l == nil {
		if l = defaultLogger.Load(); l == nil {
			return defaultLevel <= slogLevelToIQ(level)
		}
	}
	return l.Enabled(slogLevelToIQ(level))
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	c := *h
	c.prefix = h.prefix + name + "."
	c.prefixEscape = h.prefixEscape || jsonKeyNeedsEscaping(name)
	return &c
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	l := h.logger()
	cfg := l.config.Load()
	c := *h
	var err error
	c.preJSON, err = preEncodeSlogAttrs(l, cfg, true, h.preJSON, h.prefix, h.prefixEscape, attrs)
	if err != nil && c.preErr == nil {
		c.preErr = err
	}
	c.preConsole, err = preEncodeSlogAttrs(l, cfg, false, h.preConsole, h.prefix, h.prefixEscape, attrs)
	if err != nil && c.preErr == nil {
		c.preErr = err
	}
	return &c
}

// preEncodeSlogAttrs encodes attrs in one format through a scratch event
// and returns a fresh slice holding existing followed by the new bytes.
func preEncodeSlogAttrs(l *Logger, cfg *loggerConfig, jsonMode bool, existing []byte, prefix string, prefixEscape bool, attrs []slog.Attr) ([]byte, error) {
	e := acquireEvent(l, cfg)
	e.jsonMode = jsonMode
	for _, a := range attrs {
		writeSlogAttr(e, prefix, prefixEscape, a)
	}
	out := make([]byte, 0, len(existing)+len(e.output))
	out = append(out, existing...)
	out = append(out, e.output...)
	err := e.buildErr
	e.release(true)
	return out, err
}

func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	l := h.logger()
	// Enabled is not consulted here: its level test is the gate
	// newEventContextAt applies anyway, and paying it twice per record was
	// measurable. Only the closed check it adds is kept.
	if l.writer.closed.Load() {
		return nil
	}
	// Assigned field by field: a composite literal whose address is taken
	// is built in a temporary and copied into place.
	var origin eventOrigin
	origin.pc = r.PC
	origin.when = r.Time
	origin.explicitTime = true
	origin.fallback = true
	e := l.newEventContextAt(ctx, slogLevelToIQ(r.Level), &origin)
	if e == nil {
		return nil
	}
	if e.jsonMode {
		if len(h.preJSON) != 0 {
			e.output = append(e.output, h.preJSON...)
		}
	} else if len(h.preConsole) != 0 {
		e.output = append(e.output, h.preConsole...)
	}
	// Captured as values so the inlined closure does not reload them from
	// the handler on every attribute.
	prefix, prefixEscape := h.prefix, h.prefixEscape
	r.Attrs(func(a slog.Attr) bool {
		writeSlogAttr(e, prefix, prefixEscape, a)
		return true
	})
	err := e.buildErr
	// The record is finished directly rather than through Msg: no slog
	// level maps to Fatal or Panic, so the event is never a disabled
	// terminal one, and it is fresh, so it cannot have been consumed. Those
	// are the only checks Msg adds.
	if e.jsonMode {
		e.writeFinalJSON(r.Message)
	} else {
		e.writeFinalConsole(r.Message)
	}
	if err == nil {
		err = h.preErr
	}
	return err
}

// writeSlogAttr appends one attribute under prefix, flattening groups into
// dotted keys. A LogValuer is resolved first so one that yields a group is
// flattened like any other group.
//
// Value.Kind is a type switch and Value.Resolve installs a deferred recover
// for a panicking LogValue, so the kind is taken once and Resolve runs only
// for the values that need it; every other kind goes straight to its encoder.
func writeSlogAttr(e *Event, prefix string, prefixEscape bool, a slog.Attr) {
	v := a.Value
	kind := v.Kind()
	if kind == slog.KindLogValuer {
		v = v.Resolve()
		kind = v.Kind()
	}
	if kind == slog.KindGroup {
		attrs := v.Group()
		if len(attrs) == 0 {
			return
		}
		if a.Key != "" {
			// A group inside a record extends the prefix for its members
			// only. The joined path lives in a stack buffer unless it
			// outgrows it; nothing retains it past this call.
			var buf [96]byte
			joined := append(buf[:0], prefix...)
			joined = append(joined, a.Key...)
			joined = append(joined, '.')
			prefix = unsafeString(joined)
			prefixEscape = prefixEscape || jsonKeyNeedsEscaping(a.Key)
		}
		for _, ga := range attrs {
			writeSlogAttr(e, prefix, prefixEscape, ga)
		}
		return
	}
	if a.Key == "" && v.Equal(slog.Value{}) {
		return
	}

	// The key is the group prefix and the attribute name. A dotted key can
	// never collide with the record envelope, so the reserved name check
	// runs only when there is no prefix. It is written here rather than by
	// a helper to keep one call off every attribute; the slice header is
	// kept in a local across the appends for the same reason, since through
	// e it would be stored and reloaded around every call in between.
	name := a.Key
	if prefix == "" {
		name = eventFieldName(name)
	}
	out := e.output
	if !e.jsonMode {
		out = append(out, prefix...)
		out = append(out, name...)
		e.output = append(out, '=')
	} else {
		out = append(out, ',', '"')
		if e.config.escapeFieldNames || prefixEscape || jsonKeyNeedsEscaping(name) {
			out = appendJSONEscaped(out, prefix)
			out = appendJSONEscaped(out, name)
		} else {
			// Most records have no group, and an append of nothing still
			// costs a memmove call.
			if prefix != "" {
				out = append(out, prefix...)
			}
			out = append(out, name...)
		}
		e.output = append(out, '"', ':')
	}
	appendSlogValue(e, a.Key, kind, v)
}

// appendSlogValue writes a resolved, non-group value of the given kind after
// its key, using the same encodings as the typed event methods. key only
// names the field in a build error.
func appendSlogValue(e *Event, key string, kind slog.Kind, v slog.Value) {
	switch kind {
	case slog.KindString:
		// Written in place rather than through appendSlogString: strings
		// are the commonest kind, and the helper is just over the inlining
		// budget.
		s := v.String()
		out := e.output
		if e.jsonMode {
			out = append(out, '"')
			out = appendJSONEscaped(out, s)
			e.output = append(out, '"')
		} else {
			out = append(out, s...)
			e.output = append(out, ' ')
		}
	case slog.KindInt64:
		e.output = strconv.AppendInt(e.output, v.Int64(), 10)
		e.endSlogValue()
	case slog.KindUint64:
		e.output = strconv.AppendUint(e.output, v.Uint64(), 10)
		e.endSlogValue()
	case slog.KindFloat64:
		if e.jsonMode {
			e.output = appendJSONFloat(e.output, v.Float64(), 64)
		} else {
			e.output = strconv.AppendFloat(e.output, v.Float64(), 'f', -1, 64)
		}
		e.endSlogValue()
	case slog.KindBool:
		if v.Bool() {
			e.output = append(e.output, "true"...)
		} else {
			e.output = append(e.output, "false"...)
		}
		e.endSlogValue()
	case slog.KindDuration:
		e.appendSlogString(v.Duration().String())
	case slog.KindTime:
		if e.jsonMode {
			e.output = append(e.output, '"')
			e.output = v.Time().AppendFormat(e.output, time.RFC3339Nano)
			e.output = append(e.output, '"')
		} else {
			e.output = v.Time().AppendFormat(e.output, time.RFC3339Nano)
		}
		e.endSlogValue()
	default:
		appendSlogAny(e, key, v.Any())
	}
}

// appendSlogAny mirrors the concrete cases of Event.Any that slog does not
// already turn into typed kinds, then falls back to the generic encoder.
func appendSlogAny(e *Event, key string, v any) {
	switch value := v.(type) {
	case nil:
		e.output = append(e.output, "null"...)
		e.endSlogValue()
	case error:
		if isNilPointer(value) {
			e.appendGenericValue(v)
		} else {
			e.appendSlogString(value.Error())
		}
	case json.RawMessage:
		value = e.checkRawJSON(key, value)
		e.output = append(e.output, value...)
		e.endSlogValue()
	case []byte:
		if e.jsonMode {
			e.output = append(e.output, '"')
			e.output = base64.StdEncoding.AppendEncode(e.output, value)
			e.output = append(e.output, '"')
		} else {
			e.output = base64.StdEncoding.AppendEncode(e.output, value)
		}
		e.endSlogValue()
	default:
		e.appendGenericValue(v)
	}
}

func (e *Event) appendSlogString(s string) {
	out := e.output
	if e.jsonMode {
		out = append(out, '"')
		out = appendJSONEscaped(out, s)
		e.output = append(out, '"')
		return
	}
	out = append(out, s...)
	e.output = append(out, ' ')
}

// endSlogValue closes a console value; JSON values need nothing.
func (e *Event) endSlogValue() {
	if !e.jsonMode {
		e.output = append(e.output, ' ')
	}
}
