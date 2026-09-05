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
	// function and file are explicit caller text, as accepted by EventAt.
	// When function is empty and pc is zero no caller is emitted, whatever
	// CallerDepth says.
	function, file string
	// pc is a return address from runtime.Callers. It is resolved only when
	// the logger's CallerDepth is set, so an adapter that always has a PC
	// does not pay for symbolization unless callers are wanted.
	pc uintptr
	// when replaces the logger clock while explicitTime is set. A zero when
	// with explicitTime set omits the timestamp.
	when         time.Time
	explicitTime bool
}

func (l *Logger) newEvent(level Level, callerSkip int) *Event {
	return l.newEventContext(l.ctx, level, callerSkip)
}

func (l *Logger) newEventContext(ctx context.Context, level Level, callerSkip int) *Event {
	return l.newEventContextAt(ctx, level, callerSkip+1, nil)
}

func (l *Logger) newEventContextAt(ctx context.Context, level Level, callerSkip int, origin *eventOrigin) *Event {
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
	// Writing into callerData invalidates the timestamp cache that shares
	// its buffer, so timeSecond is reset on exactly the paths that write.
	if origin != nil {
		e.captureCaller = 0
		if origin.function != "" {
			e.timeSecond = invalidTimestampSecond
			e.captureCaller = 1
			e.callerData.callerFuncLen = uint(copy(e.callerData.callerFunc[:], origin.function))
			e.callerData.callerFileLen = uint(copy(e.callerData.callerFile[:], origin.file))
		} else if origin.pc != 0 && cfg.callerDepth > 0 {
			e.timeSecond = invalidTimestampSecond
			e.captureCaller = 1
			captureCallerPC(origin.pc, &e.callerData)
		}
	} else if cfg.callerDepth > 0 {
		e.timeSecond = invalidTimestampSecond
		captureCaller(cfg.callerDepth+callerSkip+1, &e.callerData)
	}
	var now time.Time
	if e.includeTime {
		if origin != nil && origin.explicitTime {
			if origin.when.IsZero() {
				e.includeTime = false
			} else {
				now = origin.when
			}
		} else {
			now = cfg.now()
		}
	}
	if e.jsonMode {
		e.writeInitialJSON(level, now)
	} else {
		e.writeInitialConsole(level, now)
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
func (l *Logger) Event(level Level) *Event { return l.newEvent(normalizeLevel(level), 1) }

// EventAt creates a structured event with an explicit caller. Empty function
// and file values suppress caller output regardless of CallerDepth.
func (l *Logger) EventAt(level Level, function, file string) *Event {
	return l.newEventContextAt(l.ctx, normalizeLevel(level), 1, &eventOrigin{function: function, file: file})
}

func (l *Logger) TraceEvent() *Event {
	if Level(l.level.Load()) > LevelTrace {
		return nil
	}
	return l.newEvent(LevelTrace, 1)
}
func (l *Logger) DebugEvent() *Event {
	if Level(l.level.Load()) > LevelDebug {
		return nil
	}
	return l.newEvent(LevelDebug, 1)
}
func (l *Logger) InfoEvent() *Event {
	return l.newEvent(LevelInfo, 1)
}
func (l *Logger) WarnEvent() *Event {
	if Level(l.level.Load()) > LevelWarn {
		return nil
	}
	return l.newEvent(LevelWarn, 1)
}
func (l *Logger) ErrorEvent() *Event {
	if Level(l.level.Load()) > LevelError {
		return nil
	}
	return l.newEvent(LevelError, 1)
}
func (l *Logger) PanicEvent() *Event { return l.newEvent(LevelPanic, 1) }
func (l *Logger) FatalEvent() *Event { return l.newEvent(LevelFatal, 1) }

func TraceEvent() *Event { return Default().TraceEvent() }
func DebugEvent() *Event { return Default().DebugEvent() }
func InfoEvent() *Event  { return Default().InfoEvent() }
func WarnEvent() *Event  { return Default().WarnEvent() }
func ErrorEvent() *Event { return Default().ErrorEvent() }
func PanicEvent() *Event { return Default().PanicEvent() }
func FatalEvent() *Event { return Default().FatalEvent() }

func (e *Event) writeInitialJSON(level Level, now time.Time) {
	if e.includeTime {
		if e.config.jsonTimeMode == JSONTimeUTC {
			e.writeDefaultJSONTimestamp(now)
		} else {
			e.output = appendTimestamp(e.output, now, e.config, true)
		}
	} else {
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
	e.addCallers()
}

func (e *Event) writeDefaultJSONTimestamp(now time.Time) {
	if e.captureCaller != 0 {
		e.output = appendDefaultJSONTimestamp(e.output, now.UTC())
		return
	}
	second := now.Unix()
	p := &e.callerData.callerFunc
	if e.timeSecond != second {
		e.timeSecond = second
		utc := now.UTC()
		year, month, day := utc.Date()
		hour, minute, sec := utc.Clock()
		p[0] = byte('0' + year/1000%10)
		p[1] = byte('0' + year/100%10)
		p[2] = byte('0' + year/10%10)
		p[3] = byte('0' + year%10)
		p[4] = '-'
		p[5] = byte('0' + int(month)/10)
		p[6] = byte('0' + int(month)%10)
		p[7] = '-'
		p[8] = byte('0' + day/10)
		p[9] = byte('0' + day%10)
		p[10] = 'T'
		p[11] = byte('0' + hour/10)
		p[12] = byte('0' + hour%10)
		p[13] = ':'
		p[14] = byte('0' + minute/10)
		p[15] = byte('0' + minute%10)
		p[16] = ':'
		p[17] = byte('0' + sec/10)
		p[18] = byte('0' + sec%10)
	}
	e.output = append(e.output, `{"time":"`...)
	e.output = append(e.output, p[:19]...)
	e.output = append(e.output, '.')
	usec := now.Nanosecond() / 1000
	e.output = append(e.output,
		byte('0'+usec/100000),
		byte('0'+usec/10000%10),
		byte('0'+usec/1000%10),
		byte('0'+usec/100%10),
		byte('0'+usec/10%10),
		byte('0'+usec%10),
		'Z', '"', ',',
	)
}

func (e *Event) writeInitialConsole(level Level, now time.Time) {
	if e.includeTime {
		e.output = appendTimestamp(e.output, now, e.config, false)
	}
	if e.color {
		e.output = append(e.output, ansiColourPrefix(level)...)
	} else {
		e.output = append(e.output, levelPrefix(level)...)
	}
	e.addCallers()
	// everything appended from here on is caller-supplied
	e.prefixLen = len(e.output)
}

func (e *Event) addCallers() {
	if e.captureCaller == 0 || e.callerData.callerFuncLen == 0 {
		return
	}
	if e.jsonMode {
		e.output = append(e.output, `,"func":"`...)
		e.output = appendJSONEscaped(e.output, unsafeString(e.callerData.callerFunc[:e.callerData.callerFuncLen]))
		e.output = append(e.output, `","file":"`...)
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
	e.output = append(e.output, e.callerData.callerFunc[:e.callerData.callerFuncLen]...)
	e.output = append(e.output, ' ')
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
