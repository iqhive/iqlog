package iqlog

import (
	"encoding/binary"
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

// jsonWordEscapeMask is nonzero when any of the eight bytes packed in w is a
// byte appendJSONEscaped cannot copy verbatim: a C0 control byte, a quote, a
// backslash, or a byte >= 0x80. Byte order does not matter, every lane is
// tested alike, and the result is a lane mask so two words can be combined
// with one OR before the branch.
//
// Each term is the usual SWAR test. A borrow can spill from a matching lane
// into the lane above and mark it too, but only after a genuine match, so the
// any-lane result is exact.
func jsonWordEscapeMask(w uint64) uint64 {
	control := (w - 0x20*swarLo) &^ w
	quote := w ^ ('"' * swarLo)
	quote = (quote - swarLo) &^ quote
	backslash := w ^ ('\\' * swarLo)
	backslash = (backslash - swarLo) &^ backslash
	return (control | quote | backslash | w) & swarHi
}

// jsonCleanPrefix returns the length of a leading part of s that can be
// copied into a JSON string verbatim, which is len(s) when the whole string
// can. A string shorter than a word is tested by byte. A longer one is
// tested two words at a time, then one; a word that needs work stops the
// scan at its start. The final partial word is tested as the string's last
// eight bytes: the bytes that overlap the previous word are already known to
// be clean, so this is exact and replaces up to seven byte tests.
func jsonCleanPrefix(s string) int {
	if len(s) < 8 {
		for i := 0; i < len(s); i++ {
			if jsonEscapeTable[s[i]] {
				return i
			}
		}
		return len(s)
	}
	// A read-only byte view of s, so each word is one unaligned load. The
	// view is consumed by reslicing, the form whose bounds checks the
	// compiler can prove away.
	all := unsafe.Slice(unsafe.StringData(s), len(s))
	b := all
	i := 0
	for len(b) >= 16 {
		if jsonWordEscapeMask(binary.LittleEndian.Uint64(b))|jsonWordEscapeMask(binary.LittleEndian.Uint64(b[8:])) != 0 {
			break
		}
		b = b[16:]
		i += 16
	}
	for len(b) >= 8 {
		if jsonWordEscapeMask(binary.LittleEndian.Uint64(b)) != 0 {
			return i
		}
		b = b[8:]
		i += 8
	}
	if len(b) != 0 && jsonWordEscapeMask(binary.LittleEndian.Uint64(all[len(all)-8:])) != 0 {
		return i
	}
	return len(s)
}

// appendJSONEscaped appends s to dst with JSON string escaping
// (without surrounding quotes).
//
// Bytes that are not part of a well-formed UTF-8 sequence are replaced with
// U+FFFD, matching encoding/json. JSON strings must be valid UTF-8, and
// emitting the raw bytes produces records that strict parsers reject.
//
// Most strings need no escaping at all: the message of every record and
// nearly every value. They are found with the word scan of jsonCleanPrefix
// and copied with one append. A string that does need work has its clean
// lead copied the same way, and the byte loop takes over from there, so the
// two paths produce identical output.
func appendJSONEscaped(dst []byte, s string) []byte {
	clean := jsonCleanPrefix(s)
	if clean == len(s) {
		return append(dst, s...)
	}
	if clean > 0 {
		dst = append(dst, s[:clean]...)
		s = s[clean:]
	}
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
// bytes that would break out of the string being built. It is for text that
// was appended to the record before it could be escaped -- envelope text that
// came from configuration, or a formatted message -- rather than for field
// values, whose encoders escape as they write.
//
// The tail is copied through a pooled scratch buffer before it is re-encoded
// because the escaped form is longer than the original and would overwrite
// bytes not yet read if written over it in place. A message that needs
// escaping therefore costs one extra copy but no allocation.
func escapeJSONTail(dst []byte, from int) []byte {
	if !jsonNeedsEscaping(dst[from:]) {
		return dst
	}
	ba := acquireBytesAppender()
	ba.Bytes = append(ba.Bytes, dst[from:]...)
	dst = appendJSONEscaped(dst[:from], unsafeString(ba.Bytes))
	releaseBytesAppender(ba)
	return dst
}

// jsonNeedsEscaping reports whether b holds a byte appendJSONEscaped would
// not copy verbatim: a C0 control, a quote, a backslash, or a byte >= 0x80,
// which has to be validated as UTF-8. It is the scan behind escapeJSONTail,
// where the text is usually a whole message rather than a short field name;
// it shares the word-at-a-time scan of jsonCleanPrefix, so a clean message
// never pays for the escaping copy.
func jsonNeedsEscaping(b []byte) bool {
	return jsonCleanPrefix(unsafeString(b)) < len(b)
}

// jsonWordNeedsEscaping is the boolean form of jsonWordEscapeMask.
func jsonWordNeedsEscaping(w uint64) bool {
	return jsonWordEscapeMask(w) != 0
}

// appendJSONFloat appends f as a JSON-safe value: quoted for non-finite
// values (bare NaN/Inf are not valid JSON), decimal otherwise.
func appendJSONFloat(dst []byte, f float64, bitSize int) []byte {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return append(dst, fallbackFloatString(f)...)
	}
	return strconv.AppendFloat(dst, f, 'f', -1, bitSize)
}

// unsafeString views b as a string without copying. The result must not
// outlive b, and b must not be mutated while it is in use.
func unsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
