package iqlog

import (
	"fmt"
	"strconv"
)

func (sbl *bytesliceLine) Msg(msg string) {
	if sbl.output == nil {
		return
	}
	if sbl.jsonMode {
		sbl.writeFinalJSON(msg)
	} else {
		sbl.writeFinalConsole(msg)
	}
}
func (sbl *bytesliceLine) Msgf(format string, args ...interface{}) {
	if sbl.output == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	sbl.Msg(msg)
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

func (bsl *bytesliceLine) writeFinalJSON(msg string, args ...interface{}) {
	// write JSON closer

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
