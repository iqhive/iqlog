package iqlog

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
)

func TestNamedLoggerAndConstructors(t *testing.T) {
	var first *Logger = MustNew(Config{})
	var second *Logger = MustNew(Config{IncludeTime: true})
	if first == nil || second == nil {
		t.Fatal("constructors returned nil")
	}
}

func TestSetDefaultRoutesPackageFunctions(t *testing.T) {
	old := Default()
	t.Cleanup(func() { SetDefault(old) })

	l := MustNew(Config{})
	l.setIncludeTime(false)
	l.setCallerDepth(0)
	l.setUseColor(false)
	sb := &syncBuffer{}
	l.setWriter(sb)
	SetDefault(l)

	Info("default")
	if got, want := sb.String(), "INFO default\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSetDefaultRejectsNil(t *testing.T) {
	defer func() {
		if got, want := recover(), any("iqlog: nil default logger"); got != want {
			t.Fatalf("panic = %#v, want %#v", got, want)
		}
	}()
	SetDefault(nil)
}

func TestWithFieldsUsesOriginatingLoggerAndPreservesTypes(t *testing.T) {
	l := MustNew(Config{Format: FormatJSON})
	l.setIncludeTime(false)
	l.setCallerDepth(0)
	sb := &syncBuffer{}
	l.setWriter(sb)

	baseFields := map[string]any{"name": "first", "count": 2, "ok": true}
	derived := l.WithFields(baseFields).WithFields(map[string]any{"name": "second", "ratio": 1.5})
	baseFields["count"] = 99
	derived.Info("message")

	var record map[string]any
	if err := json.Unmarshal([]byte(sb.String()), &record); err != nil {
		t.Fatalf("invalid JSON %q: %v", sb.String(), err)
	}
	if record["name"] != "second" || record["count"] != float64(2) || record["ok"] != true || record["ratio"] != 1.5 {
		t.Fatalf("unexpected fields: %#v", record)
	}
}

func TestWithErrorAndContextChaining(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "value")
	l := MustNew(Config{Format: FormatJSON})
	l.setIncludeTime(false)
	l.setCallerDepth(0)
	sb := &syncBuffer{}
	l.setWriter(sb)

	derived := l.WithContext(ctx).WithFields(map[string]any{"request": 7}).WithError(errors.New("boom"))
	if derived.ctx != ctx {
		t.Fatal("context was not preserved through WithFields and WithError")
	}
	derived.Error("failed")

	var record map[string]any
	if err := json.Unmarshal([]byte(sb.String()), &record); err != nil {
		t.Fatalf("invalid JSON %q: %v", sb.String(), err)
	}
	if record["request"] != float64(7) || record["error"] != "boom" {
		t.Fatalf("unexpected fields: %#v", record)
	}
}

func TestLogAndLogContextAPIs(t *testing.T) {
	l := MustNew(Config{})
	l.setIncludeTime(false)
	l.setCallerDepth(0)
	l.setUseColor(false)
	sb := &syncBuffer{}
	l.setWriter(sb)

	l.Log(LevelInfo, "plain", 1)
	l.Logf(LevelWarn, "formatted=%d", 2)
	l.Logln(LevelError, "line", 3)
	l.LogContext(context.Background(), LevelInfo, "context", 4)
	l.LogContextf(context.Background(), LevelWarn, "context=%d", 5)

	want := "INFO plain 1\nWARN formatted=2\nERRR line 3\nINFO context 4\nWARN context=5\n"
	if got := sb.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestPackageLogAndLogContextAPIs(t *testing.T) {
	old := Default()
	t.Cleanup(func() { SetDefault(old) })

	l := MustNew(Config{})
	l.setIncludeTime(false)
	l.setCallerDepth(0)
	l.setUseColor(false)
	sb := &syncBuffer{}
	l.setWriter(sb)
	SetDefault(l)

	Log(LevelInfo, "plain", 1)
	Logf(LevelWarn, "formatted=%d", 2)
	LogContext(context.Background(), LevelError, "context", 3)
	LogContextf(context.Background(), LevelInfo, "context=%d", 4)

	want := "INFO plain 1\nWARN formatted=2\nERRR context 3\nINFO context=4\n"
	if got := sb.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestConcurrentDefaultAccess(t *testing.T) {
	old := Default()
	t.Cleanup(func() { SetDefault(old) })

	loggers := []*Logger{MustNew(Config{}), MustNew(Config{})}
	for _, l := range loggers {
		l.setIncludeTime(false)
		l.setCallerDepth(0)
		l.setWriter(&syncBuffer{})
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			SetDefault(loggers[i%len(loggers)])
		}(i)
		go func() {
			defer wg.Done()
			Info("concurrent")
		}()
	}
	wg.Wait()
}
