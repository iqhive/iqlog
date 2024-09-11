package iqlog

import (
	"fmt"
	"strings"
	"time"
)

// logMessageNoAlloc handles a fast path with no expansions/callers.
func (l *logger) logMessageNoAlloc(level Level, message string) {
	if !l.Enabled(l.ctx, level) {
		return
	}
	record := l.baseRecord
	record.Level = level
	record.Time = time.Now()
	record.MsgLen = safeStringCopy(&record.Message, message)
	_ = l.WriteRecordNoAlloc(record)
}

func (l *logger) logMessage(level Level, format string, args ...interface{}) {
	// If we want zero alloc, bail out with our specialized path
	// *only* when captureCallers == false, jsonMode == true,
	// and there are no args (which would require expansions).
	if !l.captureCallers && l.jsonMode && len(args) == 0 {
		l.logMessageNoAlloc(level, format)
		return
	}

	if !l.Enabled(l.ctx, level) {
		return
	}

	if strings.Contains(format, "%w") {
		format = strings.ReplaceAll(format, "%w", "%v")
	}

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	record := l.baseRecord
	record.Level = level
	record.Time = time.Now()
	record.MsgLen = safeStringCopy(&record.Message, msg)
	_ = l.WriteRecordNoAlloc(record)
}
