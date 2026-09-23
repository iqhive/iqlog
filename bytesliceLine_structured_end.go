package iqlog

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

// sprintMessage joins a message and its Msgs arguments the way the console
// record carries them, for use as a panic value.
func sprintMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}
	ba := acquireBytesAppender()
	ba.Bytes = append(ba.Bytes, msg...)
	for _, arg := range args {
		ba.Bytes = append(ba.Bytes, ' ')
		ba.Bytes = appendMsgArg(ba.Bytes, arg, false)
	}
	s := string(ba.Bytes)
	releaseBytesAppender(ba)
	return s
}

// appendMsgArg appends one Msgs argument in the form the message carries it.
// Strings are appended verbatim; when the record is JSON the caller has
// already opened the message string, so they are escaped instead. Numbers use
// the shortest round-trip form in both formats, the same form the typed field
// methods use, and are appended in place so scalar arguments do not allocate.
func appendMsgArg(dst []byte, arg any, jsonMode bool) []byte {
	switch v := arg.(type) {
	case string:
		if jsonMode {
			return appendJSONEscaped(dst, v)
		}
		return append(dst, v...)
	case int:
		return strconv.AppendInt(dst, int64(v), 10)
	case int8:
		return strconv.AppendInt(dst, int64(v), 10)
	case int16:
		return strconv.AppendInt(dst, int64(v), 10)
	case int32:
		return strconv.AppendInt(dst, int64(v), 10)
	case int64:
		return strconv.AppendInt(dst, v, 10)
	case uint:
		return strconv.AppendUint(dst, uint64(v), 10)
	case uint8:
		return strconv.AppendUint(dst, uint64(v), 10)
	case uint16:
		return strconv.AppendUint(dst, uint64(v), 10)
	case uint32:
		return strconv.AppendUint(dst, uint64(v), 10)
	case uint64:
		return strconv.AppendUint(dst, v, 10)
	case float32:
		return strconv.AppendFloat(dst, float64(v), 'f', -1, 32)
	case float64:
		return strconv.AppendFloat(dst, v, 'f', -1, 64)
	case bool:
		return strconv.AppendBool(dst, v)
	}
	if !jsonMode {
		return fmt.Appendf(dst, "%v", arg)
	}
	// formatted into scratch first because the text lands inside a JSON string
	ba := acquireBytesAppender()
	ba.Bytes = fmt.Appendf(ba.Bytes, "%v", arg)
	dst = appendJSONEscaped(dst, unsafeString(ba.Bytes))
	releaseBytesAppender(ba)
	return dst
}

func (bsl *Event) terminateDisabled(message string) bool {
	if bsl == nil {
		return true
	}
	if !bsl.disabled {
		return false
	}
	if bsl.panicAfterWrite {
		logger := bsl.logger
		bsl.release(true)
		_ = logger.Flush()
		panic(message)
	}
	if bsl.exitAfterWrite {
		logger, exitFunc := bsl.logger, bsl.config.exitFunc
		bsl.release(true)
		_ = logger.Flush()
		exitFunc(1)
	}
	return true
}

// finish writes the completed line to the logger's writer under the
// logger's lock, recycles the line, and exits if the line was started
// by a Fatal/Panic builder.
func (bsl *Event) finish() {
	line := bsl.output
	copied := false
	if !bsl.jsonMode {
		line = sanitizeConsoleLine(line, bsl.prefixLen)
		copied = unsafe.SliceData(line) != unsafe.SliceData(bsl.output)
	}

	owned := bsl.logger.writeRecord(line, bsl.level)
	// An async writer takes ownership of line and recycles it once written.
	// When line was the sanitized copy, the event's own buffer was not handed
	// over and stays pooled with the event; when the copy was written
	// synchronously it goes straight back to the pool it came from.
	if copied && !owned {
		releaseEventBuffer(line)
	}
	keepBuffer := !owned || copied

	if !bsl.panicAfterWrite && !bsl.exitAfterWrite {
		bsl.release(keepBuffer)
		return
	}

	exit := bsl.exitAfterWrite
	logger := bsl.logger
	exitFunc := bsl.config.exitFunc
	panicMsg := bsl.panicMessage
	bsl.release(keepBuffer)

	if !exit {
		_ = logger.Flush()
		panic(panicMsg)
	}
	_ = logger.Flush()
	exitFunc(1)
}

func (bsl *Event) Msg(msg string) {
	if bsl.terminateDisabled(msg) {
		return
	}
	if bsl.consumed {
		return
	}
	if bsl.panicAfterWrite {
		bsl.panicMessage = msg
	}
	if bsl.jsonMode {
		bsl.writeFinalJSON(msg)
	} else {
		bsl.writeFinalConsole(msg)
	}
}

// Discard consumes a non-terminal event without writing. Panic and fatal
// events retain their termination semantics.
func (bsl *Event) Discard() {
	if bsl == nil {
		return
	}
	if bsl.consumed {
		return
	}
	if bsl.panicAfterWrite {
		logger := bsl.logger
		bsl.release(true)
		_ = logger.Flush()
		panic("iqlog: panic event discarded without a message")
	}
	if bsl.exitAfterWrite {
		logger := bsl.logger
		exitFunc := bsl.config.exitFunc
		bsl.release(true)
		_ = logger.Flush()
		exitFunc(1)
		return
	}
	bsl.release(true)
}
func (bsl *Event) Msgs(msg string, args ...interface{}) {
	if bsl == nil {
		// dropped by the level gate: do not format the arguments
		return
	}
	// Only a panic needs the arguments joined up front, as the panic value.
	// Formatting them for a record that will not be written would allocate on
	// the disabled path.
	if bsl.disabled {
		panicMessage := ""
		if bsl.panicAfterWrite {
			panicMessage = sprintMessage(msg, args...)
		}
		bsl.terminateDisabled(panicMessage)
		return
	}
	if bsl.consumed {
		return
	}
	if bsl.panicAfterWrite {
		bsl.panicMessage = sprintMessage(msg, args...)
	}
	if bsl.jsonMode {
		bsl.writeFinalJSON(msg, args...)
	} else {
		bsl.writeFinalConsole(msg, args...)
	}
}
func (bsl *Event) Msgf(format string, args ...interface{}) {
	if bsl == nil {
		return
	}
	if len(args) == 0 && !strings.Contains(format, "%") {
		bsl.Msg(format)
		return
	}
	panicMessage := ""
	if bsl.panicAfterWrite {
		panicMessage = fmt.Sprintf(format, args...)
	}
	if bsl.terminateDisabled(panicMessage) {
		return
	}
	if bsl.consumed {
		return
	}
	bsl.panicMessage = panicMessage
	if bsl.jsonMode {
		bsl.writeFinalJSONF(format, args...)
	} else {
		bsl.writeFinalConsoleF(format, args...)
	}
}

func (bsl *Event) writeFinalConsole(msg string, args ...interface{}) {
	bsl.output = append(bsl.output, msg...)
	for _, arg := range args {
		bsl.output = append(bsl.output, ' ')
		bsl.output = appendMsgArg(bsl.output, arg, false)
	}
	bsl.output = append(bsl.output, '\n')

	bsl.finish()
}

func (bsl *Event) writeFinalConsoleF(format string, args ...interface{}) {
	// the growable appender keeps formatted output longer than maxLineLen
	ba := acquireBytesAppender()
	fmt.Fprintf(ba, format, args...)
	bsl.output = append(bsl.output, ba.Bytes...)
	releaseBytesAppender(ba)

	bsl.output = append(bsl.output, '\n')

	bsl.finish()
}

func (bsl *Event) writeFinalJSON(msg string, args ...interface{}) {
	bsl.output = append(bsl.output, `,"message":"`...)
	bsl.output = appendJSONEscaped(bsl.output, msg)
	for _, arg := range args {
		bsl.output = append(bsl.output, ' ')
		bsl.output = appendMsgArg(bsl.output, arg, true)
	}
	bsl.output = append(bsl.output, "\"}\n"...)

	bsl.finish()
}

func (bsl *Event) writeFinalJSONF(format string, args ...interface{}) {
	bsl.output = append(bsl.output, []byte(",\"message\":\"")...)

	ba := acquireBytesAppender()
	fmt.Fprintf(ba, format, args...)
	bsl.output = appendJSONEscaped(bsl.output, unsafeString(ba.Bytes))
	releaseBytesAppender(ba)

	// bia := biapool.Get().(*byteIndexAppender)
	// bia.Index = 0
	// fmt.Fprintf(bia, format, args...)
	// bsl.output = append(bsl.output, bia.Bytes[:bia.Index]...)
	// biapool.Put(bia)

	// if bsl.logger.newLine.Load() {
	bsl.output = append(bsl.output, []byte("\"}\n")...)

	bsl.finish()
}
