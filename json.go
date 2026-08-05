package iqlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"unsafe"
)

// maxExactInt64Float is 2^63; float64 values at or beyond this magnitude
// cannot be converted to int64 safely.
const maxExactInt64Float = float64(1 << 63)

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

// appendJSONEscaped appends s to dst with JSON string escaping
// (without surrounding quotes).
func appendJSONEscaped(dst []byte, s string) []byte {
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == '\\' || c == '"' {
			if i > start {
				dst = append(dst, s[start:i]...)
			}
			switch c {
			case '\\', '"':
				dst = append(dst, '\\', c)
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			default:
				dst = append(dst, '\\', 'u', '0', '0', hex[c>>4], hex[c&0x0f])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		dst = append(dst, s[start:]...)
	}
	return dst
}

// repairTruncatedJSONLine repairs a partially assembled JSON object that hit
// the fixed line-buffer limit, so the emitted record stays parseable. It
// walks back from the end looking for the longest prefix that forms a valid
// object once closed (with `"}` when the cut lands inside a string, or `}`
// otherwise) and returns the new length. Only called on the rare truncation
// path, so validating candidates with json.Valid is acceptable.
func repairTruncatedJSONLine(buf []byte, used, max int) int {
	if used > max {
		used = max
	}
	for cut := used; cut > 1; cut-- {
		for _, closer := range []string{"\"}", "}"} {
			if cut+len(closer) > max {
				continue
			}
			cand := make([]byte, 0, cut+len(closer))
			cand = append(cand, buf[:cut]...)
			cand = append(cand, closer...)
			if json.Valid(cand) {
				copy(buf[cut:], closer)
				return cut + len(closer)
			}
		}
	}
	// give up: emit an empty object rather than an invalid record
	copy(buf, "{}")
	return 2
}

// appendJSONFloat appends f as a JSON-safe value: quoted for non-finite
// values (bare NaN/Inf are not valid JSON), decimal otherwise.
func appendJSONFloat(dst []byte, f float64, bitSize int) []byte {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return append(dst, fallbackFloatString(f)...)
	}
	return append(dst, strconv.FormatFloat(f, 'f', 6, bitSize)...)
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

	if f >= maxExactInt64Float || f <= -maxExactInt64Float {
		// Out of int64 range: the fast integer/decimal paths would overflow,
		// so fall back to the standard library formatter.
		dst.WriteString(strconv.FormatFloat(f, 'f', -1, 64))
		return nil
	}

	if f == float64(int64(f)) {
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

// jsonEscapedString returns s with JSON string escaping applied.
// It returns s unchanged when no escaping is needed.
func jsonEscapedString(s string) string {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == '\\' || c == '"' {
			return string(appendJSONEscaped(make([]byte, 0, len(s)+8), s))
		}
	}
	return s
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
	u := uint64(i)
	if neg {
		u = -u
	}
	if u == 0 {
		pos--
		b[pos] = '0'
	}
	for u > 0 {
		pos--
		b[pos] = byte('0' + u%10)
		u /= 10
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
	u := uint64(val)
	if neg {
		u = -u
	}
	// Build the digits in temp buffer
	var tmp [20]byte
	i := len(tmp)
	for u > 0 {
		i--
		tmp[i] = byte('0' + (u % 10))
		u /= 10
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
