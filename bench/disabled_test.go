// go test -v -cpu=4 -run=none -bench=. -benchtime=10s -benchmem bench_test.go
package benchmarks

import (
	"io"
	"testing"

	"github.com/iqhive/iqlog/v3"
)

func BenchmarkDisabledPreallocLine(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.WithPreallocLineDebug().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		// logger.WithPreallocLineDebug()
	}
}

func BenchmarkDisabledPreallocLine2(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.WithPreallocLine2Debug().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		// logger.WithPreallocLineDebug()
	}
}

func BenchmarkDisabledBytesliceLine(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		// logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Info(msg)
		logger.WithPreallocLineDebug().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		// logger.Info(msg)
	}
}

func BenchmarkDisabledBufferLine(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		// logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Info(msg)
		logger.WithBufferLineDebug().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		// logger.Info(msg)
	}
}

func BenchmarkDisabledBufferLineNL(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		// logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Info(msg)
		logger.WithBufferLineNLDebug().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		// logger.Info(msg)
	}
}
