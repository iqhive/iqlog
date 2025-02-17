package iqlog

import (
	"fmt"
	"strconv"
)

func (pal *preallocLine2) Msg(msg string) {
	if pal.output == nil {
		return
	}
	if pal.jsonMode {
		pal.writeFinalJSON(msg)
	} else {
		pal.writeFinalConsole(msg)
	}
}
func (pal *preallocLine2) Msgs(msg string, args ...interface{}) {
	if pal.output == nil {
		return
	}
	if pal.jsonMode {
		pal.writeFinalJSON(msg, args...)
	} else {
		pal.writeFinalConsole(msg, args...)
	}
}
func (pal *preallocLine2) Msgf(format string, args ...interface{}) {
	if pal.output == nil {
		return
	}

	if pal.jsonMode {
		pal.writeFinalJSONF(format, args...)
	} else {
		pal.writeFinalConsoleF(format, args...)
	}
}

func (pal *preallocLine2) writeFinalConsole(msg string, args ...interface{}) {

	//  put the message
	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, msg)

	// followed by args
	for _, thisarg := range args {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")

		switch thisarg := thisarg.(type) {
		case string:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, thisarg)
		case int:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.Itoa(thisarg))
		case int32:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.Itoa(int(thisarg)))
		case uint32:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatUint(uint64(thisarg), 10))
		case int64:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatInt(thisarg, 10))
		case uint64:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatUint(thisarg, 10))
		case float32:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatFloat(float64(thisarg), 'f', -1, 32))
		case float64:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatFloat(thisarg, 'f', -1, 64))
		case bool:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatBool(thisarg))
		default:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, fmt.Sprintf("%v", thisarg))
		}
	}
	// if pal.logger.newLine {
	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\n")
	// }

	pal.out.Write(pal.output[:pal.bytesUsed])

	preallocLine2Pool.Put(pal)
}

func (pal *preallocLine2) writeFinalConsoleF(format string, args ...interface{}) {

	// ba := baPool.Get().(*bytesAppender)
	// ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	// fmt.Fprintf(ba, format, args...)
	// copy(pal.output[pal.bytesUsed:], bia.Bytes[:bia.Index])
	// baPool.Put(ba)

	bia := biapool.Get().(*byteIndexAppender)
	bia.Index = 0
	fmt.Fprintf(bia, format, args...)
	copy(pal.output[pal.bytesUsed:], bia.Bytes[:bia.Index])
	pal.bytesUsed += bia.Index
	// pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\n")
	biapool.Put(bia)

	// if pal.logger.newLine {
	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\n")
	// }

	pal.out.Write(pal.output[:pal.bytesUsed])

	preallocLine2Pool.Put(pal)
}

func (pal *preallocLine2) writeFinalJSON(msg string, args ...interface{}) {
	// write JSON closer
	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"message\":\"")

	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, msg)

	// followed by args
	for _, thisarg := range args {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")

		switch thisarg := thisarg.(type) {
		case string:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, thisarg)
		case int:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.Itoa(thisarg))
		case int32:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.Itoa(int(thisarg)))
		case uint32:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatUint(uint64(thisarg), 10))
		case int64:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatInt(thisarg, 10))
		case uint64:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatUint(thisarg, 10))
		case float32:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatFloat(float64(thisarg), 'f', -1, 32))
		case float64:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatFloat(thisarg, 'f', -1, 64))
		case bool:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, strconv.FormatBool(thisarg))
		default:
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, fmt.Sprintf("%v", thisarg))
		}
	}

	// if pal.logger.newLine {
	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"}\n")
	// } else {
	// 	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"}")
	// }

	pal.out.Write(pal.output[:pal.bytesUsed])

	preallocLine2Pool.Put(pal)
}

func (pal *preallocLine2) writeFinalJSONF(format string, args ...interface{}) {
	// write JSON closer
	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"message\":\"")

	// ba := baPool.Get().(*bytesAppender)
	// ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	// fmt.Fprintf(ba, format, args...)
	// bsl.output = append(bsl.output, ba.Bytes...)
	// baPool.Put(ba)

	bia := biapool.Get().(*byteIndexAppender)
	bia.Index = 0
	fmt.Fprintf(bia, format, args...)
	copy(pal.output[pal.bytesUsed:], bia.Bytes[:bia.Index])
	pal.bytesUsed += bia.Index
	biapool.Put(bia)

	// if pal.logger.newLine {
	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"}\n")
	// } else {
	// 	pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"}")
	// }

	pal.out.Write(pal.output[:pal.bytesUsed])

	preallocLine2Pool.Put(pal)
}
