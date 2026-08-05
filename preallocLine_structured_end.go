package iqlog

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// finish writes the assembled line under the logger mutex (serializing with
// all other builders), returns the line to the pool, and honors any pending
// fatal/panic exit after the record has been written.
func (pal *preallocLine) finish() {
	exit := pal.exitAfterWrite

	// a line that filled the fixed buffer may have lost its trailing newline
	if pal.bytesUsed == maxLineLen && pal.output[pal.bytesUsed-1] != '\n' {
		pal.output[pal.bytesUsed-1] = '\n'
	}

	line := pal.output[:pal.bytesUsed]
	if !pal.jsonMode {
		line = sanitizeConsoleLine(line)
	}

	pal.logger.writeLocked(line)

	doPanic := pal.panicAfterWrite
	var panicMsg string
	if doPanic {
		panicMsg = strings.TrimRight(string(line), "\n")
	}
	pal.exitAfterWrite = false
	pal.panicAfterWrite = false
	preallocLinePool.Put(pal)

	if doPanic {
		panic(panicMsg)
	}
	if exit {
		os.Exit(1)
	}
}

// finishJSON closes the JSON object (repairing the record if it was
// truncated by the fixed buffer) and writes it out.
func (pal *preallocLine) finishJSON() {
	if pal.bytesUsed+3 <= maxLineLen {
		pal.bytesUsed += copy(pal.output[pal.bytesUsed:], "\"}\n")
	} else {
		pal.bytesUsed = repairTruncatedJSONLine(pal.output[:], pal.bytesUsed, maxLineLen-1)
		pal.output[pal.bytesUsed] = '\n'
		pal.bytesUsed++
	}
	pal.finish()
}

func (pal *preallocLine) Msg(msg string) {
	if pal.bytesUsed == 0 {
		return
	}
	if pal.jsonMode {
		pal.writeFinalJSON(msg)
	} else {
		pal.writeFinalConsole(msg)
	}
}
func (pal *preallocLine) Msgs(msg string, args ...interface{}) {
	if pal.bytesUsed == 0 {
		return
	}
	if pal.jsonMode {
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeFinalConsole(msg, args...)
	}
}
func (pal *preallocLine) Msgf(format string, args ...interface{}) {
	if pal.bytesUsed == 0 {
		return
	}

	if pal.jsonMode {
		pal.writeFinalJSONF(format, args...)
	} else {
		pal.writeFinalConsoleF(format, args...)
	}
}

func (pal *preallocLine) writeFinalConsole(msg string, args ...interface{}) {

	// put the message
	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, msg)

	// followed by args
	for _, thisarg := range args {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")

		switch thisarg := thisarg.(type) {
		case string:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, thisarg)
		case int:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.Itoa(thisarg))
		case int32:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.Itoa(int(thisarg)))
		case uint32:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatUint(uint64(thisarg), 10))
		case int64:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatInt(thisarg, 10))
		case uint64:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatUint(thisarg, 10))
		case float32:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatFloat(float64(thisarg), 'f', -1, 32))
		case float64:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatFloat(thisarg, 'f', -1, 64))
		case bool:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatBool(thisarg))
		default:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, fmt.Sprintf("%v", thisarg))
		}
	}
	// if pal.logger.newLine {
	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\n")
	// }

	pal.finish()
}

func (pal *preallocLine) writeFinalConsoleF(format string, args ...interface{}) {

	// ba := baPool.Get().(*bytesAppender)
	// ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	// fmt.Fprintf(ba, format, args...)
	// pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, string(bia.Bytes[:bia.Index]))
	// baPool.Put(ba)

	bia := biapool.Get().(*byteIndexAppender)
	bia.Index = 0
	fmt.Fprintf(bia, format, args...)
	pal.bytesUsed += copy(pal.output[pal.bytesUsed:], bia.Bytes[:bia.Index])
	biapool.Put(bia)

	// if pal.logger.newLine {
	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\n")
	// }

	pal.finish()
}

func (pal *preallocLine) writeFinalJSON(msg string, args ...interface{}) {
	// write JSON closer

	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"message\":\"")

	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(msg))

	// followed by args
	for _, thisarg := range args {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")

		switch thisarg := thisarg.(type) {
		case string:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(thisarg))
		case int:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.Itoa(thisarg))
		case int32:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.Itoa(int(thisarg)))
		case uint32:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatUint(uint64(thisarg), 10))
		case int64:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatInt(thisarg, 10))
		case uint64:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatUint(thisarg, 10))
		case float32:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatFloat(float64(thisarg), 'f', -1, 32))
		case float64:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatFloat(thisarg, 'f', -1, 64))
		case bool:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, strconv.FormatBool(thisarg))
		default:
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(fmt.Sprintf("%v", thisarg)))
		}
	}

	pal.finishJSON()
}

func (pal *preallocLine) writeFinalJSONF(format string, args ...interface{}) {
	// write JSON closer

	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"message\":\"")

	// ba := baPool.Get().(*bytesAppender)
	// ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	// fmt.Fprintf(ba, format, args...)
	// pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, string(bia.Bytes[:bia.Index]))
	// baPool.Put(ba)

	bia := biapool.Get().(*byteIndexAppender)
	bia.Index = 0
	fmt.Fprintf(bia, format, args...)
	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(unsafeString(bia.Bytes[:bia.Index])))
	biapool.Put(bia)

	pal.finishJSON()
}
