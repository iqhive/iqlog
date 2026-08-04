package iqlog

import (
	"fmt"
	"strconv"
)

func (bl *bufferLine) Msg(msg string) {
	if bl.buffer == nil {
		return
	}
	if bl.logger.jsonMode {
		bl.writeFinalJSON(msg)
	} else {
		bl.writeFinalConsole(msg)
	}
}
func (bl *bufferLine) Msgs(msg string, args ...interface{}) {
	if bl.buffer == nil {
		return
	}
	if bl.logger.jsonMode {
		bl.writeFinalJSON(msg, args...)
	} else {
		bl.writeFinalConsole(msg, args...)
	}
}
func (bl *bufferLine) Msgf(format string, args ...interface{}) {
	if bl.buffer == nil {
		return
	}

	if bl.logger.jsonMode {
		bl.writeFinalJSONF(format, args...)
	} else {
		bl.writeFinalConsoleF(format, args...)
	}
}

func (bl *bufferLine) writeFinalConsole(msg string, args ...interface{}) {
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
	if bl.logger.newLine {
		bl.buffer.WriteByte('\n')
	}

	bl.logger.mu.Lock()
	bl.logger.out.Write(bl.buffer.Bytes())
	bl.logger.mu.Unlock()
	bufferLinePool.Put(bl)
}

func (bl *bufferLine) writeFinalConsoleF(format string, args ...interface{}) {
	ba := baPool.Get().(*bytesAppender)
	ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	fmt.Fprintf(ba, format, args...)
	bl.buffer.Grow(len(ba.Bytes))
	bl.buffer.Write(ba.Bytes)
	baPool.Put(ba)

	// bia := biapool.Get().(*byteIndexAppender)
	// bia.Index = 0
	// fmt.Fprintf(bia, format, args...)
	// bl.buffer.Grow(bia.Index)
	// bl.buffer.Write(bia.Bytes[:bia.Index])
	// biapool.Put(bia)

	if bl.logger.newLine {
		bl.buffer.WriteByte('\n')
	}

	bl.logger.mu.Lock()
	bl.logger.out.Write(bl.buffer.Bytes())
	bl.logger.mu.Unlock()
	bufferLinePool.Put(bl)
}

func (bl *bufferLine) writeFinalJSON(msg string, args ...interface{}) {
	bl.buffer.WriteString(",\"message\":\"")
	bl.buffer.WriteString(jsonEscapedString(msg))

	// followed by args
	for _, thisarg := range args {
		bl.buffer.WriteByte(' ')

		switch thisarg := thisarg.(type) {
		case string:
			bl.buffer.WriteString(jsonEscapedString(thisarg))
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
			bl.buffer.WriteString(jsonEscapedString(fmt.Sprintf("%v", thisarg)))
		}
	}

	if bl.logger.newLine {
		bl.buffer.Write([]byte("\"}\n"))
	} else {
		bl.buffer.Write([]byte("\"}"))
	}
	bl.logger.mu.Lock()
	bl.logger.out.Write(bl.buffer.Bytes())
	bl.logger.mu.Unlock()
	bufferLinePool.Put(bl)
}

func (bl *bufferLine) writeFinalJSONF(format string, args ...interface{}) {
	bl.buffer.WriteString(",\"message\":\"")

	ba := baPool.Get().(*bytesAppender)
	ba.Bytes = ba.Bytes[:0] // Clear the slice before use
	fmt.Fprintf(ba, format, args...)
	bl.buffer.WriteString(jsonEscapedString(unsafeString(ba.Bytes)))
	baPool.Put(ba)

	if bl.logger.newLine {
		bl.buffer.Write([]byte("\"}\n"))
	} else {
		bl.buffer.Write([]byte("\"}"))
	}
	bl.logger.mu.Lock()
	bl.logger.out.Write(bl.buffer.Bytes())
	bl.logger.mu.Unlock()
	bufferLinePool.Put(bl)
}
