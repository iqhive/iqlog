package iqlog

import "fmt"

func (bl *bufferLineNL) Int(name string, val int) *bufferLineNL {
	if bl.buffer == nil {
		return bl
	}
	if bl.jsonMode {
		bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":"))
		appendBufferIntDecimal(bl.buffer, int64(val))
	} else {
		bl.buffer.Write([]byte(name + "="))
		appendBufferIntDecimal(bl.buffer, int64(val))
		bl.buffer.WriteByte(' ')
	}
	return bl
}

func (bl *bufferLineNL) Int64(name string, val int64) *bufferLineNL {
	if bl.buffer == nil {
		return bl
	}
	if bl.jsonMode {
		bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":"))
		appendBufferIntDecimal(bl.buffer, int64(val))
	} else {
		bl.buffer.Write([]byte(name + "="))
		appendBufferIntDecimal(bl.buffer, int64(val))
		bl.buffer.WriteByte(' ')
	}
	return bl
}

func (bl *bufferLineNL) Str(name string, s string) *bufferLineNL {
	if bl.buffer == nil {
		return bl
	}
	if bl.jsonMode {
		bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":\"" + jsonEscapedString(s) + "\""))
	} else {
		bl.buffer.Write([]byte(name + "=" + s + " "))
	}
	return bl
}

func (bl *bufferLineNL) Float32(name string, f float32) *bufferLineNL {
	if bl.buffer == nil {
		return bl
	}
	if bl.jsonMode {
		bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":"))
		appendBufferfastFloatFill(bl.buffer, float64(f), 6)
	} else {
		bl.buffer.Write([]byte(name + "="))
		appendBufferfastFloatFill(bl.buffer, float64(f), 6)
		bl.buffer.WriteByte(' ')
	}
	return bl
}

func (bl *bufferLineNL) Float64(name string, f float64) *bufferLineNL {
	if bl.buffer == nil {
		return bl
	}
	if bl.jsonMode {
		bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":"))
		appendBufferfastFloatFill(bl.buffer, float64(f), 6)
	} else {
		bl.buffer.Write([]byte(name + "="))
		appendBufferfastFloatFill(bl.buffer, float64(f), 6)
		bl.buffer.WriteByte(' ')
	}
	return bl
}

func (bl *bufferLineNL) Bool(name string, b bool) *bufferLineNL {
	if bl.buffer == nil {
		return bl
	}
	if bl.jsonMode {
		if b {
			bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":true"))
		} else {
			bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":false"))
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

func (bl *bufferLineNL) Any(name string, v any) *bufferLineNL {
	if bl.buffer == nil {
		return bl
	}
	str := fmt.Sprintf("%v", v)
	if bl.jsonMode {
		bl.buffer.Write([]byte(",\"" + jsonEscapedString(name) + "\":\"" + jsonEscapedString(str) + "\""))
	} else {
		bl.buffer.Write([]byte(name + "=\"" + str + "\" "))
	}
	return bl
}
