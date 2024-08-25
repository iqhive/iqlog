package iqlog

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"unsafe"
)

var hex = "0123456789abcdef"

func WriteJSON(dst io.Writer, fields map[string]any) error {
	buf := bytes.NewBuffer(nil)

	buf.WriteByte('{')

	firstField := true
	for k, v := range fields {
		if !firstField {
			buf.WriteByte(',')
		}
		firstField = false

		appendJSONString(buf, k)
		buf.WriteByte(':')

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

	buf.WriteByte('}')
	buf.WriteByte('\n')
	_, err := dst.Write(buf.Bytes())
	return err
}

func writeRecordAsJSON(l *logger, buf *bytes.Buffer, record LogRecord, newLine bool) {
	buf.WriteByte('{')

	firstField := true

	if !record.Time.IsZero() {
		firstField = appendJSONField(buf, firstField, "time")
		appendJSONString(buf, record.Time.Format(l.TimePrefixFormat))
	}

	firstField = appendJSONField(buf, firstField, "level")
	appendIntBuffer(buf, int64(record.Level))

	if record.MsgLen > 0 {
		firstField = appendJSONField(buf, firstField, "message")
		appendJSONString(buf, unsafeString(record.Message[:record.MsgLen]))
	}

	l.logFields.ForEach(func(k string, v any) {
		firstField = appendJSONField(buf, firstField, k)
		appendJSONValue(l, buf, v)
	})

	for _, f := range record.Fields {
		if !f.Used {
			continue
		}
		firstField = appendJSONField(buf, firstField, unsafeString(f.Key[:f.KeyLen]))
		if f.Quote {
			appendJSONString(buf, unsafeString(f.VStr[:f.VLen]))
		} else {
			buf.Write(f.VStr[:f.VLen])
		}
	}

	buf.WriteByte('}')
	if newLine {
		buf.WriteByte('\n')
	}
}

func appendJSONValue(l *logger, buf *bytes.Buffer, v any) {
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

func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

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

// appendIntBuffer is a small helper for appending decimal integers
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
	b = fastFloatDecimal(b, f, 6)
	dst.Write(b)
	return nil
}

func fastFloatDecimal(b []byte, f float64, dec int) []byte {
	if f < 0 {
		b = append(b, '-')
		f = -f
	}
	intPart := int64(f)
	b = appendIntDecimal(b, intPart)
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

func appendIntDecimal(b []byte, i int64) []byte {
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

func sprintf(format string, args ...any) string {
	// TODO: implement sprintf ??
	return fmt.Sprintf(format, args...)
}
