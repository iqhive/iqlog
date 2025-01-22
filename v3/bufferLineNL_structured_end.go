package iqlog

import (
	"fmt"
	"strconv"
)

func (bl *bufferLineNL) Msg(msg string) {
	if bl.buffer == nil {
		return
	}
	if bl.jsonMode {
		bl.writeFinalJSON(msg)
	} else {
		bl.writeFinalConsole(msg)
	}
}
func (bl *bufferLineNL) Msgf(format string, args ...interface{}) {
	if bl.buffer == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	bl.Msg(msg)
}

func (bl *bufferLineNL) writeFinalConsole(msg string, args ...interface{}) {
	bl.buffer.WriteString(msg)

	// followed by args
	for _, thisarg := range args {
		bl.buffer.WriteByte(' ')

		switch thisarg := thisarg.(type) {
		case string:
			bl.buffer.WriteString(thisarg)
		case int:
			bl.buffer.WriteString(strconv.Itoa(thisarg))
		case int32:
			bl.buffer.WriteString(strconv.Itoa(int(thisarg)))
		case uint32:
			bl.buffer.WriteString(strconv.FormatUint(uint64(thisarg), 10))
		case int64:
			bl.buffer.WriteString(strconv.FormatInt(thisarg, 10))
		case uint64:
			bl.buffer.WriteString(strconv.FormatUint(thisarg, 10))
		case float32:
			bl.buffer.WriteString(strconv.FormatFloat(float64(thisarg), 'f', -1, 32))
		case float64:
			bl.buffer.WriteString(strconv.FormatFloat(thisarg, 'f', -1, 64))
		case bool:
			bl.buffer.WriteString(strconv.FormatBool(thisarg))
		default:
			bl.buffer.WriteString(fmt.Sprintf("%v", thisarg))
		}
	}
	// if bl.logger.newLine {
	bl.buffer.WriteByte('\n')
	// }

	bl.out.Write(bl.buffer.Bytes())

	bufferLineNLPool.Put(bl)
}

func (bl *bufferLineNL) writeFinalJSON(msg string, args ...interface{}) {
	// write JSON closer

	bl.buffer.WriteString(msg)

	// followed by args
	for _, thisarg := range args {
		bl.buffer.WriteByte(' ')

		switch thisarg := thisarg.(type) {
		case string:
			bl.buffer.WriteString(thisarg)
		case int:
			bl.buffer.WriteString(strconv.Itoa(thisarg))
		case int32:
			bl.buffer.WriteString(strconv.Itoa(int(thisarg)))
		case uint32:
			bl.buffer.WriteString(strconv.FormatUint(uint64(thisarg), 10))
		case int64:
			bl.buffer.WriteString(strconv.FormatInt(thisarg, 10))
		case uint64:
			bl.buffer.WriteString(strconv.FormatUint(thisarg, 10))
		case float32:
			bl.buffer.WriteString(strconv.FormatFloat(float64(thisarg), 'f', -1, 64))
		case float64:
			bl.buffer.WriteString(strconv.FormatFloat(thisarg, 'f', -1, 64))
		case bool:
			bl.buffer.WriteString(strconv.FormatBool(thisarg))
		default:
			bl.buffer.WriteString(fmt.Sprintf("%v", thisarg))
		}
	}

	// if bl.logger.newLine {
	bl.buffer.Write([]byte("\"}\n"))
	// } else {
	// 	bl.buffer.Write([]byte("\"}"))
	// }

	bl.out.Write(bl.buffer.Bytes())

	bufferLineNLPool.Put(bl)
}
