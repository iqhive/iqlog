package iqlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"
	"unsafe"
)

var hex = "0123456789abcdef"

func (l *logger) WriteJSONRecord(record LogRecord) error {
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = l.out.Write(b)
	return err
}

// func WriteJSON(dst io.Writer, fields map[string]any) error {
// 	buf := jsonBufPool.Get().(*bytes.Buffer)
// 	buf.Reset()
// 	defer jsonBufPool.Put(buf)

// 	buf.WriteByte('{')

// 	firstField := true
// 	for k, v := range fields {
// 		if !firstField {
// 			buf.WriteByte(',')
// 		}
// 		firstField = false

// 		appendJSONString(buf, k)
// 		buf.WriteByte(':')

// 		switch vv := v.(type) {
// 		case nil:
// 			buf.WriteString("null")
// 		case bool:
// 			if vv {
// 				buf.WriteString("true")
// 			} else {
// 				buf.WriteString("false")
// 			}
// 		case int:
// 			appendIntBuffer(buf, int64(vv))
// 		case int64:
// 			appendIntBuffer(buf, vv)
// 		case float64:
// 			appendFastFloat64(buf, vv)
// 		case float32:
// 			appendFastFloat64(buf, float64(vv))
// 		case string:
// 			appendJSONString(buf, vv)
// 		default:
// 			appendJSONString(buf, valToString(vv))
// 		}
// 	}

// 	buf.WriteByte('}')
// 	_, err := dst.Write(buf.Bytes())
// 	return err
// }

// WriteJSON serializes the LogRecord as JSON without creating an intermediate map.
// It uses record's fields and appends them into a stack-allocated buffer.
func WriteJSON(w io.Writer, record LogRecord, newLine bool) error {
	// For illustration, a 1KB stack buffer. Adjust size depending on your needs.
	var buf [1024]byte
	used := 0

	// Open the JSON object
	buf[used] = '{'
	used++

	// Write out time
	{
		timeStr := record.Time.Format("2006-01-02T15:04:05.999Z07:00")
		used += copy(buf[used:], `"time":"`)
		used += copy(buf[used:], timeStr)
		buf[used] = '"'
		used++
	}

	// Write out level
	{
		used += copy(buf[used:], `,"level":"`)
		// Convert Level to string or numeric. Example as string:
		switch record.Level {
		case LevelDebug:
			used += copy(buf[used:], "debug")
		case LevelInfo:
			used += copy(buf[used:], "info")
		case LevelWarn:
			used += copy(buf[used:], "warn")
		case LevelError:
			used += copy(buf[used:], "error")
		case LevelFatal:
			used += copy(buf[used:], "fatal")
		case LevelPanic:
			used += copy(buf[used:], "panic")
		default:
			used += copy(buf[used:], "unknown")
		}
		buf[used] = '"'
		used++
	}

	// Write out message
	{
		used += copy(buf[used:], `,"message":"`)
		msgLen := record.MsgLen
		if msgLen > len(record.Message) {
			msgLen = len(record.Message)
		}
		used += copy(buf[used:], record.Message[:msgLen])
		buf[used] = '"'
		used++
	}

	// Write out any extra fields stored in record.Fields ( LogField )
	for i := 0; i < record.UsedFields; i++ {
		f := record.Fields[i]
		if !f.Used || f.KeyLen == 0 {
			continue
		}
		used += copy(buf[used:], `,"`)
		used += copy(buf[used:], f.Key[:f.KeyLen])
		used += copy(buf[used:], `":`)
		if f.Quote {
			// If the value is a string and needs quotes:
			buf[used] = '"'
			used++
			used += copy(buf[used:], f.VStr[:f.VLen])
			buf[used] = '"'
			used++
		} else {
			// If the value is numeric/bool/null (no quotes needed):
			used += copy(buf[used:], f.VStr[:f.VLen])
		}
	}

	// Close the JSON object
	buf[used] = '}'
	used++

	if newLine {
		buf[used] = '\n'
		used++
	}

	// Finally, write it out
	_, err := w.Write(buf[:used])
	return err
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

func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func sprintf(format string, args ...any) string {
	// TODO: implement a fast sprintf
	return fmt.Sprintf(format, args...)
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
		appendValue(l, buf, v)
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

// WriteRecord writes out a log record (replacing the old Handle method).
func (l *logger) WriteRecord(ctx context.Context, record LogRecord) error {
	// If JSON mode is enabled, we now produce JSON straight from the record fields,
	// avoiding any map[string] allocations.
	if l.jsonMode {
		return WriteJSON(l.out, record, l.newLine)
	}

	// If the log level is disabled, short-circuit quickly:
	if !l.Enabled(ctx, record.Level) {
		return nil
	}

	// Create a stack buffer of fixed size.  No heap allocations here.
	var out [512]byte
	used := 0

	// -------------------------------------------------------------
	// 1. Time prefix (manually encoded, no calls to .Format)
	// -------------------------------------------------------------
	if l.IncludeTimePrefix {
		t := record.Time
		if t.IsZero() {
			t = time.Now()
		}
		// Example: 2006-01-02T15:04:05.123456   (26 bytes)
		// You can change the width/format as needed
		used += appendTimeRFC3339Micro(t, out[used:])
		out[used] = ' '
		used++
	}

	// -------------------------------------------------------------
	// 2. Level prefix (with optional color)
	// -------------------------------------------------------------
	if l.useColour {
		// ansiColourPrefix returns a []byte constant which reuses a global slice,
		// so it does not allocate at runtime.
		used += copy(out[used:], ansiColourPrefix(record.Level))
	} else {
		// Same idea: just pick a constant []byte for each level
		used += copy(out[used:], levelPrefix(record.Level))
	}

	// -------------------------------------------------------------
	// 3. Optional application name
	// -------------------------------------------------------------
	if l.applicationName != "" {
		out[used] = '['
		used++
		used += copy(out[used:], l.applicationName)
		out[used] = ']'
		used++
		out[used] = ' '
		used++
	}

	// -------------------------------------------------------------
	// 4. The message
	// -------------------------------------------------------------
	msgLen := record.MsgLen
	if msgLen > maxStringLen {
		msgLen = maxStringLen
	}
	copy(out[used:], record.Message[:msgLen])
	used += msgLen

	// -------------------------------------------------------------
	// 5. End with a newline (optional)
	// -------------------------------------------------------------
	if l.newLine {
		out[used] = '\n'
		used++
	}

	// -------------------------------------------------------------
	// 6. Write the data to l.out
	// -------------------------------------------------------------
	if asyncWr, ok := l.out.(*asyncWriter); ok {
		// For async, copy out to a pooled bytes.Buffer, then enqueue:
		lineCopy := bufferPool.Get().(*bytes.Buffer)
		lineCopy.Reset()
		lineCopy.Write(out[:used])
		asyncWr.ch <- lineCopy
		return nil
	}

	_, err := l.out.Write(out[:used])
	return err
}

// appendIntBuffer is a small helper for appending decimal integers.
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

// appendIntBytes writes a zero-padded integer of given width, returning how many bytes were written.
// For example, appendInt(..., 5, 2) writes "05".
func appendIntBytes(dst []byte, val int64, width int) int {
	neg := val < 0
	if neg {
		val = -val
	}
	// Build the digits in a temporary buffer.
	var tmp [20]byte
	i := len(tmp)
	for val > 0 {
		i--
		tmp[i] = byte('0' + (val % 10))
		val /= 10
	}
	// If nothing was written, write "0".
	if i == len(tmp) {
		i--
		tmp[i] = '0'
	}
	// Zero-padding to meet width:
	numLen := len(tmp) - i
	for pad := width - numLen; pad > 0; pad-- {
		i--
		tmp[i] = '0'
	}
	// If negative, append '-'.
	if neg {
		i--
		tmp[i] = '-'
	}
	n := copy(dst, tmp[i:])
	return n
}
