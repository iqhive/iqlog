package iqlog

import (
	"fmt"
	"strconv"
	"strings"
)

func sprintMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}
	output := msg
	for _, arg := range args {
		output += " " + fmt.Sprint(arg)
	}
	return output
}

func (bsl *Event) terminateDisabled(message string) bool {
	if bsl == nil {
		return true
	}
	if !bsl.disabled {
		return false
	}
	if bsl.panicAfterWrite {
		bsl.release(true)
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
	if !bsl.jsonMode {
		line = sanitizeConsoleLine(line)
	}

	owned := bsl.logger.writeRecord(line, bsl.level)

	if !bsl.panicAfterWrite && !bsl.exitAfterWrite {
		bsl.release(!owned)
		return
	}

	exit := bsl.exitAfterWrite
	logger := bsl.logger
	exitFunc := bsl.config.exitFunc
	panicMsg := bsl.panicMessage
	bsl.release(!owned)

	if !exit {
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
		bsl.release(true)
		panic("")
	}
	if bsl.exitAfterWrite {
		exitFunc := bsl.config.exitFunc
		bsl.release(true)
		exitFunc(1)
		return
	}
	bsl.release(true)
}
func (bsl *Event) Msgs(msg string, args ...interface{}) {
	panicMessage := sprintMessage(msg, args...)
	if bsl.terminateDisabled(panicMessage) {
		return
	}
	if bsl.consumed {
		return
	}
	if bsl.panicAfterWrite {
		bsl.panicMessage = panicMessage
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

	//put the message
	bsl.output = append(bsl.output, msg...)

	// followed by args
	for _, thisarg := range args {
		bsl.output = append(bsl.output, ' ')

		switch thisarg := thisarg.(type) {
		case string:
			bsl.output = append(bsl.output, thisarg...)
		case int:
			bsl.output = append(bsl.output, strconv.Itoa(thisarg)...)
		case int32:
			bsl.output = append(bsl.output, strconv.Itoa(int(thisarg))...)
		case uint32:
			bsl.output = append(bsl.output, strconv.FormatUint(uint64(thisarg), 10)...)
		case int64:
			bsl.output = append(bsl.output, strconv.FormatInt(thisarg, 10)...)
		case uint64:
			bsl.output = append(bsl.output, strconv.FormatUint(thisarg, 10)...)
		case float32:
			bsl.output = append(bsl.output, strconv.FormatFloat(float64(thisarg), 'f', -1, 32)...)
		case float64:
			bsl.output = append(bsl.output, strconv.FormatFloat(thisarg, 'f', -1, 64)...)
		case bool:
			bsl.output = append(bsl.output, strconv.FormatBool(thisarg)...)
		default:
			bsl.output = append(bsl.output, fmt.Sprintf("%v", thisarg)...)
		}
	}
	// if bsl.logger.newLine.Load() {
	bsl.output = append(bsl.output, '\n')
	// }

	bsl.finish()
}

func (bsl *Event) writeFinalConsoleF(format string, args ...interface{}) {
	// the growable appender keeps formatted output longer than maxLineLen
	ba := baPool.Get().(*bytesAppender)
	ba.Bytes = ba.Bytes[:0]
	fmt.Fprintf(ba, format, args...)
	bsl.output = append(bsl.output, ba.Bytes...)
	baPool.Put(ba)

	bsl.output = append(bsl.output, '\n')

	bsl.finish()
}

func (bsl *Event) writeFinalJSON(msg string, args ...interface{}) {
	// write JSON closer
	bsl.output = append(bsl.output, []byte(",\"message\":\"")...)

	bsl.output = appendJSONEscaped(bsl.output, msg)

	// followed by args
	for _, thisarg := range args {
		bsl.output = append(bsl.output, ' ')

		switch thisarg := thisarg.(type) {
		case string:
			bsl.output = appendJSONEscaped(bsl.output, thisarg)
		case int:
			bsl.output = append(bsl.output, strconv.Itoa(thisarg)...)
		case int32:
			bsl.output = append(bsl.output, strconv.Itoa(int(thisarg))...)
		case uint32:
			bsl.output = append(bsl.output, strconv.FormatUint(uint64(thisarg), 10)...)
		case int64:
			bsl.output = append(bsl.output, strconv.FormatInt(thisarg, 10)...)
		case uint64:
			bsl.output = append(bsl.output, strconv.FormatUint(thisarg, 10)...)
		case float32:
			bsl.output = appendJSONStringFloat(bsl.output, float64(thisarg), 32)
		case float64:
			bsl.output = appendJSONStringFloat(bsl.output, thisarg, 64)
		case bool:
			bsl.output = append(bsl.output, strconv.FormatBool(thisarg)...)
		default:
			bsl.output = appendJSONEscaped(bsl.output, fmt.Sprintf("%v", thisarg))
		}
	}

	// if bsl.newLine {
	bsl.output = append(bsl.output, []byte("\"}\n")...)
	// } else {
	// 	bsl.output = append(bsl.output, []byte("\"}")...)
	// }

	bsl.finish()
}

func (bsl *Event) writeFinalJSONF(format string, args ...interface{}) {
	bsl.output = append(bsl.output, []byte(",\"message\":\"")...)

	ba := baPool.Get().(*bytesAppender)
	ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	fmt.Fprintf(ba, format, args...)
	bsl.output = appendJSONEscaped(bsl.output, unsafeString(ba.Bytes))
	baPool.Put(ba)

	// bia := biapool.Get().(*byteIndexAppender)
	// bia.Index = 0
	// fmt.Fprintf(bia, format, args...)
	// bsl.output = append(bsl.output, bia.Bytes[:bia.Index]...)
	// biapool.Put(bia)

	// if bsl.logger.newLine.Load() {
	bsl.output = append(bsl.output, []byte("\"}\n")...)

	bsl.finish()
}
