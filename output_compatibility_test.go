package iqlog

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"
)

func TestJSONDefaultsAndFieldOrder(t *testing.T) {
	sb := &syncBuffer{}
	now := time.Date(2026, time.August, 29, 12, 34, 56, 123000000, time.UTC)
	l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelDebug, Now: func() time.Time { return now }})

	l.ErrorEvent().Int("err_code", 500).Str("err_field", "").Str("err_message", "operation failed").
		Str("source_file", "apierror/examples/log-output/main.go:49").
		Str("source_func", "main.runCases.func1").Str("err_type", "ERROR").Msg("operation failed")

	want := "{\"time\":\"2026-08-29T12:34:56.123000\",\"level\":\"ERROR\",\"err_code\":500,\"err_field\":\"\",\"err_message\":\"operation failed\",\"source_file\":\"apierror/examples/log-output/main.go:49\",\"source_func\":\"main.runCases.func1\",\"err_type\":\"ERROR\",\"message\":\"operation failed\"}\n"
	if got := sb.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestJSONTimeCanBeDisabled(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, DisableTime: true, Writer: sb})
	l.InfoEvent().Msg("message")
	if matched := regexp.MustCompile(`^\{"level":"INFO","message":"message"\}\n$`).MatchString(sb.String()); !matched {
		t.Fatalf("unexpected output %q", sb.String())
	}
	var record map[string]any
	if err := json.Unmarshal([]byte(sb.String()), &record); err != nil {
		t.Fatal(err)
	}
	if _, ok := record["time"]; ok {
		t.Fatal("time was not disabled")
	}
}

func TestEventAtConsoleCaller(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Writer: sb, CallerDepth: 1, DisableTime: true})
	l.EventAt(LevelError, "main.runCases.func1", "log-output/main.go:49").Msg("operation failed")
	want := "ERRR [main.runCases.func1 log-output/main.go:49] operation failed\n"
	if got := sb.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
