package iqlog

import (
	"encoding/binary"
	"math"
	"strconv"
	"unicode/utf8"
)

// fallbackFloatString formats floats the fast paths cannot handle. Non-finite
// values are quoted so JSON output stays parseable (bare NaN/Inf are not
// valid JSON tokens).
func fallbackFloatString(f float64) string {
	switch {
	case math.IsNaN(f):
		return `"NaN"`
	case math.IsInf(f, 1):
		return `"Infinity"`
	case math.IsInf(f, -1):
		return `"-Infinity"`
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func ansiColourPrefix(level Level) []byte {
	switch level {
	case LevelInfo:
		return []byte("\x1b[34mINFO\x1b[0m ")
	case LevelDebug, LevelTrace:
		return []byte("\x1b[37mDBUG\x1b[0m ")
	case LevelWarn:
		return []byte("\x1b[33mWARN\x1b[0m ")
	case LevelError:
		return []byte("\x1b[31mERRR\x1b[0m ")
	case LevelFatal:
		return []byte("\x1b[31mFATL\x1b[0m ")
	case LevelPanic:
		return []byte("\x1b[31mPANC\x1b[0m ")
	default:
		return []byte("\x1b[34m????\x1b[0m ")
	}
}

// levelPrefix returns a constant []byte for each level (no allocations).
func levelPrefix(level Level) []byte {
	switch level {
	case LevelInfo:
		return []byte("INFO ")
	case LevelDebug, LevelTrace:
		return []byte("DBUG ")
	case LevelWarn:
		return []byte("WARN ")
	case LevelError:
		return []byte("ERRR ")
	case LevelFatal:
		return []byte("FATL ")
	case LevelPanic:
		return []byte("PANC ")
	default:
		return []byte("???? ")
	}
}

const (
	// SWAR lane masks: one byte per lane of a uint64.
	swarLo = 0x0101010101010101
	swarHi = 0x8080808080808080
)

// consoleEscapeLen reports how many bytes at b[i] must not reach a terminal
// verbatim, or 0 when the byte can be copied through.
//
// Two things qualify. The C0 range covers CR and LF (used to forge extra log
// lines) as well as ESC and backspace (used to rewrite what the reader sees);
// DEL goes with them. Tab is left alone: it is common in logged values and
// cannot forge a line or drive the terminal. The C1 range U+0080-U+009F is
// the other half of the same threat -- U+009B is CSI, which a terminal
// honouring C1 treats exactly like "ESC [" -- and in valid UTF-8 it can only
// be written as 0xC2 followed by 0x80-0x9F, so the pair is matched directly.
// inWellFormedUTF8 reports whether b[i] (a byte in 0x80-0x9F) is a
// continuation byte of a well-formed UTF-8 sequence. The candidate lead
// byte at most 3 positions back is decoded with utf8.DecodeRune, which
// enforces length, truncation, overlong, surrogate, and range rules; the
// rune's byte span must then cover i. Truncated/overlong/out-of-range
// sequences return false so the C1 byte falls through to escaping.
func inWellFormedUTF8(b []byte, i int) bool {
	for start := i - 3; start <= i; start++ {
		if start < 0 {
			continue
		}
		if b[start] < 0x80 {
			continue // ASCII cannot start a multi-byte rune
		}
		if b[start] <= 0xbf {
			continue // continuation byte, not a lead
		}
		r, size := utf8.DecodeRune(b[start:])
		if r == utf8.RuneError && size <= 1 {
			continue // invalid, truncated, overlong, surrogate, or out of range
		}
		if size <= 0 || start+size-1 < i || i-start+1 > size {
			continue // rune does not cover i
		}
		return true
	}
	return false
}

// A bare 0x80-0x9F byte is not valid UTF-8, but a terminal honouring C1
// decodes it as C1 anyway, so it is escaped like any other control byte --
// unless inWellFormedUTF8 shows it is a continuation byte of real text, in
// which case escaping it would corrupt ordinary multi-byte output.
func consoleEscapeLen(b []byte, i int) int {
	c := b[i]
	if (c < 0x20 && c != '\t') || c == 0x7f {
		return 1
	}
	if c == 0xc2 && i+1 < len(b) && b[i+1] >= 0x80 && b[i+1] <= 0x9f {
		return 2
	}
	if c >= 0x80 && c <= 0x9f && !inWellFormedUTF8(b, i) {
		return 1
	}
	return 0
}

// indexConsoleEscape returns the index of the first byte in b that starts
// something consoleEscapeLen rejects, or -1 when there is none.
//
// Records are usually clean, so the common case is a full scan that finds
// nothing; doing that a byte at a time is the dominant cost on long lines.
// Eight bytes are tested at once instead, with the classic word tests for "a
// lane below 0x20" and "a lane equal to" 0x7f, 0xC2, or in 0x80-0x9F (the
// latter via "bit 7 set, bits 6 and 5 clear"). The first three can
// over-report -- a borrow between lanes, a tab, or a 0xC2 that does not begin
// a C1 sequence -- so a word that tests positive is rechecked a byte at a
// time. None of them can under-report, which is what makes the fast path safe
// to trust.
func indexConsoleEscape(b []byte) int {
	i := 0
	for ; i+8 <= len(b); i += 8 {
		if !consoleWordSuspect(binary.LittleEndian.Uint64(b[i:])) {
			continue
		}
		for j := i; j < i+8; j++ {
			if consoleEscapeLen(b, j) > 0 {
				return j
			}
		}
	}
	for ; i < len(b); i++ {
		if consoleEscapeLen(b, i) > 0 {
			return i
		}
	}
	return -1
}

// consoleWordSuspect is the word-at-a-time test behind indexConsoleEscape:
// it reports whether any lane of w might hold a byte consoleEscapeLen
// rejects. It may over-report (borrows, tab, a stray 0xC2) but never
// under-reports, and for a word with a single suspect byte it is exact.
func consoleWordSuspect(w uint64) bool {
	del := w ^ (swarLo * 0x7f)
	lead := w ^ (swarLo * 0xc2)
	// 0x80-0x9F: bit 7 set, bits 6 and 5 clear. The shifts move each lane's
	// bits 6 and 5 up into its bit 7 to clear hi there; what they spill into
	// the next lane lands in bits 0-1, where hi is zero.
	hi := w & swarHi
	c1 := hi &^ ((w << 1) | (w << 2))
	return (w-swarLo*0x20)&^w&swarHi != 0 ||
		(del-swarLo)&^del&swarHi != 0 ||
		(lead-swarLo)&^lead&swarHi != 0 ||
		c1 != 0
}

// appendConsoleEscaped appends src to dst with everything consoleEscapeLen
// rejects replaced by a printable escape.
func appendConsoleEscaped(dst, src []byte) []byte {
	for i := 0; i < len(src); {
		n := consoleEscapeLen(src, i)
		if n == 0 {
			dst = append(dst, src[i])
			i++
			continue
		}
		switch c := src[i]; {
		case n == 2:
			// name the C1 code point rather than its UTF-8 bytes
			dst = append(dst, '\\', 'u', '0', '0', hex[src[i+1]>>4], hex[src[i+1]&0x0f])
		case c == '\n':
			dst = append(dst, '\\', 'n')
		case c == '\r':
			dst = append(dst, '\\', 'r')
		default:
			dst = append(dst, '\\', 'x', hex[c>>4], hex[c&0x0f])
		}
		i += n
	}
	return dst
}

// escapeConsoleTail escapes dst[from:], which holds envelope text that came
// from configuration or from an explicit caller rather than from the encoder
// itself. The envelope is exempt from the whole-line scan in
// sanitizeConsoleLine because it carries this package's own ANSI sequences, so
// the parts of it that are not this package's own bytes have to be escaped
// where they are written. Callers keep their ANSI sequences outside from.
func escapeConsoleTail(dst []byte, from int) []byte {
	if indexConsoleEscape(dst[from:]) < 0 {
		return dst
	}
	tail := append([]byte(nil), dst[from:]...)
	return appendConsoleEscaped(dst[:from], tail)
}

// sanitizeConsoleLine escapes control bytes in the caller-supplied part of a
// console-mode line: everything from prefixLen up to the trailing newline.
// Bytes before prefixLen are the record envelope this package generated, so
// they are left alone -- they legitimately contain the ANSI colour sequences
// the console encoder emits. Returns the input unchanged when no escaping is
// needed; otherwise the escaped copy is built in a buffer from the event
// buffer pool, and the caller returns it there (or hands it to the async
// writer, which does) so a stream of such records does not allocate.
func sanitizeConsoleLine(line []byte, prefixLen int) []byte {
	end := len(line)
	if end > 0 && line[end-1] == '\n' {
		end--
	}
	if prefixLen > end {
		prefixLen = end
	}
	first := indexConsoleEscape(line[prefixLen:end])
	if first < 0 {
		return line
	}
	first += prefixLen
	out := append(acquireEventBuffer(), line[:first]...)
	out = appendConsoleEscaped(out, line[first:end])
	return append(out, line[end:]...)
}
