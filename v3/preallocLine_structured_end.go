package iqlog

import (
	"fmt"
	"strconv"
)

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
func (pal *preallocLine) Msgf(format string, args ...interface{}) {
	if pal.bytesUsed == 0 {
		return
	}
	msg := fmt.Sprintf(format, args...)
	pal.Msg(msg)
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

	pal.out.Write(pal.output[:pal.bytesUsed])

	preallocLinePool.Put(pal)
}

func (pal *preallocLine) writeFinalJSON(msg string, args ...interface{}) {
	// write JSON closer

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
	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"}\n")
	// } else {
	// 	pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"}")
	// }
	pal.out.Write(pal.output[:pal.bytesUsed])

	preallocLinePool.Put(pal)
}
