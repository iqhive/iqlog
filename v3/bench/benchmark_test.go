// go test -v -cpu=4 -run=none -bench=. -benchtime=10s -benchmem bench_test.go
package benchmarks

import (
	"io"
	"log"
	"log/slog"
	"testing"

	"bitbucket.org/iqhive/iqlog/v3"
)

const msg = "The quick brown fox jumps over the lazy dog"

var obj = struct {
	Rate string
	Low  int
	High float32
}{"15", 16, 123.2}

func BenchmarkSlogDisabled(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	for i := 0; i < b.N; i++ {
		logger.Debug(msg, "rate", "15", "low", 16, "high", 123.2)
	}
}

func BenchmarkSlogSimple(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	for i := 0; i < b.N; i++ {
		logger.Info(msg, "rate", "15", "low", 16, "high", 123.2)
	}
}

func BenchmarkSlogPrintf(b *testing.B) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	for i := 0; i < b.N; i++ {
		log.Printf("rate=%s low=%d high=%f msg=%s", "15", 16, 123.2, msg)
	}
}

func BenchmarkSlogAny(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	for i := 0; i < b.N; i++ {
		logger.Info(msg, "rate", "15", "low", 16, "object", &obj)
	}
}

func BenchmarkIQLogDisabled(b *testing.B) {
	// logger := phuslog.Logger{Level: phuslog.InfoLevel, Writer: phuslog.IOWriter{io.Discard}}
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Debug(msg)
		// logger.Debug(msg)
	}
}

func BenchmarkIQLogSimple(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Info(msg)
		// logger.Info(msg)
	}
}

func BenchmarkIQLogPrintf(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.Infof("rate=%s low=%d high=%f msg=%s", "15", 16, 123.2, msg)
		// logger.Info(msg)
	}
}

func BenchmarkIQLogAny(b *testing.B) {
	logger := iqlog.NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.Any("rate", "15").Any("low", 16).Any("object", &obj).Info(msg)
		// logger.Info(msg)
	}
}
