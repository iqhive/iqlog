package iqlog

// Correctness coverage for the two output sanitizers: the console control-byte
// scan (which uses word-at-a-time tests that must never under-report) and the
// JSON string escaper (which must agree with encoding/json, including its
// U+FFFD substitution for ill-formed UTF-8).

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"
)

func naiveIndexConsoleEscape(b []byte) int {
	for i := range b {
		if consoleEscapeLen(b, i) > 0 {
			return i
		}
	}
	return -1
}

// every single byte value, at every offset within and across a word
func TestIndexConsoleEscapeAllBytesAllOffsets(t *testing.T) {
	for c := 0; c < 256; c++ {
		for n := 0; n <= 40; n++ {
			for pos := 0; pos < n; pos++ {
				b := make([]byte, n)
				for i := range b {
					b[i] = 'a'
				}
				b[pos] = byte(c)
				if got, want := indexConsoleEscape(b), naiveIndexConsoleEscape(b); got != want {
					t.Fatalf("byte %#x at %d in len %d: got %d want %d", c, pos, n, got, want)
				}
			}
		}
	}
}

// adjacent low bytes are where inter-lane borrows would bite
func TestIndexConsoleEscapeBorrowPatterns(t *testing.T) {
	for a := 0; a < 256; a++ {
		for c := 0; c < 256; c++ {
			for pos := 0; pos < 16; pos++ {
				b := make([]byte, 24)
				for i := range b {
					b[i] = 'x'
				}
				b[pos] = byte(a)
				b[pos+1] = byte(c)
				if got, want := indexConsoleEscape(b), naiveIndexConsoleEscape(b); got != want {
					t.Fatalf("pair %#x,%#x at %d: got %d want %d", a, c, pos, got, want)
				}
			}
		}
	}
}

func TestIndexConsoleEscapeRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for iter := 0; iter < 300000; iter++ {
		b := make([]byte, r.Intn(64))
		for i := range b {
			switch r.Intn(4) {
			case 0:
				b[i] = byte(r.Intn(0x21)) // heavy on the control range
			case 1:
				b[i] = []byte{0x7e, 0x7f, 0x80, 0xc1, 0xc2, 0xc3, 0x9b, 0x9f, 0xa0}[r.Intn(9)]
			default:
				b[i] = byte(r.Intn(256))
			}
		}
		if got, want := indexConsoleEscape(b), naiveIndexConsoleEscape(b); got != want {
			t.Fatalf("%q: got %d want %d", b, got, want)
		}
	}
}

// the escaper must produce the same string encoding/json would, modulo the
// extra escapes encoding/json applies for HTML/JS safety
func TestAppendJSONEscapedMatchesStdlib(t *testing.T) {
	check := func(in string) {
		t.Helper()
		got := string(appendJSONEscaped(nil, in))
		if !utf8.ValidString(got) {
			t.Fatalf("input %q produced invalid UTF-8: %q", in, got)
		}
		var round string
		if err := json.Unmarshal([]byte(`"`+got+`"`), &round); err != nil {
			t.Fatalf("input %q produced unparseable JSON string %q: %v", in, got, err)
		}
		// encoding/json applies the same U+FFFD substitution
		var viaStdlib string
		b, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, &viaStdlib); err != nil {
			t.Fatal(err)
		}
		if round != viaStdlib {
			t.Errorf("input %q: got %q, encoding/json gives %q", in, round, viaStdlib)
		}
	}

	for _, s := range []string{
		"", "plain", "a\"b", `a\b`, "a\nb", "a\rb", "a\tb", "a\x00b", "a\x1fb",
		"héllo", "日本語", "emoji 🎉 here", "\xff", "\xff\xfe", "a\xffb",
		"\xe4\xb8", "\xf0\x9f", "\xc3", "\xc3\x28",
		strings.Repeat("a\xc3\xa9", 500), // deep ASCII/non-ASCII alternation
		strings.Repeat("\xff", 1000),
		strings.Repeat("x\xffy", 2000),
	} {
		check(s)
	}

	r := rand.New(rand.NewSource(7))
	for i := 0; i < 20000; i++ {
		b := make([]byte, r.Intn(40))
		for j := range b {
			b[j] = byte(r.Intn(256))
		}
		check(string(b))
	}
}

// alternating input must not grow the stack without bound
func TestAppendJSONEscapedDeepAlternationNoStackGrowth(t *testing.T) {
	in := strings.Repeat("a\xc3\xa9", 300000)
	got := appendJSONEscaped(nil, in)
	if !utf8.Valid(got) {
		t.Fatal("invalid UTF-8")
	}
	if len(got) != len(in) {
		t.Fatalf("len %d, want %d", len(got), len(in))
	}
}

// U+009B is CSI in the C1 control set: on a terminal that honours C1 it is
// equivalent to "ESC [". In valid UTF-8 the C1 range U+0080-U+009F is encoded
// as 0xC2 followed by 0x80-0x9F, and nothing else can produce it.
func TestC1ControlsReachTerminal(t *testing.T) {
	csi := ""
	osc := ""
	payload := "user" + csi + "2K" + csi + "1;31mROOT" + osc + "0;title"
	if !utf8.ValidString(payload) {
		t.Fatal("premise: payload should be valid UTF-8")
	}
	var buf bytes.Buffer
	l := MustNew(Config{Format: FormatConsole, Writer: &buf, Level: LevelInfo})
	l.InfoEvent().Str("u", payload).Msg("login")
	out := buf.String()
	t.Logf("console: %q", out)
	body := out[strings.Index(out, "u="):]
	for _, r := range body {
		if r >= 0x80 && r <= 0x9f {
			t.Errorf("BUG: C1 control U+%04X reached the terminal in %q", r, body)
			break
		}
	}
}

// normal multi-byte text must not be mangled by whatever fix is applied
func TestC1FixDoesNotMangleText(t *testing.T) {
	for _, s := range []string{
		"日本語のログ", "héllo wörld", "emoji 🎉", "Ω≈ç√∫", "naïve café",
		" nbsp", "¿¡", "Ā ā Ē ē", "→←↑↓", "Ǆǅǆ",
	} {
		var buf bytes.Buffer
		l := MustNew(Config{Format: FormatConsole, Writer: &buf, Level: LevelInfo})
		l.InfoEvent().Str("k", s).Msg("m")
		out := buf.String()
		if !strings.Contains(out, s) {
			t.Errorf("text mangled: %q not present in %q", s, out)
		}
	}
}
