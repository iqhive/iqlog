// go test -v -cpu=4 -run=none -bench=. -benchtime=10s -benchmem bench_test.go
package benchmarks

import (
	"context"
	"io"
	"log"
	"log/slog"
	"testing"

	"github.com/iqhive/iqlog"
	phuslog "github.com/phuslu/log"
	"github.com/rs/zerolog"
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
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Level: iqlog.LevelInfo, Writer: io.Discard})
	for i := 0; i < b.N; i++ {
		// logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Debug(msg)
		logger.DebugEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		// logger.Debug(msg)
	}
}

func BenchmarkIQLogDisabledGuarded(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Level: iqlog.LevelInfo, Writer: io.Discard})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if event := logger.DebugEvent(); event != nil {
			event.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		}
	}
}

func BenchmarkIQLogSimple(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard})
	for i := 0; i < b.N; i++ {
		// logger.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Info(msg)
		logger.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		// logger.Info(msg)
	}
}

func BenchmarkIQLogPrintf(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard})
	for i := 0; i < b.N; i++ {
		// logger.Infof("rate=%s low=%d high=%f msg=%s", "15", 16, 123.2, msg)
		logger.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf(msg)
		// logger.Info(msg)
	}
}

func BenchmarkIQLogAny(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard})
	for i := 0; i < b.N; i++ {
		// logger.Any("rate", "15").Any("low", 16).Any("object", &obj).Info(msg)
		logger.InfoEvent().Any("rate", "15").Any("low", 16).Any("object", &obj).Msg(msg)
		// logger.Info(msg)
	}
}

func BenchmarkZerologSimple(b *testing.B) {
	logger := zerolog.New(io.Discard)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkPhusluSimple(b *testing.B) {
	logger := phuslog.Logger{Level: phuslog.InfoLevel, Writer: phuslog.IOWriter{Writer: io.Discard}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkZerologDisabled(b *testing.B) {
	logger := zerolog.New(io.Discard).Level(zerolog.InfoLevel)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Debug().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkPhusluDisabled(b *testing.B) {
	logger := phuslog.Logger{Level: phuslog.InfoLevel, Writer: phuslog.IOWriter{Writer: io.Discard}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if event := logger.Debug(); event != nil {
			event.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		}
	}
}

func BenchmarkIQLogParallel(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard})
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		}
	})
}

func BenchmarkZerologParallel(b *testing.B) {
	logger := zerolog.New(io.Discard)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		}
	})
}

func BenchmarkPhusluParallel(b *testing.B) {
	logger := phuslog.Logger{Level: phuslog.InfoLevel, Writer: phuslog.IOWriter{Writer: io.Discard}}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		}
	})
}

func BenchmarkIQLogInfo(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Writer: io.Discard})
	for i := 0; i < b.N; i++ {
		logger.Info(msg)
	}
}

func BenchmarkIQLogPackageInfo(b *testing.B) {
	iqlog.SetDefault(iqlog.MustNew(iqlog.Config{Writer: io.Discard}))
	for i := 0; i < b.N; i++ {
		iqlog.Info(msg)
	}
}

func BenchmarkIQLogLegacyConstructorInfo(b *testing.B) {
	logger := iqlog.NewIQLogger(false)
	_ = logger.SetWriter(io.Discard)
	for i := 0; i < b.N; i++ {
		logger.Info(msg)
	}
}

func BenchmarkIQLogLegacyInfoWith(b *testing.B) {
	iqlog.SetDefault(iqlog.MustNew(iqlog.Config{Writer: io.Discard}))
	for i := 0; i < b.N; i++ {
		iqlog.InfoWith().Str("rate", "15").Int("low", 16).Msg(msg)
	}
}

func BenchmarkIQLogLegacyWarning(b *testing.B) {
	iqlog.SetDefault(iqlog.MustNew(iqlog.Config{Writer: io.Discard}))
	for i := 0; i < b.N; i++ {
		iqlog.Warning(msg)
	}
}

func BenchmarkIQLogLegacyLogWithFields(b *testing.B) {
	logger := iqlog.NewIQLogger(false)
	_ = logger.SetWriter(io.Discard)
	fields := map[string]any{"rate": "15", "low": 16}
	for i := 0; i < b.N; i++ {
		logger.LogWithFields(context.Background(), iqlog.LevelInfo, fields, msg)
	}
}
