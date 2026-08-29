package iqlog

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestEventFieldTypesAndSingleUse(t *testing.T) {
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb})
	now := time.Date(2026, 1, 2, 3, 4, 5, 6, time.UTC)
	e := l.InfoEvent().Str("str", "value").Int("int", -1).Int64("int64", -2).
		Uint("uint", 3).Uint64("uint64", 4).Float32("float32", 1.5).
		Float64("float64", 2.5).Bool("bool", true).Err(errors.New("boom")).
		Time("time", now).Duration("duration", time.Second+2*time.Millisecond).
		Bytes("bytes", []byte{0, 1, 2}).RawJSON("raw", []byte(`{"nested":true}`)).
		RawJSON("invalid", []byte(`{`)).Any("any", struct{ N int }{1})
	e.Msg("message")
	e.Msg("second")

	if strings.Count(sb.String(), "\n") != 1 {
		t.Fatalf("event wrote more than once: %q", sb.String())
	}
	var record map[string]any
	if err := json.Unmarshal([]byte(sb.String()), &record); err != nil {
		t.Fatal(err)
	}
	if record["str"] != "value" || record["int"] != float64(-1) || record["uint64"] != float64(4) || record["bool"] != true {
		t.Fatalf("unexpected record: %#v", record)
	}
	if record["error"] != "boom" || record["field_time"] != now.Format(time.RFC3339Nano) || record["duration"] != "1.002s" {
		t.Fatalf("unexpected text fields: %#v", record)
	}
	if record["bytes"] != base64.StdEncoding.EncodeToString([]byte{0, 1, 2}) || record["invalid"] != nil {
		t.Fatalf("unexpected encoded fields: %#v", record)
	}
}

func TestContextExtractorAndFieldPrecedence(t *testing.T) {
	type key struct{}
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb, ContextExtractor: func(ctx context.Context) map[string]any {
		return map[string]any{"request": ctx.Value(key{}), "source": "context"}
	}}).WithFields(map[string]any{"source": "persistent", "stable": 1}).WithContext(context.WithValue(context.Background(), key{}, 42))
	l.Info("message")
	var record map[string]any
	if err := json.Unmarshal([]byte(sb.String()), &record); err != nil {
		t.Fatal(err)
	}
	if record["request"] != float64(42) || record["source"] != "context" || record["stable"] != float64(1) {
		t.Fatalf("unexpected fields: %#v", record)
	}
}

func TestTimestampLayouts(t *testing.T) {
	for _, layout := range []string{"2006", "2006-01-02T15:04:05.000", "2006-01-02T15:04:05.000000", time.RFC3339Nano, time.RFC822} {
		t.Run(layout, func(t *testing.T) {
			sb := &syncBuffer{}
			l := MustNew(Config{Format: FormatJSON, Writer: sb, JSONTimeMode: JSONTimeCustom, TimestampLayout: layout})
			l.Info("message")
			var record map[string]any
			if err := json.Unmarshal([]byte(sb.String()), &record); err != nil {
				t.Fatal(err)
			}
			if record["time"] == "" {
				t.Fatal("missing timestamp")
			}
		})
	}
}

func TestInjectedClockAndReservedFields(t *testing.T) {
	sb := &syncBuffer{}
	now := time.Date(2026, 2, 3, 4, 5, 6, 7, time.UTC)
	l := MustNew(Config{Format: FormatJSON, Writer: sb, IncludeTime: true, Now: func() time.Time { return now }})
	l.InfoEvent().Str("message", "field").Str("level", "field").Msg("actual")
	var record map[string]any
	if err := json.Unmarshal([]byte(sb.String()), &record); err != nil {
		t.Fatal(err)
	}
	if record["message"] != "actual" || record["field_message"] != "field" || record["field_level"] != "field" {
		t.Fatalf("reserved fields: %#v", record)
	}
	if record["time"] != now.UTC().Format("2006-01-02T15:04:05.000000Z") {
		t.Fatalf("clock not used: %#v", record["time"])
	}
}

func TestInvalidRawJSONReportsBuildError(t *testing.T) {
	e := MustNew(Config{Writer: io.Discard}).InfoEvent().RawJSON("raw", []byte("{"))
	if e.BuildError() == nil {
		t.Fatal("expected build error")
	}
}

func TestDiscardConsumesEvent(t *testing.T) {
	sb := &syncBuffer{}
	e := MustNew(Config{Writer: sb}).InfoEvent().Str("key", "value")
	e.Discard()
	e.Msg("ignored")
	if sb.String() != "" {
		t.Fatalf("discard wrote %q", sb.String())
	}
}

func TestInjectedFatalExit(t *testing.T) {
	sb := &syncBuffer{}
	code := 0
	l := MustNew(Config{Writer: sb, ExitFunc: func(got int) { code = got }})
	l.FatalEvent().Msg("fatal")
	if code != 1 || !strings.Contains(sb.String(), "fatal") {
		t.Fatalf("code=%d output=%q", code, sb.String())
	}

	code = 0
	filtered := MustNew(Config{Writer: sb, Level: LevelFatal + 1, ExitFunc: func(got int) { code = got }})
	before := sb.String()
	filtered.FatalEvent().Msg("filtered")
	if code != 1 || sb.String() != before {
		t.Fatalf("filtered code=%d output=%q", code, sb.String())
	}
}
