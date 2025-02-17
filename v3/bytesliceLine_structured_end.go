package iqlog

import (
	"fmt"
	"strconv"
)

func (bsl *bytesliceLine) Msg(msg string) {
	if bsl.output == nil {
		return
	}
	if bsl.jsonMode {
		bsl.writeFinalJSON(msg)
	} else {
		bsl.writeFinalConsole(msg)
	}
}
func (bsl *bytesliceLine) Msgs(msg string, args ...interface{}) {
	if bsl.output == nil {
		return
	}
	if bsl.jsonMode {
		bsl.writeFinalJSON(msg, args...)
	} else {
		bsl.writeFinalConsole(msg, args...)
	}
}
func (bsl *bytesliceLine) Msgf(format string, args ...interface{}) {
	if bsl.output == nil {
		return
	}
	if bsl.jsonMode {
		bsl.writeFinalJSONF(format, args...)
	} else {
		bsl.writeFinalConsoleF(format, args...)
	}
}

func (bsl *bytesliceLine) writeFinalConsole(msg string, args ...interface{}) {

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
	// if bsl.logger.newLine {
	bsl.output = append(bsl.output, '\n')
	// }

	bsl.out.Write(bsl.output)

	bytesliceLinePool.Put(bsl)
}

func (bsl *bytesliceLine) writeFinalConsoleF(format string, args ...interface{}) {

	// ba := baPool.Get().(*bytesAppender)
	// ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	// fmt.Fprintf(ba, format, args...)
	// bsl.output = append(bsl.output, ba.Bytes...)
	// baPool.Put(ba)

	bia := biapool.Get().(*byteIndexAppender)
	bia.Index = 0
	fmt.Fprintf(bia, format, args...)
	bsl.output = append(bsl.output, bia.Bytes[:bia.Index]...)
	biapool.Put(bia)

	// if bsl.logger.newLine {
	bsl.output = append(bsl.output, '\n')
	// }

	bsl.out.Write(bsl.output)

	bytesliceLinePool.Put(bsl)
}

func (bsl *bytesliceLine) writeFinalJSON(msg string, args ...interface{}) {
	// write JSON closer
	bsl.output = append(bsl.output, []byte(",\"message\":\"")...)

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
			bsl.output, _ = appendfastFloatFill(bsl.output, float64(thisarg), 6)
		case float64:
			bsl.output, _ = appendfastFloatFill(bsl.output, thisarg, 6)
		case bool:
			bsl.output = append(bsl.output, strconv.FormatBool(thisarg)...)
		default:
			bsl.output = append(bsl.output, fmt.Sprintf("%v", thisarg)...)
		}
	}

	// if bsl.newLine {
	bsl.output = append(bsl.output, []byte("\"}\n")...)
	// } else {
	// 	bsl.output = append(bsl.output, []byte("\"}")...)
	// }

	bsl.out.Write(bsl.output)

	bytesliceLinePool.Put(bsl)
}

func (bsl *bytesliceLine) writeFinalJSONF(format string, args ...interface{}) {
	bsl.output = append(bsl.output, []byte(",\"message\":\"")...)

	ba := baPool.Get().(*bytesAppender)
	ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	fmt.Fprintf(ba, format, args...)
	bsl.output = append(bsl.output, ba.Bytes...)
	baPool.Put(ba)

	// bia := biapool.Get().(*byteIndexAppender)
	// bia.Index = 0
	// fmt.Fprintf(bia, format, args...)
	// bsl.output = append(bsl.output, bia.Bytes[:bia.Index]...)
	// biapool.Put(bia)

	// if bsl.logger.newLine {
	bsl.output = append(bsl.output, []byte("\"}\n")...)

	bsl.out.Write(bsl.output)

	bytesliceLinePool.Put(bsl)
}
