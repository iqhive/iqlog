// go test -v -cpu=4 -run=none -bench=. -benchtime=10s -benchmem bench_test.go
package benchmarks

import (
	"io"
	"testing"

	"github.com/iqhive/iqlog/v3"
)

func BenchmarkPreallocLineF(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.WithPreallocLineInfo().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("testing printf %d = %s", i, msg)
		// logger.WithPreallocLineInfo()
	}
}

func BenchmarkPreallocLine2F(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.WithPreallocLine2Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("testing printf %d = %s", i, msg)
		// logger.WithPreallocLineInfo()
	}
}

func BenchmarkBytesliceLineF(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		// logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Info(msg)
		logger.WithByteSliceLineInfo().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("testing printf %d = %s", i, msg)
		// logger.Info(msg)
	}
}

func BenchmarkBufferLineF(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		// logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Info(msg)
		logger.WithBufferLineInfo().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("testing printf %d = %s", i, msg)
		// logger.Info(msg)
	}
}

func BenchmarkBufferLineNLF(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		// logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Info(msg)
		logger.WithBufferLineNLInfo().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("testing printf %d = %s", i, msg)
		// logger.Info(msg)
	}
}

func BenchmarkVarsTimestampMsgf(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.InfoWith().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("testing printf %d = %s", i, msg)
		// logger.WithPreallocLineInfo()
	}
}

func BenchmarkVarsTimestampMsgInfof(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.Infof("testing printf %d = %s", i, msg)
		// logger.WithPreallocLineInfo()
	}
}

func BenchmarkGlobalVarsTimestampMsgf(b *testing.B) {
	iqlog.SetDebugMode(false)
	iqlog.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		iqlog.InfoWith().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf("testing printf %d = %s", i, msg)
		// logger.WithPreallocLineInfo()
	}
}

func BenchmarkGlobalVarsTimestampMsgInfof(b *testing.B) {
	iqlog.SetDebugMode(false)
	iqlog.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		iqlog.Infof("testing printf %d = %s", i, msg)
		// logger.WithPreallocLineInfo()
	}
}
