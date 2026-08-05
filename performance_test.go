package iqlog

import (
	"io"
	"testing"
)

func TestHotPathsDoNotAllocate(t *testing.T) {
	enabled := MustNew(Config{Format: FormatJSON, Writer: io.Discard})
	disabled := MustNew(Config{Format: FormatJSON, Level: LevelInfo, Writer: io.Discard})

	if got := testing.AllocsPerRun(1000, func() {
		enabled.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg("message")
	}); got != 0 {
		t.Fatalf("enabled typed event allocated %.2f times", got)
	}
	if got := testing.AllocsPerRun(1000, func() {
		disabled.DebugEvent().Str("rate", "15").Int("low", 16).Msg("message")
	}); got != 0 {
		t.Fatalf("disabled event allocated %.2f times", got)
	}
}
