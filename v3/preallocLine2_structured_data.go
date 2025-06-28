package iqlog

import "fmt"

func (pal *preallocLine2) Int(name string, val int) *preallocLine2 {
	if pal.output == nil {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += writeIntDecimal(pal.output[pal.bytesUsed:], int64(val))
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "=")
		// fmt.Printf("Int output: (%d/%d) |%s|\n", pal.bytesUsed, len(pal.output), string(pal.output))
		pal.bytesUsed += writeIntDecimal(pal.output[pal.bytesUsed:], int64(val))
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine2) Int64(name string, val int64) *preallocLine2 {
	if pal.output == nil {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += writeIntDecimal(pal.output[pal.bytesUsed:], val)
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += writeIntDecimal(pal.output[pal.bytesUsed:], val)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine2) Str(name string, s string) *preallocLine2 {
	if pal.output == nil {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\":\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, s)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"")
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, s)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine2) Float32(name string, f float32) *preallocLine2 {
	if pal.output == nil {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += fastFloatFill(pal.output[pal.bytesUsed:], float64(f), 6)
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += fastFloatFill(pal.output[pal.bytesUsed:], float64(f), 6)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine2) Float64(name string, f float64) *preallocLine2 {
	if pal.output == nil {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += fastFloatFill(pal.output[pal.bytesUsed:], f, 6)
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += fastFloatFill(pal.output[pal.bytesUsed:], f, 6)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine2) Bool(name string, b bool) *preallocLine2 {
	if pal.output == nil {
		return pal
	}
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\":")
		if b {
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "true")
		} else {
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "false")
		}
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "=")
		if b {
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "true")
		} else {
			pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "false")
		}
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")
	}
	return pal
}

func (pal *preallocLine2) Any(name string, v any) *preallocLine2 {
	if pal.output == nil {
		return pal
	}
	str := fmt.Sprintf("%v", v)
	if pal.jsonMode {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, ",\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\":")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, str)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "\"")
	} else {
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, name)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, "=")
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, str)
		pal.bytesUsed += safeOutputCopy(pal.output, pal.bytesUsed, " ")
	}
	return pal
}
