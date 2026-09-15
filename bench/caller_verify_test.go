package benchmarks

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"

	legacy "bitbucket.org/iqhive/iqlog/v3"
	"github.com/iqhive/iqlog"
)

//go:noinline
func directInfo(l *iqlog.Logger) { l.Info("direct") }

//go:noinline
func directInfoEvent(l *iqlog.Logger) { l.InfoEvent().Msg("direct event") }

//go:noinline
func packageInfo(l *iqlog.Logger) {
	previous := iqlog.Default()
	iqlog.SetDefault(l)
	defer iqlog.SetDefault(previous)
	iqlog.Info("package")
}

//go:noinline
func slogInfo(l *iqlog.Logger) { slog.New(l.SlogHandler()).Info("slog") }

//go:noinline
func legacyWarnf() { legacy.Warnf("legacy %s", "formatted") }

//go:noinline
func legacyWarn() { legacy.Warn("legacy") }

func callerRecord(t *testing.T, output string) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal([]byte(output), &record); err != nil {
		t.Fatalf("decode log output %q: %v", output, err)
	}
	return record
}

func assertCaller(t *testing.T, output, helper string) {
	t.Helper()
	if strings.Contains(output, "bitbucket.org/iqhive/iqlog/v3") || strings.Contains(output, "github.com/iqhive/iqlog.") {
		t.Fatalf("caller contains library frame: %q", output)
	}
	if !strings.Contains(output, helper) {
		t.Fatalf("caller output %q does not contain helper %q", output, helper)
	}
}

func TestExternalCallerAttribution(t *testing.T) {
	cases := []struct {
		name   string
		helper string
		call   func(*iqlog.Logger, io.Writer)
	}{
		{"direct", "directInfo", func(l *iqlog.Logger, _ io.Writer) { directInfo(l) }},
		{"direct event", "directInfoEvent", func(l *iqlog.Logger, _ io.Writer) { directInfoEvent(l) }},
		{"package-level", "packageInfo", func(l *iqlog.Logger, _ io.Writer) { packageInfo(l) }},
		{"slog", "slogInfo", func(l *iqlog.Logger, _ io.Writer) { slogInfo(l) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, format := range []iqlog.Format{iqlog.FormatConsole, iqlog.FormatJSON} {
				t.Run(formatName(format), func(t *testing.T) {
					var buf bytes.Buffer
					l := iqlog.MustNew(iqlog.Config{Format: format, Writer: &buf, Level: iqlog.LevelTrace, CallerDepth: 1, DisableColor: true})
					tc.call(l, &buf)
					assertCaller(t, buf.String(), tc.helper)
					if format == iqlog.FormatJSON {
						record := callerRecord(t, buf.String())
						if _, ok := record["func"]; !ok {
							t.Fatalf("caller function missing: %#v", record)
						}
					}
				})
			}
		})
	}
}

func TestExternalLegacyShimCallerAttribution(t *testing.T) {
	for _, format := range []iqlog.Format{iqlog.FormatConsole, iqlog.FormatJSON} {
		t.Run(formatName(format), func(t *testing.T) {
			var buf bytes.Buffer
			shimLogger := legacy.NewGlobalIQLogger()
			shimLogger.Logger = iqlog.MustNew(iqlog.Config{Format: format, Writer: &buf, Level: iqlog.LevelTrace, CallerDepth: 1, DisableColor: true})
			shimLogger.SetCallerDepth(1)
			previous := legacy.GlobalLogger
			legacy.GlobalLogger = shimLogger
			t.Cleanup(func() { legacy.GlobalLogger = previous })

			legacyWarnf()
			assertCaller(t, buf.String(), "legacyWarnf")
			buf.Reset()
			legacyWarn()
			assertCaller(t, buf.String(), "legacyWarn")
			if format == iqlog.FormatJSON {
				_ = callerRecord(t, buf.String())
			}
		})
	}
}

func formatName(format iqlog.Format) string {
	if format == iqlog.FormatJSON {
		return "json"
	}
	return "console"
}
