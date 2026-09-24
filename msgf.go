package iqlog

import (
	"fmt"
	"strconv"
	"strings"
)

// appendFormatted appends the text fmt.Sprintf(format, args...) would
// produce, directly into dst. The plain verbs almost every log line uses are
// rendered without fmt; everything else is handed to fmt.Appendf, which
// still writes straight into dst rather than through a scratch buffer.
func appendFormatted(dst []byte, format string, args []any) []byte {
	start := len(dst)
	if out, ok := appendSimpleFormat(dst, format, args); ok {
		return out
	}
	// The fast path may have appended part of the text before declining, so
	// fmt starts over from where the message began. fmt owns every case the
	// fast path does not: flags, width, precision, argument indexes, the
	// remaining verbs, methods and reflection, and the %! diagnostic forms.
	return fmt.Appendf(dst[:start], format, args...)
}

// appendSimpleFormat is the fast path behind Msgf. It handles the verbs
// %d, %s, %v, %t and %% with no flag, width, precision or argument index,
// applied to the basic types fmt prints for those verbs without consulting
// methods or reflection: the integer types, string and bool. For that subset
// fmt's output is fixed -- the decimal digits, the string verbatim, "true" or
// "false" -- and is reproduced here with the strconv routines fmt itself
// uses, so the record skips fmt's printer state machine, its pooled printer
// and its intermediate buffer.
//
// It reports false, possibly after appending part of the text, for anything
// outside that subset: a flag or width byte after the %, an unknown verb, a
// lone % at the end, an argument whose dynamic type is not one the verb
// prints verbatim (including nil and every named or composite type), a
// missing argument, or a surplus one. The caller then truncates and runs the
// whole format through fmt, which is the only way to keep fmt's exact
// diagnostics for those cases.
func appendSimpleFormat(dst []byte, format string, args []any) ([]byte, bool) {
	next := 0 // index of the next argument to consume
	for {
		i := strings.IndexByte(format, '%')
		if i < 0 {
			break
		}
		dst = append(dst, format[:i]...)
		if i+1 == len(format) {
			return dst, false // fmt writes %!(NOVERB)
		}
		verb := format[i+1]
		format = format[i+2:]
		if verb == '%' {
			dst = append(dst, '%')
			continue
		}
		if next == len(args) {
			return dst, false // fmt writes %!d(MISSING)
		}
		var ok bool
		if dst, ok = appendSimpleVerb(dst, verb, args[next]); !ok {
			return dst, false
		}
		next++
	}
	if next != len(args) {
		return dst, false // fmt appends %!(EXTRA ...)
	}
	return append(dst, format...), true
}

// appendSimpleVerb appends arg as fmt renders it for verb when arg's dynamic
// type is one of the basic types that verb prints without flags. The type
// switch matches exact types only, as fmt's own does, so a named integer or
// string type falls through to fmt, which may find a String or Format method
// on it.
func appendSimpleVerb(dst []byte, verb byte, arg any) ([]byte, bool) {
	switch verb {
	case 'd', 'v':
		switch v := arg.(type) {
		case int:
			return strconv.AppendInt(dst, int64(v), 10), true
		case int64:
			return strconv.AppendInt(dst, v, 10), true
		case int32:
			return strconv.AppendInt(dst, int64(v), 10), true
		case int16:
			return strconv.AppendInt(dst, int64(v), 10), true
		case int8:
			return strconv.AppendInt(dst, int64(v), 10), true
		case uint:
			return strconv.AppendUint(dst, uint64(v), 10), true
		case uint64:
			return strconv.AppendUint(dst, v, 10), true
		case uint32:
			return strconv.AppendUint(dst, uint64(v), 10), true
		case uint16:
			return strconv.AppendUint(dst, uint64(v), 10), true
		case uint8:
			return strconv.AppendUint(dst, uint64(v), 10), true
		case string:
			if verb == 'v' {
				return append(dst, v...), true
			}
		case bool:
			if verb == 'v' {
				return strconv.AppendBool(dst, v), true
			}
		}
	case 's':
		if v, ok := arg.(string); ok {
			return append(dst, v...), true
		}
	case 't':
		if v, ok := arg.(bool); ok {
			return strconv.AppendBool(dst, v), true
		}
	}
	return dst, false
}
