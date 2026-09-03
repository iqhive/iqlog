package iqlog

import (
	"io"
	"runtime"
	"testing"
)

func TestHotPathsDoNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	enabled := MustNew(Config{Format: FormatJSON, Writer: io.Discard})
	customTime := MustNew(Config{Format: FormatJSON, JSONTimeMode: JSONTimeCustom, TimestampLayout: "2006-01-02", Writer: io.Discard})
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
	if got := testing.AllocsPerRun(1000, func() {
		customTime.InfoEvent().Str("rate", "15").Msg("message")
	}); got != 0 {
		t.Fatalf("custom timestamp event allocated %.2f times", got)
	}
}

// The buffer pool hands slices around by value, so the *[]byte handle
// sync.Pool.Put needs must be recycled rather than freshly allocated.
func TestEventBufferPoolRoundTripDoesNotAllocate(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	n := testing.AllocsPerRun(10000, func() {
		b := acquireEventBuffer()
		b = append(b, "some record bytes"...)
		releaseEventBuffer(b)
	})
	if n > 0 {
		t.Errorf("buffer pool round trip allocates %.2f times", n)
	}
}

// AllocsPerRun with a deep queue measures the warm-up, not steady state: the
// caller can run thousands of records ahead of the drain, so every one needs a
// fresh buffer. Pace the caller against the drain instead.
func TestAsyncSteadyStateAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	for _, size := range []int{1, 8, 64} {
		l := MustNew(Config{Format: FormatJSON, Writer: io.Discard, Level: LevelInfo,
			WriterMode: WriterAsync, BufferSize: size})
		// warm the pools
		for i := 0; i < 50000; i++ {
			l.InfoEvent().Str("rate", "15").Int("low", 16).Msg("message")
		}
		_ = l.Flush()
		runtime.GC()

		var m0, m1 runtime.MemStats
		runtime.ReadMemStats(&m0)
		const n = 200000
		for i := 0; i < n; i++ {
			l.InfoEvent().Str("rate", "15").Int("low", 16).Msg("message")
		}
		_ = l.Flush()
		runtime.ReadMemStats(&m1)
		perOp := float64(m1.Mallocs-m0.Mallocs) / float64(n)
		bytesPerOp := float64(m1.TotalAlloc-m0.TotalAlloc) / float64(n)
		t.Logf("BufferSize=%-3d  %.3f allocs/op  %.1f B/op", size, perOp, bytesPerOp)
		if perOp > 0.05 {
			t.Errorf("BufferSize=%d: async steady state allocates %.3f/op", size, perOp)
		}
		_ = l.Close()
	}
}

func TestSyncSteadyStateAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector allocates for its own bookkeeping")
	}
	l := MustNew(Config{Format: FormatJSON, Writer: io.Discard, Level: LevelInfo})
	for i := 0; i < 10000; i++ {
		l.InfoEvent().Str("rate", "15").Msg("m")
	}
	runtime.GC()
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	const n = 200000
	for i := 0; i < n; i++ {
		l.InfoEvent().Str("rate", "15").Int("low", 16).Msg("message")
	}
	runtime.ReadMemStats(&m1)
	t.Logf("sync %.3f allocs/op", float64(m1.Mallocs-m0.Mallocs)/float64(n))
}
