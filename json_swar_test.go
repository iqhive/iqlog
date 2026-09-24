package iqlog

import (
	"encoding/binary"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// The word test must agree with the byte table for every byte value in
// every lane, alone and next to other bytes: the fast path copies whatever
// it declares clean without looking again.
func TestJSONWordNeedsEscapeMatchesByteTable(t *testing.T) {
	for lane := 0; lane < 8; lane++ {
		for c := 0; c < 256; c++ {
			var word [8]byte
			for i := range word {
				word[i] = 'a'
			}
			word[lane] = byte(c)
			got := jsonWordEscapeMask(binary.LittleEndian.Uint64(word[:])) != 0
			if got != jsonEscapeTable[c] {
				t.Fatalf("byte %#x in lane %d: word test %v, table %v", c, lane, got, jsonEscapeTable[c])
			}
		}
	}
	// Neighbouring values that sit just either side of every boundary the
	// arithmetic relies on, in every pair of lanes.
	edges := []byte{0, 0x1f, 0x20, 0x21, '"' - 1, '"', '"' + 1, '\\' - 1, '\\', '\\' + 1, 0x7f, 0x80, 0xff}
	for _, a := range edges {
		for _, b := range edges {
			for lane := 0; lane < 7; lane++ {
				var word [8]byte
				for i := range word {
					word[i] = 'x'
				}
				word[lane], word[lane+1] = a, b
				want := jsonEscapeTable[a] || jsonEscapeTable[b]
				if got := jsonWordEscapeMask(binary.LittleEndian.Uint64(word[:])) != 0; got != want {
					t.Fatalf("bytes %#x %#x at lane %d: word test %v, table %v", a, b, lane, got, want)
				}
			}
		}
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 200000; i++ {
		var word [8]byte
		rng.Read(word[:])
		want := false
		for _, c := range word {
			want = want || jsonEscapeTable[c]
		}
		if got := jsonWordEscapeMask(binary.LittleEndian.Uint64(word[:])) != 0; got != want {
			t.Fatalf("word %x: word test %v, table %v", word, got, want)
		}
	}
}

// appendJSONEscapedBytewise is the reference: the byte loop alone, with no
// word fast path in front of it.
func appendJSONEscapedBytewise(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !jsonEscapeTable[c] {
			dst = append(dst, c)
			continue
		}
		if c >= 0x80 {
			var n int
			dst, n = appendJSONEscapedRunes(dst, s[i:])
			i += n - 1
			continue
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
	}
	return dst
}

func TestJSONCleanPrefixAndEscapeAgreeWithBytewise(t *testing.T) {
	cases := []string{
		"", "a", "1234567", "12345678", "123456789", "The quick brown fox jumps over the lazy dog",
		"\"", "\\", "\n", "\x00", "\x7f", "\xff", "ünï", "日本語",
		"1234567\"", "12345678\"", "123456789\"", "\"12345678", "1234\"5678", "12345678\"12345678",
		"12345678ü", "1234567ü", "ü12345678", "12345678\xff", "1234567\xffabcdefgh",
		"12345678\x00", "12345678\t", "12345678\\", "abcdefgh\x80", "abcdefgh\xc3", "abcdefgh\xc3\xa9",
		strings.Repeat("x", 63) + "\"", strings.Repeat("x", 64) + "\"", strings.Repeat("x", 65) + "\"",
		strings.Repeat("\"", 17), strings.Repeat("é", 9), "a\"b\\c\nd\re\tf\x01g\x1fh\x7fi\x80j\xffk",
		// word-pair, word and overlapping-tail boundaries
		"123456789012345", "1234567890123456", "12345678901234567", "123456789012345678901234", "1234567890123456789012345",
		"123456789012345\"", "1234567890123456\"", "12345678901234567\"", "123456789012345678901234\"", "1234567890123456789012345\"",
		"\"234567890123456", "1234567\"90123456", "12345678\"0123456", "123456789012345\"7", "123456789012345678901234\"6",
		"12345678901234567890123\"", "1234567890123456789012\"4", strings.Repeat("y", 200), strings.Repeat("y", 199) + "\n",
		"\n" + strings.Repeat("y", 199), strings.Repeat("y", 100) + "\x80" + strings.Repeat("y", 99),
	}
	rng := rand.New(rand.NewSource(2))
	alphabet := []byte("abcdefghijklmnopqrstuvwxyz0123456789 \"\\\n\r\t\x00\x01\x1f\x7f\x80\xc3\xa9\xe6\x97\xa5\xff")
	for i := 0; i < 5000; i++ {
		n := rng.Intn(40)
		b := make([]byte, n)
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		cases = append(cases, string(b))
	}
	for _, s := range cases {
		clean := jsonCleanPrefix(s)
		if clean < 0 || clean > len(s) {
			t.Fatalf("%q: clean prefix %d out of range", s, clean)
		}
		for i := 0; i < clean; i++ {
			if jsonEscapeTable[s[i]] {
				t.Fatalf("%q: byte %#x at %d inside clean prefix %d", s, s[i], i, clean)
			}
		}
		if clean == len(s) {
			for i := 0; i < len(s); i++ {
				if jsonEscapeTable[s[i]] {
					t.Fatalf("%q: declared clean but byte %#x at %d needs escaping", s, s[i], i)
				}
			}
		}
		want := string(appendJSONEscapedBytewise(nil, s))
		if got := string(appendJSONEscaped([]byte("pre:"), s)); got != "pre:"+want {
			t.Fatalf("%q: fast path %q, bytewise %q", s, got, want)
		}
		if got, want := jsonKeyNeedsEscaping(s), clean != len(s); got != want {
			t.Fatalf("%q: key scan %v, clean prefix %d of %d", s, got, clean, len(s))
		}
	}
}

// The key scan's packed-halves path covers lengths four to seven; every
// position in every such length must see a dirty byte of every kind.
func TestJSONKeyNeedsEscapingEveryPosition(t *testing.T) {
	for n := 0; n <= 12; n++ {
		base := strings.Repeat("k", n)
		if jsonKeyNeedsEscaping(base) {
			t.Fatalf("clean key of length %d reported dirty", n)
		}
		for pos := 0; pos < n; pos++ {
			for _, c := range []byte{0, 0x1f, '"', '\\', 0x80, 0xff} {
				key := base[:pos] + string([]byte{c}) + base[pos+1:]
				if !jsonKeyNeedsEscaping(key) {
					t.Fatalf("key %q: byte %#x at %d not detected", key, c, pos)
				}
			}
			// DEL is not escaped in JSON, so it is clean like the others.
			for _, c := range []byte{0x20, 0x21, '!', '~', 'z', '0', 0x7f} {
				key := base[:pos] + string([]byte{c}) + base[pos+1:]
				if jsonKeyNeedsEscaping(key) {
					t.Fatalf("key %q: clean byte %#x at %d reported dirty", key, c, pos)
				}
			}
		}
	}
}

func TestIsInfraFramePrefilterKeepsExactMatch(t *testing.T) {
	cases := map[string]bool{
		"runtime.goexit":                        true,
		"runtime.(*g).x":                        true,
		"log.Printf":                            true,
		"log.(*Logger).Output":                  true,
		"log/slog.(*Logger).log":                true,
		"log/slog.(*Logger).log.func1":          true,
		"log/slog.Info":                         true,
		"runtime/debug.Stack":                   false,
		"log.example.com/x.Func":                false,
		"log.example.com/x.(*T).M":              false,
		"logx.Func":                             false,
		"mylog.Func":                            false,
		"example.com/log.Func":                  false,
		"example.com/runtime.Func":              false,
		"github.com/iqhive/iqlog.(*Logger).Msg": false,
		"main.main":                             false,
		"":                                      false,
		"runtime":                               false,
		"log":                                   false,
	}
	for name, want := range cases {
		if got := isInfraFrame(name); got != want {
			t.Fatalf("isInfraFrame(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestCallerDataSetLineFormatting(t *testing.T) {
	for _, line := range []int{0, 1, 9, 10, 99, 100, 346, 4096, 65535, 1 << 20, 1<<31 - 1, -1, -12} {
		var data callerData
		data.set("pkg.Func", "/a/b/file.go", line)
		want := "file.go:" + strconv.Itoa(line)
		if got := string(data.callerFile[:data.callerFileLen]); got != want {
			t.Fatalf("line %d: %q want %q", line, got, want)
		}
	}
	// A long basename is cut to leave room for the complete line suffix.
	var data callerData
	data.set("pkg.Func", "/x/"+strings.Repeat("f", 200)+".go", 123456)
	got := string(data.callerFile[:data.callerFileLen])
	if len(got) != callerDataMaxLen || !strings.HasSuffix(got, ":123456") {
		t.Fatalf("truncated file %q (len %d)", got, len(got))
	}
}
