package iqlog

import (
	"context"
	"sort"
)

func (l *Logger) newEvent(level Level, callerSkip int) *Event {
	return l.newEventContext(l.ctx, level, callerSkip)
}

func (l *Logger) newEventContext(ctx context.Context, level Level, callerSkip int) *Event {
	return l.newEventContextAt(ctx, level, callerSkip+1, "", "", false)
}

func (l *Logger) newEventContextAt(ctx context.Context, level Level, callerSkip int, function, file string, explicitCaller bool) *Event {
	if Level(l.level.Load()) > level {
		if level != LevelFatal && level != LevelPanic {
			return nil
		}
		cfg := l.config.Load()
		e := acquireEvent(l, cfg)
		e.disabled = true
		e.exitAfterWrite = level == LevelFatal
		e.panicAfterWrite = level == LevelPanic
		return e
	}
	cfg := l.config.Load()
	if cfg.level > level {
		return nil
	}
	e := acquireEvent(l, cfg)
	e.jsonMode = cfg.format == FormatJSON
	e.includeTime = cfg.includeTime
	e.color = cfg.color
	e.captureCaller = cfg.callerDepth
	e.exitAfterWrite = level == LevelFatal
	e.panicAfterWrite = level == LevelPanic
	if explicitCaller {
		e.captureCaller = 0
		if function != "" {
			e.captureCaller = 1
			e.callerData.callerFuncLen = uint(copy(e.callerData.callerFunc[:], function))
			e.callerData.callerFileLen = uint(copy(e.callerData.callerFile[:], file))
		}
	} else if cfg.callerDepth > 0 {
		captureCaller(cfg.callerDepth+callerSkip+1, &e.callerData)
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
func (l *Logger) Event(level Level) *Event { return l.newEvent(normalizeLevel(level), 1) }

// EventAt creates a structured event with an explicit caller. Empty function
// and file values suppress caller output regardless of CallerDepth.
func (l *Logger) EventAt(level Level, function, file string) *Event {
	return l.newEventContextAt(l.ctx, normalizeLevel(level), 1, function, file, true)
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
	if Level(l.level.Load()) > LevelInfo {
		return nil
	}
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

func (e *Event) writeInitialJSON(level Level) {
	if e.includeTime {
		e.output = appendTimestamp(e.output, e.config.now(), e.config, true)
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

func (e *Event) writeInitialConsole(level Level) {
	if e.includeTime {
		e.output = appendTimestamp(e.output, e.config.now(), e.config, false)
	}
	if e.color {
		e.output = append(e.output, ansiColourPrefix(level)...)
	} else {
		e.output = append(e.output, levelPrefix(level)...)
	}
	e.addCallers()
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
	e.output = append(e.output, e.callerData.callerFunc[:e.callerData.callerFuncLen]...)
	e.output = append(e.output, ' ')
	e.output = append(e.output, e.callerData.callerFile[:e.callerData.callerFileLen]...)
	if e.color {
		e.output = append(e.output, "]\x1b[0m "...)
	} else {
		e.output = append(e.output, ']', ' ')
	}
}
