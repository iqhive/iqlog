package iqlog

import (
	"math"
	"strconv"
	"unicode/utf8"
	"unsafe"
)

var hex = "0123456789abcdef"

// jsonEscapeTable marks the bytes that cannot be copied into a JSON string
// verbatim: the C0 control range, the quote and backslash, and every byte
// >= 0x80, which has to be validated as UTF-8 first. jsonKeyNeedsEscaping
// uses it because one indexed load per byte measured faster there than the
// equivalent comparisons; appendJSONEscaped uses comparisons instead, where
// they measured faster.
var jsonEscapeTable = func() (t [256]bool) {
	for c := 0; c < 0x20; c++ {
		t[c] = true
	}
	t['"'] = true
	t['\\'] = true
	for c := utf8.RuneSelf; c < 256; c++ {
		t[c] = true
	}
	return t
}()

// appendJSONEscaped appends s to dst with JSON string escaping
// (without surrounding quotes).
//
// Bytes that are not part of a well-formed UTF-8 sequence are replaced with
// U+FFFD, matching encoding/json. JSON strings must be valid UTF-8, and
// emitting the raw bytes produces records that strict parsers reject.
func appendJSONEscaped(dst []byte, s string) []byte {
	for {
		start := 0
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			// c-0x20 wraps, so this one unsigned compare rejects both the C0
			// range and every byte >= 0x80 in a single test, keeping the
			// ASCII fast path at the three comparisons it had before UTF-8
			// validation was added.
			if c-0x20 < 0x60 && c != '"' && c != '\\' {
				continue
			}
			if i > start {
				dst = append(dst, s[start:i]...)
			}
			if c >= utf8.RuneSelf {
				// hand the multi-byte run to the rune loop below
				break
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
		if i == len(s) {
			if start < len(s) {
				dst = append(dst, s[start:]...)
			}
			return dst
		}
		var n int
		dst, n = appendJSONEscapedRunes(dst, s[i:])
		s = s[i+n:]
	}
}

// appendJSONEscapedRunes appends the leading run of non-ASCII bytes of s,
// replacing ill-formed sequences with U+FFFD, and reports how many bytes it
// consumed. It stops at the first ASCII byte so the caller's ASCII loop takes
// over again; keeping rune decoding out of that loop is what makes the common
// all-ASCII record cost nothing for UTF-8 correctness.
func appendJSONEscapedRunes(dst []byte, s string) ([]byte, int) {
	i := 0
	for i < len(s) && s[i] >= utf8.RuneSelf {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			dst = append(dst, '\\', 'u', 'f', 'f', 'f', 'd')
		} else {
			dst = append(dst, s[i:i+size]...)
		}
		i += size
	}
	return dst, i
}

// escapeJSONTail re-encodes dst[from:] with JSON string escaping when it holds
// bytes that would break out of the string being built. It is for envelope
// text that came from configuration rather than from a field encoder, which
// escapes as it writes.
func escapeJSONTail(dst []byte, from int) []byte {
	if !jsonKeyNeedsEscaping(unsafeString(dst[from:])) {
		return dst
	}
	tail := string(dst[from:])
	return appendJSONEscaped(dst[:from], tail)
}

// appendJSONFloat appends f as a JSON-safe value: quoted for non-finite
// values (bare NaN/Inf are not valid JSON), decimal otherwise.
func appendJSONFloat(dst []byte, f float64, bitSize int) []byte {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return append(dst, fallbackFloatString(f)...)
	}
	return strconv.AppendFloat(dst, f, 'f', -1, bitSize)
}

// appendJSONStringFloat appends f as text suitable for embedding inside a
// JSON string (e.g. a vararg inside the "message" field). Unlike
// appendJSONFloat, non-finite values are emitted as bare escaped text
// (NaN/Infinity/-Infinity) rather than quoted JSON tokens, so the surrounding
// string stays parseable.
func appendJSONStringFloat(dst []byte, f float64, bitSize int) []byte {
	var s string
	switch {
	case math.IsNaN(f):
		s = "NaN"
	case math.IsInf(f, 1):
		s = "Infinity"
	case math.IsInf(f, -1):
		s = "-Infinity"
	default:
		return append(dst, strconv.FormatFloat(f, 'f', 6, bitSize)...)
	}
	return appendJSONEscaped(dst, s)
}

// unsafeString views b as a string without copying. The result must not
// outlive b, and b must not be mutated while it is in use.
func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
