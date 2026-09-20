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

// R-02: a lone C1 byte (0x80-0x9F outside well-formed UTF-8) must not reach
// the terminal, while ordinary multi-byte text passes through and the valid
// 2-byte UTF-8 C1 pair is neutralised as a named escape.
func TestConsoleSanitizesLoneC1(t *testing.T) {
	for _, c := range []byte{0x80, 0x9b, 0x9d, 0x9f} {
		line := append([]byte("k=a"), c)
		line = append(line, "b\n"...)
		got := string(sanitizeConsoleLine(line, 0))
		if bytes.Contains([]byte(got), []byte{c}) {
			t.Errorf("lone C1 %#x reached output: %q", c, got)
		}
		want := `\x` + string(hex[c>>4]) + string(hex[c&0x0f])
		if !strings.Contains(got, want) {
			t.Errorf("lone C1 %#x: want escape %q in %q", c, want, got)
		}
	}
	// valid UTF-8 C1 pair is neutralised as a named escape, not raw bytes
	pair := []byte{0xc2, 0x9b}
	pline := append(append([]byte("k="), pair...), '\n')
	if got := string(sanitizeConsoleLine(pline, 0)); strings.Contains(got, string(pair)) || !strings.Contains(got, `\u009b`) {
		t.Errorf("valid C1 pair not neutralised: %q", got)
	}
	// tab stays unescaped
	tab := []byte("k=a\tb\n")
	if got := sanitizeConsoleLine(tab, 0); !bytes.Equal(got, tab) {
		t.Errorf("tab mangled: got %q", got)
	}
}

// R-02 regression: a 0x80-0x9F byte must be escaped unless it is a
// continuation byte of a genuinely well-formed UTF-8 rune. Truncated,
// overlong, surrogate, or out-of-range lead bytes must NOT shield it.
func TestConsoleEscapeC1Sequences(t *testing.T) {
	mustEscape := []struct {
		name  string
		input []byte
		idx   int // index of the C1 byte under test
	}{
		{"lone 9B", []byte{0x9b}, 0},
		{"lone 9D", []byte{0x9d}, 0},
		{"lone 80", []byte{0x80}, 0},
		{"lone 9F", []byte{0x9f}, 0},
		{"truncated E0 9B", []byte{0xe0, 0x9b}, 1},
		{"overlong E0 80 9B", []byte{0xe0, 0x80, 0x9b}, 2},
		{"surrogate ED 9B", []byte{0xed, 0x9b}, 1},
		{"overlong F0 80 9B", []byte{0xf0, 0x80, 0x9b}, 2},
		{"truncated F0 90 80 at end", []byte{0xf0, 0x90, 0x80}, 2},
		{"lone C1 after text", []byte{'a', 0x9b}, 1},
	}
	for _, tc := range mustEscape {
		t.Run(tc.name, func(t *testing.T) {
			if got := consoleEscapeLen(tc.input, tc.idx); got == 0 {
				t.Errorf("consoleEscapeLen(% X, %d) = 0, want > 0", tc.input, tc.idx)
			}
			line := append(append([]byte("k="), tc.input...), '\n')
			got := string(sanitizeConsoleLine(line, 0))
			if strings.Contains(got, string([]byte{tc.input[tc.idx]})) {
				t.Errorf("C1 byte %#x reached output: %q", tc.input[tc.idx], got)
			}
			if idx := indexConsoleEscape(tc.input); idx < 0 {
				t.Errorf("indexConsoleEscape(% X) = -1, want %d", tc.input, tc.idx)
			}
		})
	}

	mustPreserve := []struct {
		name  string
		input []byte
	}{
		{"valid C2 9B pair shape", []byte{0xc2, 0x9b}},
		{"valid 3-byte E0 A0 9B", []byte{0xe0, 0xa0, 0x9b}},
		{"valid 4-byte F0 90 80 9F", []byte{0xf0, 0x90, 0x80, 0x9f}},
		{"ascii", []byte("hello world")},
		{"tab", []byte("a\tb")},
	}
	for _, tc := range mustPreserve {
		t.Run(tc.name, func(t *testing.T) {
			if !utf8.Valid(tc.input) {
				t.Fatalf("premise: % X should be valid UTF-8", tc.input)
			}
			// every 0x80-0x9F byte in a valid rune must not hit the lone-C1 rule
			for i, c := range tc.input {
				if c >= 0x80 && c <= 0x9f && !inWellFormedUTF8(tc.input, i) {
					t.Errorf("valid continuation % X at %d flagged for escape", tc.input, i)
				}
			}
			// C2-pair input is neutralised as a unit; genuine text passes through
			if len(tc.input) == 2 && tc.input[0] == 0xc2 {
				return
			}
			line := append(append([]byte(nil), tc.input...), '\n')
			if got := sanitizeConsoleLine(line, 0); !bytes.Equal(got, line) {
				t.Errorf("text mangled: got %q want %q", got, line)
			}
		})
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
