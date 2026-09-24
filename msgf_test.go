package iqlog

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

type simpleFormatStringer int

func (simpleFormatStringer) String() string { return "S" }

type simpleFormatNamedInt int

type simpleFormatNamedStr string

// The fast path must produce exactly fmt's bytes for every format it
// accepts, and must decline every format it cannot reproduce so that fmt's
// diagnostics come through untouched.
func TestAppendSimpleFormatMatchesFmt(t *testing.T) {
	cases := []struct {
		format string
		args   []any
		fast   bool // whether the fast path is expected to handle it
	}{
		{"The race was %d minutes and %d seconds long", []any{16, 32}, true},
		{"", nil, true},
		{"plain", nil, true},
		{"%d", []any{7}, true},
		{"%d%s%v%t", []any{1, "a", 2, true}, true},
		{"%%", nil, true},
		{"%%%d%%", []any{5}, true},
		{"a %% b %d c %% d", []any{-1}, true},
		{"%s", []any{""}, true},
		{"%s", []any{"quote \" backslash \\ newline \n \xff"}, true},
		{"%v %v %v %v", []any{"s", 7, true, false}, true},
		{"%d %d %d %d %d %d %d %d %d %d", []any{int8(-8), int16(-16), int32(-32), int64(-64), uint(1), uint8(8), uint16(16), uint32(32), uint64(math.MaxUint64), math.MinInt64}, true},
		{"%v %v %v %v %v %v %v %v %v %v", []any{int8(-8), int16(-16), int32(-32), int64(-64), uint(1), uint8(8), uint16(16), uint32(32), uint64(math.MaxUint64), math.MaxInt64}, true},
		{"%t %t", []any{true, false}, true},
		{"héllo %s 日本", []any{"wörld"}, true},

		// declined: fmt must handle these
		{"%", nil, false},
		{"trailing %", []any{1}, false},
		{"%d", nil, false},
		{"%d %d", []any{1}, false},
		{"%d", []any{1, 2}, false},
		{"plain", []any{1}, false},
		{"", []any{1}, false},
		{"%5d", []any{1}, false},
		{"%-5d", []any{1}, false},
		{"%05d", []any{1}, false},
		{"%+d", []any{1}, false},
		{"% d", []any{1}, false},
		{"%#v", []any{"s"}, false},
		{"%.2f", []any{1.5}, false},
		{"%[1]d", []any{1}, false},
		{"%*d", []any{3, 1}, false},
		{"%x", []any{255}, false},
		{"%q", []any{"s"}, false},
		{"%c", []any{'x'}, false},
		{"%w", []any{errors.New("e")}, false},
		{"%é", []any{1}, false},
		{"%d", []any{"s"}, false},
		{"%d", []any{1.5}, false},
		{"%d", []any{uintptr(1)}, false},
		{"%d", []any{simpleFormatNamedInt(1)}, false},
		{"%d", []any{simpleFormatStringer(1)}, false},
		{"%s", []any{simpleFormatNamedStr("s")}, false},
		{"%s", []any{[]byte("b")}, false},
		{"%s", []any{errors.New("e")}, false},
		{"%s", []any{nil}, false},
		{"%v", []any{nil}, false},
		{"%v", []any{1.5}, false},
		{"%v", []any{float32(1.5)}, false},
		{"%v", []any{[]int{1}}, false},
		{"%v", []any{time.Second}, false},
		{"%v", []any{struct{}{}}, false},
		{"%t", []any{1}, false},
		{"%t", []any{"true"}, false},
		{"%d %s", []any{1, 2}, false},
	}
	for _, c := range cases {
		got, ok := appendSimpleFormat([]byte("prefix|"), c.format, c.args)
		if ok != c.fast {
			t.Errorf("appendSimpleFormat(%q, %v) handled=%v, want %v", c.format, c.args, ok, c.fast)
			continue
		}
		want := "prefix|" + fmt.Sprintf(c.format, c.args...)
		if ok && string(got) != want {
			t.Errorf("appendSimpleFormat(%q, %v) = %q, want %q", c.format, c.args, got, want)
		}
		if full := appendFormatted([]byte("prefix|"), c.format, c.args); string(full) != want {
			t.Errorf("appendFormatted(%q, %v) = %q, want %q", c.format, c.args, full, want)
		}
	}
}

// Random formats built from a mix of accepted and declined tokens, with
// random argument types and counts, must always match fmt whichever path
// they take.
func TestAppendFormattedMatchesFmtRandom(t *testing.T) {
	tokens := []string{"lit ", "%d", "%s", "%v", "%t", "%%", "%5d", "%x", "%", "%.2f", "%q", "%[1]d", " ", "é", "\"", "\n"}
	argPool := []any{0, -1, 42, math.MinInt64, int8(-3), uint8(200), uint64(math.MaxUint64), int32(7), "", "s", "with \"quote\"", "\xff", true, false, nil, 1.5, float32(2.5), []byte("b"), simpleFormatStringer(1), simpleFormatNamedInt(2), simpleFormatNamedStr("n"), errors.New("e"), time.Second, []int{1, 2}}
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		format := ""
		for n := r.Intn(6); n > 0; n-- {
			format += tokens[r.Intn(len(tokens))]
		}
		args := make([]any, r.Intn(5))
		for j := range args {
			args[j] = argPool[r.Intn(len(argPool))]
		}
		want := fmt.Sprintf(format, args...)
		if got := appendFormatted(nil, format, args); string(got) != want {
			t.Fatalf("appendFormatted(%q, %#v) = %q, want %q", format, args, got, want)
		}
	}
}

// The benchmark format and its neighbours take the fast path; a verb it
// cannot reproduce sends the whole format to fmt.
func TestAppendSimpleFormatCoverage(t *testing.T) {
	if _, ok := appendSimpleFormat(nil, "The race was %d minutes and %d seconds long", []any{16, 32}); !ok {
		t.Fatal("benchmark format did not take the fast path")
	}
	if _, ok := appendSimpleFormat(nil, "%s=%v (%t)", []any{"k", 1, true}); !ok {
		t.Fatal("plain verbs did not take the fast path")
	}
	if _, ok := appendSimpleFormat(nil, "%s=%.1f", []any{"k", 1.0}); ok {
		t.Fatal("precision verb took the fast path")
	}
}

// Re-encoding a tail that needs escaping must not allocate, since Msgf
// relies on it for messages that carry quotes or control bytes.
func TestEscapeJSONTailDoesNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	buf := make([]byte, 0, 256)
	if got := testing.AllocsPerRun(1000, func() {
		b := append(buf[:0], `{"message":"`...)
		b = append(b, "a \"quoted\" \n message"...)
		b = escapeJSONTail(b, len(`{"message":"`))
		if string(b) != `{"message":"a \"quoted\" \n message` {
			t.Fatalf("escaped tail = %q", b)
		}
	}); got != 0 {
		t.Fatalf("escapeJSONTail allocated %.2f times", got)
	}
}
