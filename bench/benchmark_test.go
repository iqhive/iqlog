// go test -v -cpu=4 -run=none -bench=. -benchtime=10s -benchmem bench_test.go
package benchmarks

import (
	"context"
	"io"
	"log"
	"log/slog"
	"testing"
	"time"

	"github.com/iqhive/iqlog"
	phuslog "github.com/phuslu/log"
	"github.com/rs/zerolog"
)

const msg = "The quick brown fox jumps over the lazy dog"
const msgf = "The race was %d minutes and %d seconds long"

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

func BenchmarkIQLogDisabled(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Level: iqlog.LevelInfo, Writer: io.Discard})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if event := logger.DebugEvent(); event != nil {
			event.Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
		}
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

func BenchmarkSlogSimple(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	for i := 0; i < b.N; i++ {
		logger.Info(msg, "rate", "15", "low", 16, "high", 123.2)
	}
}

func BenchmarkIQLogSimple(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard, JSONTimeMode: iqlog.JSONTimeDisabled})
	for i := 0; i < b.N; i++ {
		logger.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
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
	logger := phuslog.Logger{Level: phuslog.InfoLevel, Writer: phuslog.IOWriter{Writer: io.Discard}, TimeField: ""}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkSlogPrintf(b *testing.B) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	for i := 0; i < b.N; i++ {
		log.Printf("rate=%s low=%d high=%f msg=%s", "15", 16, 123.2, msg)
	}
}

func BenchmarkIQLogPrintf(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard, JSONTimeMode: iqlog.JSONTimeDisabled})
	for i := 0; i < b.N; i++ {
		logger.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf(msgf, 16, 32)
	}
}

func BenchmarkZerologPrintf(b *testing.B) {
	logger := zerolog.New(io.Discard)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf(msgf, 16, 32)
	}
}

func BenchmarkPhusluPrintf(b *testing.B) {
	logger := phuslog.Logger{Level: phuslog.InfoLevel, Writer: phuslog.IOWriter{Writer: io.Discard}, TimeField: ""}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf(msgf, 16, 32)
	}
}

func BenchmarkSlogAny(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	for i := 0; i < b.N; i++ {
		logger.Info(msg, "rate", "15", "low", 16, "object", &obj)
	}
}

func BenchmarkIQLogAny(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard, JSONTimeMode: iqlog.JSONTimeDisabled})
	for i := 0; i < b.N; i++ {
		// logger.Any("rate", "15").Any("low", 16).Any("object", &obj).Info(msg)
		logger.InfoEvent().Any("rate", "15").Any("low", 16).Any("object", &obj).Msg(msg)
		// logger.Info(msg)
	}
}

func BenchmarkZerologAny(b *testing.B) {
	logger := zerolog.New(io.Discard)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Any("object", &obj).Msg(msg)
	}
}

func BenchmarkPhusluAny(b *testing.B) {
	logger := phuslog.Logger{Level: phuslog.InfoLevel, Writer: phuslog.IOWriter{Writer: io.Discard}, TimeField: ""}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Any("object", &obj).Msg(msg)
	}
}

func BenchmarkIQLogSimpleWithTimestamp(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard, JSONTimeMode: iqlog.JSONTimeUTC})
	for i := 0; i < b.N; i++ {
		logger.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkZerologSimpleWithTimestamp(b *testing.B) {
	logger := zerolog.New(io.Discard).With().Timestamp().Logger()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

func BenchmarkPhusluSimpleWithTimestamp(b *testing.B) {
	logger := phuslog.Logger{Level: phuslog.InfoLevel, Writer: phuslog.IOWriter{Writer: io.Discard}, TimeField: "time", TimeLocation: time.UTC, TimeFormat: time.RFC3339Nano}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msg(msg)
	}
}

// func BenchmarkIQLogPrintfWithTimestamp(b *testing.B) {
// 	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard, JSONTimeMode: iqlog.JSONTimeUTC})
// 	for i := 0; i < b.N; i++ {
// 		logger.InfoEvent().Str("rate", "15").Int("low", 16).Float32("high", 123.2).Msgf(msg)
// 	}
// }

func BenchmarkIQLogParallel(b *testing.B) {
	logger := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: io.Discard, JSONTimeMode: iqlog.JSONTimeUTC})
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
