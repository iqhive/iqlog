package iqlog

import "fmt"

func (bsl *bytesliceLine) Int(name string, val int) *bytesliceLine {
	if bsl.output == nil {
		return bsl
	}
	if bsl.jsonMode {
		bsl.output = append(bsl.output, []byte(",\""+name+"\":")...)
		bsl.output, _ = appendIntDecimal(bsl.output, int64(val))
	} else {
		bsl.output = append(bsl.output, []byte(name+"=")...)
		bsl.output, _ = appendIntDecimal(bsl.output, int64(val))
		bsl.output = append(bsl.output, ' ')
	}
	return bsl
}

func (bsl *bytesliceLine) Int64(name string, val int64) *bytesliceLine {
	if bsl.output == nil {
		return bsl
	}
	if bsl.jsonMode {
		bsl.output = append(bsl.output, []byte(",\""+name+"\":")...)
		bsl.output, _ = appendIntDecimal(bsl.output, int64(val))
	} else {
		bsl.output = append(bsl.output, []byte(name+"=")...)
		bsl.output, _ = appendIntDecimal(bsl.output, int64(val))
		bsl.output = append(bsl.output, ' ')
	}
	return bsl
}

func (bsl *bytesliceLine) Str(name string, s string) *bytesliceLine {
	if bsl.output == nil {
		return bsl
	}
	if bsl.jsonMode {
		bsl.output = append(bsl.output, ',', '"')
		bsl.output = appendJSONEscaped(bsl.output, name)
		bsl.output = append(bsl.output, '"', ':', '"')
		bsl.output = appendJSONEscaped(bsl.output, s)
		bsl.output = append(bsl.output, '"')
	} else {
		bsl.output = append(bsl.output, []byte(name+"="+s+" ")...)
	}
	return bsl
}

func (bsl *bytesliceLine) Float32(name string, f float32) *bytesliceLine {
	if bsl.output == nil {
		return bsl
	}
	if bsl.jsonMode {
		bsl.output = append(bsl.output, []byte(",\""+name+"\":")...)
		bsl.output, _ = appendfastFloatFill(bsl.output, float64(f), 6)
	} else {
		bsl.output = append(bsl.output, []byte(name+"=")...)
		bsl.output, _ = appendfastFloatFill(bsl.output, float64(f), 6)
		bsl.output = append(bsl.output, []byte(" ")...)
	}
	return bsl
}

func (bsl *bytesliceLine) Float64(name string, f float64) *bytesliceLine {
	if bsl.output == nil {
		return bsl
	}
	if bsl.jsonMode {
		bsl.output = append(bsl.output, []byte(",\""+name+"\":")...)
		bsl.output, _ = appendfastFloatFill(bsl.output, float64(f), 6)
	} else {
		bsl.output = append(bsl.output, []byte(name+"=")...)
		bsl.output, _ = appendfastFloatFill(bsl.output, float64(f), 6)
		bsl.output = append(bsl.output, []byte(" ")...)
	}
	return bsl
}

func (bsl *bytesliceLine) Bool(name string, b bool) *bytesliceLine {
	if bsl.output == nil {
		return bsl
	}
	if bsl.jsonMode {
		if b {
			bsl.output = append(bsl.output, []byte(",\""+name+"\":true")...)
		} else {
			bsl.output = append(bsl.output, []byte(",\""+name+"\":false")...)
		}
	} else {
		if b {
			bsl.output = append(bsl.output, []byte(name+"=true ")...)
		} else {
			bsl.output = append(bsl.output, []byte(name+"=false ")...)
		}
	}
	return bsl
}

func (bsl *bytesliceLine) Any(name string, v any) *bytesliceLine {
	if bsl.output == nil {
		return bsl
	}
	str := fmt.Sprintf("%v", v)
	if bsl.jsonMode {
		bsl.output = append(bsl.output, ',', '"')
		bsl.output = appendJSONEscaped(bsl.output, name)
		bsl.output = append(bsl.output, '"', ':', '"')
		bsl.output = appendJSONEscaped(bsl.output, str)
		bsl.output = append(bsl.output, '"')
	} else {
		bsl.output = append(bsl.output, []byte(name+"="+str+" ")...)
	}
	return bsl
}
