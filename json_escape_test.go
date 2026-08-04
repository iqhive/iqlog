package iqlog

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/iqhive/iqlog/ringbuffer"
)

// syncBuffer is a goroutine-safe buffer for capturing log output in tests.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *syncBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

const injection = "a\"b\\c\nd\te\r{}"

func newJSONTestLogger(buf *syncBuffer) *logger {
	l := NewIQLogger(true)
	l.IncludeTime = false
	l.SetWriter(buf)
	return l
}

func requireValidJSONLines(t *testing.T, output string) {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		var v map[string]any
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			t.Errorf("invalid JSON output %q: %v", line, err)
		}
	}
}

func TestByteSliceLineJSONEscaping(t *testing.T) {
	buf := &syncBuffer{}
	l := newJSONTestLogger(buf)
	l.WithByteSliceLineInfo().Str(injection, injection).Any("any", injection).Msg(injection)
	l.WithByteSliceLineInfo().Int(injection, 1).Int64(injection, 2).Float64(injection, 3.5).Bool(injection, true).Msg("keys")
	l.WithByteSliceLineInfo().Msgs("msg", injection)
	l.WithByteSliceLineInfo().Msgf("formatted %s", injection)
	requireValidJSONLines(t, buf.String())
}

func TestBufferLineJSONEscaping(t *testing.T) {
	buf := &syncBuffer{}
	l := newJSONTestLogger(buf)
	l.WithBufferLineInfo().Str(injection, injection).Any("any", injection).Msg(injection)
	l.WithBufferLineInfo().Int(injection, 1).Int64(injection, 2).Float64(injection, 3.5).Bool(injection, true).Msg("keys")
	l.WithBufferLineInfo().Msgs("msg", injection)
	l.WithBufferLineInfo().Msgf("formatted %s", injection)
	requireValidJSONLines(t, buf.String())
}

func TestBufferLineNLJSONEscaping(t *testing.T) {
	buf := &syncBuffer{}
	l := newJSONTestLogger(buf)
	l.WithBufferLineNLInfo().Str(injection, injection).Any("any", injection).Msg(injection)
	l.WithBufferLineNLInfo().Int(injection, 1).Int64(injection, 2).Float64(injection, 3.5).Bool(injection, true).Msg("keys")
	l.WithBufferLineNLInfo().Msgs("msg", injection)
	l.WithBufferLineNLInfo().Msgf("formatted %s", injection)
	requireValidJSONLines(t, buf.String())
}

func TestPreallocLineJSONEscaping(t *testing.T) {
	buf := &syncBuffer{}
	l := newJSONTestLogger(buf)
	l.WithPreallocLineInfo().Str(injection, injection).Any("any", injection).Msg(injection)
	l.WithPreallocLineInfo().Int(injection, 1).Int64(injection, 2).Float64(injection, 3.5).Bool(injection, true).Msg("keys")
	l.WithPreallocLineInfo().Msgs("msg", injection)
	l.WithPreallocLineInfo().Msgf("formatted %s", injection)
	requireValidJSONLines(t, buf.String())
}

func TestPreallocLine2JSONEscaping(t *testing.T) {
	buf := &syncBuffer{}
	l := newJSONTestLogger(buf)
	l.WithPreallocLine2Info().Str(injection, injection).Any("any", injection).Msg(injection)
	l.WithPreallocLine2Info().Int(injection, 1).Int64(injection, 2).Float64(injection, 3.5).Bool(injection, true).Msg("keys")
	l.WithPreallocLine2Info().Msgs("msg", injection)
	l.WithPreallocLine2Info().Msgf("formatted %s", injection)
	requireValidJSONLines(t, buf.String())
}

func TestVarStackJSONEscaping(t *testing.T) {
	buf := &syncBuffer{}
	l := newJSONTestLogger(buf)
	l.WithVarStackInfo().Str(injection, injection).Any("any", injection).Msg(injection)
	l.WithVarStackInfo().Msgs("msg", injection)
	l.WithVarStackInfo().Msgf("formatted %s", injection)
	requireValidJSONLines(t, buf.String())
}

func TestLargeFloatFormatting(t *testing.T) {
	buf := &syncBuffer{}
	l := newJSONTestLogger(buf)
	l.WithByteSliceLineInfo().Float64("big", math.MaxFloat64).Float64("small", -math.MaxFloat64).Msg("floats")
	l.WithBufferLineInfo().Float64("big", math.MaxFloat64).Msg("floats")
	l.WithPreallocLine2Info().Float64("big", 1e300).Msg("floats")
	l.WithByteSliceLineInfo().Float64("nan", math.NaN()).Float64("inf", math.Inf(1)).Float64("ninf", math.Inf(-1)).Msg("nonfinite")
	l.WithBufferLineInfo().Float64("nan", math.NaN()).Float64("inf", math.Inf(1)).Msg("nonfinite")
	l.WithPreallocLine2Info().Float64("nan", math.NaN()).Float64("inf", math.Inf(1)).Msg("nonfinite")
	requireValidJSONLines(t, buf.String())
}

func TestRingBufferBlockingWakeup(t *testing.T) {
	rb := ringbuffer.NewRingBuffer[int](8)
	done := make(chan int, 100)
	for i := 0; i < 4; i++ {
		go func() {
			for {
				v, ok := rb.DequeueBlocking()
				if !ok {
					return
				}
				done <- v
			}
		}()
	}
	for i := 0; i < 100; i++ {
		for !rb.Enqueue(i) {
		}
	}
	for i := 0; i < 100; i++ {
		<-done
	}
}
