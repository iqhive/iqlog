package iqlog

import "fmt"

func (pal *preallocLine) Int(name string, val int) *preallocLine {
	if pal.bytesUsed == 0 {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(name))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += writeIntDecimal(pal.output[pal.bytesUsed:], int64(val))
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += writeIntDecimal(pal.output[pal.bytesUsed:], int64(val))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine) Int64(name string, val int64) *preallocLine {
	if pal.bytesUsed == 0 {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(name))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += writeIntDecimal(pal.output[pal.bytesUsed:], val)
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += writeIntDecimal(pal.output[pal.bytesUsed:], val)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine) Str(name string, s string) *preallocLine {
	if pal.bytesUsed == 0 {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(name))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\":\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(s))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"")
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, s)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine) Float32(name string, f float32) *preallocLine {
	if pal.bytesUsed == 0 {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(name))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += fastFloatFill(pal.output[pal.bytesUsed:], float64(f), 6)
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += fastFloatFill(pal.output[pal.bytesUsed:], float64(f), 6)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine) Float64(name string, f float64) *preallocLine {
	if pal.bytesUsed == 0 {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(name))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += fastFloatFill(pal.output[pal.bytesUsed:], f, 6)
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += fastFloatFill(pal.output[pal.bytesUsed:], f, 6)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine) Bool(name string, b bool) *preallocLine {
	if pal.bytesUsed == 0 {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(name))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\":")
		if b {
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "true")
		} else {
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "false")
		}
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "=")
		if b {
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "true")
		} else {
			pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "false")
		}
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine) Any(name string, v any) *preallocLine {
	if pal.bytesUsed == 0 {
		return pal
	}
	str := fmt.Sprintf("%v", v)
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(name))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, jsonEscapedString(str))
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "\"")
	} else {
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, str)
		pal.bytesUsed += safeOutputCopyMaxLineLen(pal.output, pal.bytesUsed, " ")
	}
	return pal
}
