package iqlog

import (
	"bytes"
	"fmt"
	"math"
	"unsafe"
)

var hex = "0123456789abcdef"

func appendJSONString(dst *bytes.Buffer, s string) {
	dst.WriteByte('"')
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == '\\' || c == '"' {
			if i > start {
				dst.WriteString(s[start:i])
			}
			switch c {
			case '\\', '"':
				dst.WriteByte('\\')
				dst.WriteByte(c)
			case '\n':
				dst.WriteString(`\n`)
			case '\r':
				dst.WriteString(`\r`)
			case '\t':
				dst.WriteString(`\t`)
			default:
				dst.WriteString(`\u00`)
				dst.WriteByte(hex[c>>4])
				dst.WriteByte(hex[c&0x0f])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		dst.WriteString(s[start:])
	}
	dst.WriteByte('"')
}

func appendFastFloat64(dst *bytes.Buffer, f float64) error {
	if math.IsNaN(f) {
		dst.WriteString(`"NaN"`)
		return nil
	}
	if math.IsInf(f, 1) {
		dst.WriteString(`"Infinity"`)
		return nil
	}
	if math.IsInf(f, -1) {
		dst.WriteString(`"-Infinity"`)
		return nil
	}

	if f == float64(int64(f)) && f <= math.MaxInt64 && f >= math.MinInt64 {
		appendIntBuffer(dst, int64(f))
		return nil
	}

	var scratch [64]byte
	b := scratch[:0]
	b = setfastFloatDecimal(b, f, 6)
	dst.Write(b)
	return nil
}

func setfastFloatDecimal(b []byte, f float64, dec int) []byte {
	if f < 0 {
		b = append(b, '-')
		f = -f
	}
	intPart := int64(f)
	b = setIntDecimal(b, intPart)
	f -= float64(intPart)

	if f == 0.0 {
		return b
	}
	b = append(b, '.')

	for i := 0; i < dec; i++ {
		f *= 10
		digit := int64(f)
		b = append(b, byte('0'+digit))
		f -= float64(digit)
		if f == 0.0 {
			break
		}
	}
	return b
}

func setIntDecimal(b []byte, i int64) []byte {
	var tmp [20]byte
	pos := len(tmp)
	if i == 0 {
		pos--
		tmp[pos] = '0'
	} else {
		for i > 0 {
			pos--
			tmp[pos] = byte('0' + i%10)
			i /= 10
		}
	}
	return append(b, tmp[pos:]...)
}

func valToString(v any) string {
	return unsafeString([]byte(sprintf("%v", v)))
}

func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func sprintf(format string, args ...any) string {
	// TODO: implement a fast sprintf
	return fmt.Sprintf(format, args...)
}

func appendJSONField(buf *bytes.Buffer, firstField bool, key string) bool {
	if !firstField {
		buf.WriteByte(',')
	} else {
		firstField = false
	}
	appendJSONString(buf, key)
	buf.WriteByte(':')
	return firstField
}

func appendValue(l *logger, buf *bytes.Buffer, v any) {
	switch vv := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if vv {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case int:
		appendIntBuffer(buf, int64(vv))
	case int64:
		appendIntBuffer(buf, vv)
	case float64:
		appendFastFloat64(buf, vv)
	case float32:
		appendFastFloat64(buf, float64(vv))
	case string:
		appendJSONString(buf, vv)
	default:
		appendJSONString(buf, valToString(vv))
	}
}

// appendIntBuffer appends decimal integers
func appendIntBuffer(dst *bytes.Buffer, i int64) {
	var b [20]byte
	pos := len(b)
	neg := i < 0
	if neg {
		i = -i
	}
	if i == 0 {
		pos--
		b[pos] = '0'
	}
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	dst.Write(b[pos:])
}

// setIntBytes writes a zero-padded integer of given width, returning how many bytes were written.
// eg setIntBytes(..., 5, 2) writes "05".
func setIntBytes(dst []byte, val int64, width int) int {
	neg := val < 0
	if neg {
		val = -val
	}
	// Build the digits in temp buffer
	var tmp [20]byte
	i := len(tmp)
	for val > 0 {
		i--
		tmp[i] = byte('0' + (val % 10))
		val /= 10
	}
	// If nothing was written, write "0"
	if i == len(tmp) {
		i--
		tmp[i] = '0'
	}
	// Zero-padding to meet width
	numLen := len(tmp) - i
	for pad := width - numLen; pad > 0; pad-- {
		i--
		tmp[i] = '0'
	}
	// If negative, append '-'
	if neg {
		i--
		tmp[i] = '-'
	}
	n := copy(dst, tmp[i:])
	return n
}
