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
//   - The record's program counter is resolved into "func" and "file" only
//     when Config.CallerDepth is set, exactly as for native events. A record
//     without a program counter has no caller.
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
// See Logger.SlogHandler for the mapping rules.
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

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.logger().Enabled(slogLevelToIQ(level))
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
	level := slogLevelToIQ(r.Level)
	if !l.Enabled(level) {
		return nil
	}
	origin := eventOrigin{pc: r.PC, when: r.Time, explicitTime: true}
	e := l.newEventContextAt(ctx, level, 1, &origin)
	if e == nil {
		return nil
	}
	if e.jsonMode {
		e.output = append(e.output, h.preJSON...)
	} else {
		e.output = append(e.output, h.preConsole...)
	}
	r.Attrs(func(a slog.Attr) bool {
		writeSlogAttr(e, h.prefix, h.prefixEscape, a)
		return true
	})
	err := e.buildErr
	e.Msg(r.Message)
	if err == nil {
		err = h.preErr
	}
	return err
}

// writeSlogAttr appends one attribute under prefix, flattening groups into
// dotted keys. Values are resolved first so a LogValuer that yields a group
// is flattened like any other group.
func writeSlogAttr(e *Event, prefix string, prefixEscape bool, a slog.Attr) {
	v := a.Value.Resolve()
	if v.Kind() == slog.KindGroup {
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
	e.appendSlogKey(prefix, prefixEscape, a.Key)
	appendSlogValue(e, a.Key, v)
}

// appendSlogKey writes a field key made of the group prefix and name. A
// dotted key can never collide with the record envelope, so the reserved
// name check runs only when there is no prefix.
func (e *Event) appendSlogKey(prefix string, prefixEscape bool, name string) {
	if prefix == "" {
		name = eventFieldName(name)
	}
	if !e.jsonMode {
		e.output = append(e.output, prefix...)
		e.output = append(e.output, name...)
		e.output = append(e.output, '=')
		return
	}
	e.output = append(e.output, ',', '"')
	if e.config.escapeFieldNames || prefixEscape || jsonKeyNeedsEscaping(name) {
		e.output = appendJSONEscaped(e.output, prefix)
		e.output = appendJSONEscaped(e.output, name)
	} else {
		e.output = append(e.output, prefix...)
		e.output = append(e.output, name...)
	}
	e.output = append(e.output, '"', ':')
}

// appendSlogValue writes a resolved, non-group value after its key, using the
// same encodings as the typed event methods. key only names the field in a
// build error.
func appendSlogValue(e *Event, key string, v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		e.appendSlogString(v.String())
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
		e.appendSlogString(value.Error())
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
	if e.jsonMode {
		e.output = append(e.output, '"')
		e.output = appendJSONEscaped(e.output, s)
		e.output = append(e.output, '"')
		return
	}
	e.output = append(e.output, s...)
	e.output = append(e.output, ' ')
}

// endSlogValue closes a console value; JSON values need nothing.
func (e *Event) endSlogValue() {
	if !e.jsonMode {
		e.output = append(e.output, ' ')
	}
}
