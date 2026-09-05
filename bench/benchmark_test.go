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

// ---------------------------------------------------------------------------
// log/slog handlers.
//
// Every handler below is driven through a *slog.Logger with the same typed
// attributes, so the numbers include slog's own record construction and PC
// capture, which no handler can avoid. Payloads are kept equivalent: no
// timestamp unless the name says so, no source unless the name says so. The
// standard library JSON handler always emits a timestamp, so its Simple
// number is comparable with the WithTimestamp variants of the others.
// ---------------------------------------------------------------------------

func iqlogSlog(cfg iqlog.Config) *slog.Logger {
	cfg.Writer = io.Discard
	return slog.New(iqlog.MustNew(cfg).SlogHandler())
}

func phusluSlog(l phuslog.Logger) *slog.Logger {
	l.Writer = phuslog.IOWriter{Writer: io.Discard}
	return l.Slog()
}

func zerologSlog(l zerolog.Logger) *slog.Logger {
	return slog.New(zerolog.NewSlogHandler(l))
}

func runSlogDisabled(b *testing.B, logger *slog.Logger) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(context.Background(), slog.LevelDebug, msg, slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
	}
}

func runSlogSimple(b *testing.B, logger *slog.Logger) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(context.Background(), slog.LevelInfo, msg, slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
	}
}

func runSlogGrouped(b *testing.B, logger *slog.Logger) {
	logger = logger.WithGroup("http").With("method", "GET", "path", "/orders")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(context.Background(), slog.LevelInfo, msg, slog.Int("status", 200), slog.String("client", "10.0.0.1"))
	}
}

func runSlogParallel(b *testing.B, logger *slog.Logger) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.LogAttrs(context.Background(), slog.LevelInfo, msg, slog.String("rate", "15"), slog.Int("low", 16), slog.Float64("high", 123.2))
		}
	})
}

func BenchmarkSlogHandlerDisabled(b *testing.B) {
	runSlogDisabled(b, slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func BenchmarkIQLogSlogHandlerDisabled(b *testing.B) {
	runSlogDisabled(b, iqlogSlog(iqlog.Config{Format: iqlog.FormatJSON, Level: iqlog.LevelInfo}))
}

func BenchmarkZerologSlogHandlerDisabled(b *testing.B) {
	runSlogDisabled(b, zerologSlog(zerolog.New(io.Discard).Level(zerolog.InfoLevel)))
}

func BenchmarkPhusluSlogHandlerDisabled(b *testing.B) {
	runSlogDisabled(b, phusluSlog(phuslog.Logger{Level: phuslog.InfoLevel, TimeField: ""}))
}

func BenchmarkSlogHandlerSimple(b *testing.B) {
	// slog.JSONHandler always writes a timestamp; compare with the
	// WithTimestamp variants of the other handlers.
	runSlogSimple(b, slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func BenchmarkIQLogSlogHandlerSimple(b *testing.B) {
	runSlogSimple(b, iqlogSlog(iqlog.Config{Format: iqlog.FormatJSON}))
}

func BenchmarkZerologSlogHandlerSimple(b *testing.B) {
	runSlogSimple(b, zerologSlog(zerolog.New(io.Discard)))
}

func BenchmarkPhusluSlogHandlerSimple(b *testing.B) {
	runSlogSimple(b, phusluSlog(phuslog.Logger{Level: phuslog.InfoLevel, TimeField: ""}))
}

func BenchmarkIQLogSlogHandlerSimpleWithTimestamp(b *testing.B) {
	runSlogSimple(b, iqlogSlog(iqlog.Config{Format: iqlog.FormatJSON, JSONTimeMode: iqlog.JSONTimeUTC}))
}

func BenchmarkZerologSlogHandlerSimpleWithTimestamp(b *testing.B) {
	// zerolog's handler writes the record time itself when the logger has
	// no timestamp hook, so a plain logger already carries the timestamp.
	runSlogSimple(b, zerologSlog(zerolog.New(io.Discard)))
}

func BenchmarkPhusluSlogHandlerSimpleWithTimestamp(b *testing.B) {
	runSlogSimple(b, phusluSlog(phuslog.Logger{Level: phuslog.InfoLevel, TimeField: "time", TimeLocation: time.UTC, TimeFormat: time.RFC3339Nano}))
}

func BenchmarkSlogHandlerWithSource(b *testing.B) {
	runSlogSimple(b, slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{AddSource: true})))
}

func BenchmarkIQLogSlogHandlerWithSource(b *testing.B) {
	runSlogSimple(b, iqlogSlog(iqlog.Config{Format: iqlog.FormatJSON, CallerDepth: 1}))
}

func BenchmarkPhusluSlogHandlerWithSource(b *testing.B) {
	runSlogSimple(b, phusluSlog(phuslog.Logger{Level: phuslog.InfoLevel, TimeField: "", Caller: 1}))
}

func BenchmarkSlogHandlerGrouped(b *testing.B) {
	runSlogGrouped(b, slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func BenchmarkIQLogSlogHandlerGrouped(b *testing.B) {
	runSlogGrouped(b, iqlogSlog(iqlog.Config{Format: iqlog.FormatJSON}))
}

func BenchmarkZerologSlogHandlerGrouped(b *testing.B) {
	runSlogGrouped(b, zerologSlog(zerolog.New(io.Discard)))
}

func BenchmarkPhusluSlogHandlerGrouped(b *testing.B) {
	runSlogGrouped(b, phusluSlog(phuslog.Logger{Level: phuslog.InfoLevel, TimeField: ""}))
}

func BenchmarkSlogHandlerConsole(b *testing.B) {
	runSlogSimple(b, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func BenchmarkIQLogSlogHandlerConsole(b *testing.B) {
	runSlogSimple(b, iqlogSlog(iqlog.Config{Format: iqlog.FormatConsole, DisableColor: true}))
}

func BenchmarkSlogHandlerParallel(b *testing.B) {
	runSlogParallel(b, slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func BenchmarkIQLogSlogHandlerParallel(b *testing.B) {
	runSlogParallel(b, iqlogSlog(iqlog.Config{Format: iqlog.FormatJSON}))
}

func BenchmarkZerologSlogHandlerParallel(b *testing.B) {
	runSlogParallel(b, zerologSlog(zerolog.New(io.Discard)))
}

func BenchmarkPhusluSlogHandlerParallel(b *testing.B) {
	runSlogParallel(b, phusluSlog(phuslog.Logger{Level: phuslog.InfoLevel, TimeField: ""}))
}
