package iqlog

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestEventLongAndHostileValues(t *testing.T) {
	for _, format := range []Format{FormatConsole, FormatJSON} {
		sb := &syncBuffer{}
		l := MustNew(Config{Format: format, Writer: sb, EscapeFieldNames: true})
		long := strings.Repeat(`a"b\c`, 1000)
		l.InfoEvent().Str("hostile\nkey", long).Msgf("prefix %s END", long)
		out := sb.String()
		if !strings.Contains(out, "END") {
			t.Fatal("long output truncated")
		}
		if strings.Count(out, "\n") != 1 {
			t.Fatalf("record framing broken: %q", out)
		}
		if format == FormatJSON && !json.Valid([]byte(strings.TrimSpace(out))) {
			t.Fatalf("invalid JSON: %q", out)
		}
	}
}

func TestEventNonFiniteJSONValid(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb})
	l.InfoEvent().Float64("nan", math.NaN()).Float64("inf", math.Inf(1)).Msgs("values", math.Inf(-1))
	if !json.Valid([]byte(strings.TrimSpace(sb.String()))) {
		t.Fatalf("invalid JSON: %q", sb.String())
	}
}

func TestPanicAndFilteredPanic(t *testing.T) {
	for _, filtered := range []bool{false, true} {
		sb := &syncBuffer{}
		level := LevelInfo
		if filtered {
			level = LevelFatal
		}
		l := MustNew(Config{Writer: sb, Level: level})
		func() {
			defer func() {
				if got := recover(); got != "boom 42" {
					t.Fatalf("panic=%v", got)
				}
			}()
			l.PanicEvent().Msgf("boom %d", 42)
		}()
		if filtered && sb.String() != "" {
			t.Fatalf("filtered panic wrote %q", sb.String())
		}
		if !filtered && !strings.Contains(sb.String(), "boom 42") {
			t.Fatal("panic record missing")
		}
	}
}
