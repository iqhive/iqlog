package iqlog

import (
	"context"
	"sort"
	"time"
)

// eventOrigin carries what an adapter already knows about a record: its call
// site and its timestamp. The default path passes nil and captures both
// itself.
type eventOrigin struct {
	// function and file are explicit caller text, as accepted by EventAt. A
	// non-empty function is required; a file alone is ignored. When function
	// is empty and pc is zero no caller is emitted, whatever CallerDepth says.
	function, file string
	// pc is a return address from runtime.Callers. It is resolved only when
	// the logger's CallerDepth is set, so an adapter that always has a PC
	// does not pay for symbolization unless callers are wanted.
	pc uintptr
	// when replaces the logger clock while explicitTime is set. A zero when
	// with explicitTime set omits the timestamp.
	when         time.Time
	explicitTime bool
	// fallback asks for a stack scan when the adapter has no PC to offer
	// (pc == 0) and function is empty. Without it a missing PC means the
	// adapter deliberately wants no caller, as for a hand-built slog.Record.
	fallback bool
}

// newEvent begins an event for level using the logger's own context.
func (l *Logger) newEvent(level Level) *Event {
	return l.newEventContext(l.ctx, level)
}

// newEventContext begins an event for level, capturing the caller by scanning
// the stack outward past the library boundary when CallerDepth is set.
func (l *Logger) newEventContext(ctx context.Context, level Level) *Event {
	return l.newEventContextAt(ctx, level, nil)
}

// newEventContextAt begins an event whose caller and timestamp may already be
// known to an adapter. A nil origin captures both from the current stack and
// clock.
func (l *Logger) newEventContextAt(ctx context.Context, level Level, origin *eventOrigin) *Event {
	// Fatal and Panic always produce an event: even when their record is
	// suppressed they still have to terminate.
	terminal := level == LevelFatal || level == LevelPanic
	gate := Level(l.level.Load())
	if gate > level && !terminal {
		// disabled fast path: no configuration load, no allocation
		return nil
	}
	cfg := l.config.Load()
	// Both gates are consulted because a copy or an in-flight reconfiguration
	// can briefly publish one before the other; a terminal level is never
	// dropped by either.
	if gate > level || cfg.level > level {
		if !terminal {
			return nil
		}
		e := acquireEvent(l, cfg)
		e.disabled = true
		e.level = level
		e.exitAfterWrite = level == LevelFatal
		e.panicAfterWrite = level == LevelPanic
		return e
	}
	e := acquireEvent(l, cfg)
	e.level = level
	e.jsonMode = cfg.format == FormatJSON
	e.includeTime = cfg.includeTime
	e.color = cfg.color
	e.captureCaller = cfg.callerDepth
	e.exitAfterWrite = level == LevelFatal
	e.panicAfterWrite = level == LevelPanic
	if origin != nil {
		e.captureCaller = 0
		if origin.function != "" {
			e.captureCaller = 1
			// explicit text overwrites whatever PC resolution was remembered
			e.callerData.pc = 0
			e.callerData.callerFuncLen = uint(copy(e.callerData.callerFunc[:], origin.function))
			e.callerData.callerFileLen = uint(copy(e.callerData.callerFile[:], origin.file))
		} else if origin.pc != 0 && cfg.callerDepth > 0 {
			e.captureCaller = 1
			captureCallerPC(origin.pc, cfg.callerDepth, cfg.callerPathMode, &e.callerData)
		} else if origin.fallback && cfg.callerDepth > 0 {
			e.captureCaller = 1
			captureCallerScan(cfg.callerDepth, cfg.callerPathMode, &e.callerData)
		}
	} else if cfg.callerDepth > 0 {
		captureCallerScan(cfg.callerDepth, cfg.callerPathMode, &e.callerData)
	}
	// The timestamp opens the record for every encoder, so it is written
	// before the level and caller. The clock is read as late as possible,
	// after the caller scan.
	if e.includeTime {
		if origin != nil && origin.explicitTime {
			if origin.when.IsZero() {
				e.includeTime = false
			} else {
				e.writeTimestamp(origin.when)
			}
		} else if cfg.fastClock {
			// One reading of the wall clock, written here rather than
			// through a helper so the stamp costs no extra frame. fastClock
			// guarantees the encoder is one of the two fixed-width
			// microsecond layouts, which is all the reading resolves.
			sec, usec := wallClock()
			if e.jsonMode {
				e.writeJSONTimestampMicros(sec, usec)
			} else {
				// time.Unix builds a Local time without allocating, so the
				// console encoder keeps the clock's own zone exactly as it
				// does for time.Now
				e.output = appendTimestamp(e.output, time.Unix(sec, usec*1000), cfg, false)
			}
		} else {
			e.writeTimestamp(cfg.now())
		}
	}
	if e.jsonMode {
		e.writeInitialJSON(level)
	} else {
		e.writeInitialConsole(level)
	}
	for _, field := range l.fields {
		addField(e, field.key, field.value)
	}
	if cfg.contextExtractor != nil && ctx != nil {
		fields := cfg.contextExtractor(ctx)
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			addField(e, key, fields[key])
		}
	}
	return e
}

// Event creates a structured event at level.
func (l *Logger) Event(level Level) *Event { return l.newEvent(normalizeLevel(level)) }

// EventAt creates a structured event with an explicit caller. A non-empty
// function is required; a file alone is ignored. An empty function suppresses
// caller output regardless of CallerDepth.
func (l *Logger) EventAt(level Level, function, file string) *Event {
	return l.newEventContextAt(l.ctx, normalizeLevel(level), &eventOrigin{function: function, file: file})
}

// The gated constructors below are shaped to stay within the compiler's
// inlining budget (TestGateConstructorsInline guards this): the level is
// compared as the raw int32 the atomic holds, and the enabled path calls
// newEventContextAt directly rather than through the newEvent wrappers, whose
// inlined bodies would push the cost over the limit. With the constructor
// inlined, a disabled call is a load and a compare in the caller's own frame
// with no call at all, and an enabled call reaches newEventContextAt one
// frame sooner. The result is identical to Level(l.level.Load()) > level.
func (l *Logger) TraceEvent() *Event {
	if l.level.Load() > int32(LevelTrace) {
		return nil
	}
	return l.newEventContextAt(l.ctx, LevelTrace, nil)
}
func (l *Logger) DebugEvent() *Event {
	if l.level.Load() > int32(LevelDebug) {
		return nil
	}
	return l.newEventContextAt(l.ctx, LevelDebug, nil)
}
func (l *Logger) InfoEvent() *Event {
	return l.newEvent(LevelInfo)
}
func (l *Logger) WarnEvent() *Event {
	if l.level.Load() > int32(LevelWarn) {
		return nil
	}
	return l.newEventContextAt(l.ctx, LevelWarn, nil)
}
func (l *Logger) ErrorEvent() *Event {
	if l.level.Load() > int32(LevelError) {
		return nil
	}
	return l.newEventContextAt(l.ctx, LevelError, nil)
}
func (l *Logger) PanicEvent() *Event { return l.newEvent(LevelPanic) }
func (l *Logger) FatalEvent() *Event { return l.newEvent(LevelFatal) }

func TraceEvent() *Event { return Default().TraceEvent() }
func DebugEvent() *Event { return Default().DebugEvent() }
func InfoEvent() *Event  { return Default().InfoEvent() }
func WarnEvent() *Event  { return Default().WarnEvent() }
func ErrorEvent() *Event { return Default().ErrorEvent() }
func PanicEvent() *Event { return Default().PanicEvent() }
func FatalEvent() *Event { return Default().FatalEvent() }

// writeTimestamp stamps the record with a time from a caller's clock or an
// adapter's record. The JSON UTC encoder keys its cache on the Unix second,
// which is the same instant whatever zone now carries. The default clock
// does not come through here: newEventContextAt reads it in place.
func (e *Event) writeTimestamp(now time.Time) {
	if e.jsonMode && e.config.jsonTimeMode == JSONTimeUTC {
		e.writeDefaultJSONTimestamp(now.Unix(), int64(now.Nanosecond()))
		return
	}
	e.output = appendTimestamp(e.output, now, e.config, e.jsonMode)
}

func (e *Event) writeInitialJSON(level Level) {
	if !e.includeTime {
		// with a timestamp the record was opened by the timestamp itself
		e.output = append(e.output, '{')
	}
	switch level {
	case LevelTrace:
		e.output = append(e.output, `"level":"TRACE"`...)
	case LevelDebug:
		e.output = append(e.output, `"level":"DEBUG"`...)
	case LevelInfo:
		e.output = append(e.output, `"level":"INFO"`...)
	case LevelWarn:
		e.output = append(e.output, `"level":"WARN"`...)
	case LevelError:
		e.output = append(e.output, `"level":"ERROR"`...)
	case LevelPanic:
		e.output = append(e.output, `"level":"PANIC"`...)
	case LevelFatal:
		e.output = append(e.output, `"level":"FATAL"`...)
	}
	// Guarded here as well as inside: addCallers is too large to inline, and
	// the call is dead weight for every record without caller capture.
	if e.captureCaller != 0 {
		e.addCallers()
	}
}

func (e *Event) writeInitialConsole(level Level) {
	if e.color {
		e.output = append(e.output, ansiColourPrefix(level)...)
	} else {
		e.output = append(e.output, levelPrefix(level)...)
	}
	if e.captureCaller != 0 {
		e.addCallers()
	}
	// everything appended from here on is caller-supplied
	e.prefixLen = len(e.output)
}

func (e *Event) addCallers() {
	if e.captureCaller == 0 || e.callerData.callerFuncLen == 0 {
		return
	}
	// CallerPathFile reports the file alone, so an explicit caller from
	// EventAt that has no file has nothing to show in that form.
	fileOnly := e.config.callerPathMode == CallerPathFile
	if fileOnly && e.callerData.callerFileLen == 0 {
		return
	}
	if e.jsonMode {
		if fileOnly {
			e.output = append(e.output, `,"file":"`...)
		} else {
			e.output = append(e.output, `,"func":"`...)
			e.output = appendJSONEscaped(e.output, unsafeString(e.callerData.callerFunc[:e.callerData.callerFuncLen]))
			e.output = append(e.output, `","file":"`...)
		}
		e.output = appendJSONEscaped(e.output, unsafeString(e.callerData.callerFile[:e.callerData.callerFileLen]))
		e.output = append(e.output, '"')
		return
	}
	if e.color {
		e.output = append(e.output, "\x1b[32m["...)
	} else {
		e.output = append(e.output, '[')
	}
	start := len(e.output)
	if !fileOnly {
		e.output = append(e.output, e.callerData.callerFunc[:e.callerData.callerFuncLen]...)
		e.output = append(e.output, ' ')
	}
	e.output = append(e.output, e.callerData.callerFile[:e.callerData.callerFileLen]...)
	// EventAt lets the caller supply these, and they land in the envelope,
	// which the whole-line scan does not cover. The colour sequences above
	// and below stay outside the escaped range.
	e.output = escapeConsoleTail(e.output, start)
	if e.color {
		e.output = append(e.output, "]\x1b[0m "...)
	} else {
		e.output = append(e.output, ']', ' ')
	}
}
