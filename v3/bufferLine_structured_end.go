package iqlog

import (
	"fmt"
	"strconv"
)

func (bl *bufferLine) Trace(msg string) {
	if bl.logger.Level > LevelTrace {
		return
	}
	bl.level = LevelTrace
	if bl.logger.jsonMode {
		bl.writeFinalJSON(msg)
	} else {
		bl.writeFinalConsole(msg)
	}
}
func (bl *bufferLine) Tracef(format string, args ...interface{}) {
	if bl.logger.Level > LevelTrace {
		return
	}
	msg := fmt.Sprintf(format, args...)
	bl.Trace(msg)
}

func (bl *bufferLine) Debug(msg string) {
	if bl.logger.Level > LevelDebug {
		return
	}
	bl.level = LevelDebug
	if bl.logger.jsonMode {
		bl.writeFinalJSON(msg)
	} else {
		bl.writeFinalConsole(msg)
	}
}
func (bl *bufferLine) Debugf(format string, args ...interface{}) {
	if bl.logger.Level > LevelDebug {
		return
	}
	msg := fmt.Sprintf(format, args...)
	bl.Debug(msg)
}

func (bl *bufferLine) Info(msg string) {
	if bl.logger.Level > LevelInfo {
		return
	}
	bl.level = LevelInfo
	if bl.logger.jsonMode {
		bl.writeFinalJSON(msg)
	} else {
		bl.writeFinalConsole(msg)
	}
}
func (bl *bufferLine) Infof(format string, args ...interface{}) {
	if bl.logger.Level > LevelInfo {
		return
	}
	msg := fmt.Sprintf(format, args...)
	bl.Info(msg)
}

func (bl *bufferLine) Warn(msg string) {
	if bl.logger.Level > LevelWarn {
		return
	}
	bl.level = LevelWarn
	if bl.logger.jsonMode {
		bl.writeFinalJSON(msg)
	} else {
		bl.writeFinalConsole(msg)
	}
}
func (bl *bufferLine) Warnf(format string, args ...interface{}) {
	if bl.logger.Level > LevelWarn {
		return
	}
	msg := fmt.Sprintf(format, args...)
	bl.Warn(msg)
}

func (bl *bufferLine) Error(msg string) {
	if bl.logger.Level > LevelError {
		return
	}
	bl.level = LevelError
	if bl.logger.jsonMode {
		bl.writeFinalJSON(msg)
	} else {
		bl.writeFinalConsole(msg)
	}
}
func (bl *bufferLine) Errorf(format string, args ...interface{}) {
	if bl.logger.Level > LevelError {
		return
	}
	msg := fmt.Sprintf(format, args...)
	bl.Error(msg)
}

func (bl *bufferLine) Msg(msg string) {
	if bl.logger.Level > bl.level {
		return
	}
	if bl.logger.jsonMode {
		bl.writeFinalJSON(msg)
	} else {
		bl.writeFinalConsole(msg)
	}
}
func (bl *bufferLine) Msgf(format string, args ...interface{}) {
	if bl.logger.Level > bl.level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	bl.Msg(msg)
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

	bl.logger.out.Write(bl.buffer.Bytes())
	bufferLinePool.Put(bl)
}

func (bl *bufferLine) writeFinalJSON(msg string, args ...interface{}) {
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

	if bl.logger.newLine {
		bl.buffer.Write([]byte("\"}\n"))
	} else {
		bl.buffer.Write([]byte("\"}"))
	}
	bl.logger.out.Write(bl.buffer.Bytes())
	bufferLinePool.Put(bl)
}
