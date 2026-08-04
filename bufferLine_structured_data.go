package iqlog

import "fmt"

func (bl *bufferLine) Int(name string, val int) *bufferLine {
	if bl.buffer == nil {
		return bl
	}
	if bl.logger.jsonMode {
		bl.buffer.Write([]byte(",\"" + name + "\":"))
		appendBufferIntDecimal(bl.buffer, int64(val))
	} else {
		bl.buffer.Write([]byte(name + "="))
		appendBufferIntDecimal(bl.buffer, int64(val))
		bl.buffer.WriteByte(' ')
	}
	return bl
}

func (bl *bufferLine) Int64(name string, val int64) *bufferLine {
	if bl.buffer == nil {
		return bl
	}
	if bl.logger.jsonMode {
		bl.buffer.Write([]byte(",\"" + name + "\":"))
		appendBufferIntDecimal(bl.buffer, int64(val))
	} else {
		bl.buffer.Write([]byte(name + "="))
		appendBufferIntDecimal(bl.buffer, int64(val))
		bl.buffer.WriteByte(' ')
	}
	return bl
}

func (bl *bufferLine) Str(name string, s string) *bufferLine {
	if bl.buffer == nil {
		return bl
	}
	if bl.logger.jsonMode {
		bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":\"" + jsonEscapedString(s) + "\""))
	} else {
		bl.buffer.Write([]byte(name + "=" + s + " "))
	}
	return bl
}

func (bl *bufferLine) Float32(name string, f float32) *bufferLine {
	if bl.buffer == nil {
		return bl
	}
	if bl.logger.jsonMode {
		bl.buffer.Write([]byte(",\"" + name + "\":"))
		appendBufferfastFloatFill(bl.buffer, float64(f), 6)
	} else {
		bl.buffer.Write([]byte(name + "="))
		appendBufferfastFloatFill(bl.buffer, float64(f), 6)
		bl.buffer.WriteByte(' ')
	}
	return bl
}

func (bl *bufferLine) Float64(name string, f float64) *bufferLine {
	if bl.buffer == nil {
		return bl
	}
	if bl.logger.jsonMode {
		bl.buffer.Write([]byte(",\"" + name + "\":"))
		appendBufferfastFloatFill(bl.buffer, float64(f), 6)
	} else {
		bl.buffer.Write([]byte(name + "="))
		appendBufferfastFloatFill(bl.buffer, float64(f), 6)
		bl.buffer.WriteByte(' ')
	}
	return bl
}

func (bl *bufferLine) Bool(name string, b bool) *bufferLine {
	if bl.buffer == nil {
		return bl
	}
	if bl.logger.jsonMode {
		if b {
			bl.buffer.Write([]byte(",\"" + name + "\":true"))
		} else {
			bl.buffer.Write([]byte(",\"" + name + "\":false"))
		}
	} else {
		if b {
			bl.buffer.Write([]byte(name + "=true "))
		} else {
			bl.buffer.Write([]byte(name + "=false "))
		}
	}
	return bl
}

func (bl *bufferLine) Any(name string, v any) *bufferLine {
	if bl.buffer == nil {
		return bl
	}
	str := fmt.Sprintf("%v", v)
	if bl.logger.jsonMode {
		bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":\"" + jsonEscapedString(str) + "\""))
	} else {
		bl.buffer.Write([]byte(name + "=\"" + str + "\" "))
	}
	return bl
}
