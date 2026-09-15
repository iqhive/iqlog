package iqlog

import (
	"context"
	"io"
	"log/slog"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// callerAttributionHelper invokes call through a func value so the closure
// is a real stack frame: it is the application frame that enters the library
// and therefore the frame CallerDepth 1 must report.
//
//go:noinline
func callerAttributionHelper(l *Logger, call func(*Logger)) {
	call(l)
}

// closureLine returns the source line on which the closure fn is defined.
// Every case closure below is written on a single line, so this is also the
// line of its logging call.
func closureLine(t *testing.T, fn func(*Logger)) (name string, line int) {
	t.Helper()
	pc := reflect.ValueOf(fn).Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		t.Fatal("closure has no function")
	}
	_, line = f.FileLine(f.Entry())
	return f.Name(), line
}

func TestCallerAttributionEntryPoints(t *testing.T) {
	type entry struct {
		name string
		call func(*Logger)
	}
	// Every closure below must stay on one line: closureLine reports the
	// line of the closure's definition as the expected call site.
	ctx, kv := context.Background(), map[string]any{"k": "v"}
	cases := []entry{
		{"logger info", func(l *Logger) { l.Info("m") }},
		{"logger infof", func(l *Logger) { l.Infof("%s", "m") }},
		{"logger infoln", func(l *Logger) { l.Infoln("m") }},
		{"logger info event", func(l *Logger) { l.InfoEvent().Msg("m") }},
		{"package info", func(*Logger) { Info("m") }},
		{"package infof", func(*Logger) { Infof("%s", "m") }},
		{"package infoln", func(*Logger) { Infoln("m") }},
		{"package info event", func(*Logger) { InfoEvent().Msg("m") }},
		{"logger log", func(l *Logger) { l.Log(LevelInfo, "m") }},
		{"logger logf", func(l *Logger) { l.Logf(LevelInfo, "%s", "m") }},
		{"logger logln", func(l *Logger) { l.Logln(LevelInfo, "m") }},
		{"logger log context", func(l *Logger) { l.LogContext(ctx, LevelInfo, "m") }},
		{"logger log contextf", func(l *Logger) { l.LogContextf(ctx, LevelInfo, "%s", "m") }},
		{"logger print", func(l *Logger) { l.Print("m") }},
		{"logger printf", func(l *Logger) { l.Printf("%s", "m") }},
		{"logger println", func(l *Logger) { l.Println("m") }},
		{"logger warning", func(l *Logger) { l.Warning("m") }},
		{"logger warningf", func(l *Logger) { l.Warningf("%s", "m") }},
		{"logger warningln", func(l *Logger) { l.Warningln("m") }},
		{"package warning", func(*Logger) { Warning("m") }},
		{"package warningf", func(*Logger) { Warningf("%s", "m") }},
		{"package warningln", func(*Logger) { Warningln("m") }},
		{"log with fields", func(l *Logger) { l.LogWithFields(ctx, LevelInfo, kv, "m") }},
		{"log f with fields", func(l *Logger) { l.LogFWithFields(ctx, LevelInfo, kv, "%s", "m") }},
		{"trace with", func(l *Logger) { l.TraceWith().Msg("m") }},
		{"warn with", func(l *Logger) { l.WarnWith().Msg("m") }},
		{"error with", func(l *Logger) { l.ErrorWith().Msg("m") }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &syncBuffer{}
			l := MustNew(Config{Format: FormatJSON, Writer: buf, Level: LevelTrace, CallerDepth: 1})
			if strings.HasPrefix(tc.name, "package ") {
				previous := Default()
				t.Cleanup(func() { SetDefault(previous) })
				SetDefault(l)
			}
			wantFunc, line := closureLine(t, tc.call)
			callerAttributionHelper(l, tc.call)
			record := lastJSONMap(t, buf)
			fn, _ := record["func"].(string)
			file, _ := record["file"].(string)
			if fn != wantFunc {
				t.Fatalf("func %q want %q", fn, wantFunc)
			}
			if want := "caller_attribution_test.go:" + strconv.Itoa(line); file != want {
				t.Fatalf("file %q want %q", file, want)
			}
		})
	}
}

//go:noinline
func callerDepthOne(l *Logger) { l.InfoEvent().Msg("depth") }

//go:noinline
func callerDepthTwo(l *Logger) { callerDepthOne(l) }

//go:noinline
func callerDepthThree(l *Logger) { callerDepthTwo(l) }

func TestCallerAttributionDepth(t *testing.T) {
	for depth, want := range []string{"callerDepthOne", "callerDepthTwo", "callerDepthThree"} {
		t.Run(strconv.Itoa(depth+1), func(t *testing.T) {
			buf := &syncBuffer{}
			l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: depth + 1})
			callerDepthThree(l)
			record := lastJSONMap(t, buf)
			fn, _ := record["func"].(string)
			if !strings.HasSuffix(fn, "."+want) {
				t.Fatalf("func %q want suffix .%s", fn, want)
			}
		})
	}
}

func TestCallerAttributionNeverIqlogFrame(t *testing.T) {
	paths := []string{"callers.go:", "global_funcs.go:", "logger_funcs.go:", "compatibility.go:", "legacy.go:", "iqlog.go:", "bytesliceLine_"}
	cases := []struct {
		name string
		call func(*Logger)
	}{
		{"info", func(l *Logger) { l.Info("m") }},
		{"info event", func(l *Logger) { l.InfoEvent().Msg("m") }},
		{"log", func(l *Logger) { l.Log(LevelInfo, "m") }},
		{"print", func(l *Logger) { l.Print("m") }},
		{"warning", func(l *Logger) { l.Warning("m") }},
		{"fields", func(l *Logger) { l.LogWithFields(context.Background(), LevelInfo, map[string]any{"k": "v"}, "m") }},
		{"trace with", func(l *Logger) { l.TraceWith().Msg("m") }},
	}
	for _, format := range []Format{FormatConsole, FormatJSON} {
		t.Run(strconv.Itoa(int(format)), func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					buf := &syncBuffer{}
					l := MustNew(Config{Format: format, Writer: buf, Level: LevelTrace, CallerDepth: 1})
					callerAttributionHelper(l, tc.call)
					out := buf.String()
					if format == FormatJSON {
						record := lastJSONMap(t, buf)
						file, fileOK := record["file"].(string)
						fn, funcOK := record["func"].(string)
						if !fileOK || !funcOK {
							t.Fatalf("caller not emitted: %#v", record)
						}
						out = file + " " + fn
					} else if !strings.Contains(out, "caller_attribution_test.go:") {
						t.Fatalf("caller not emitted in %q", out)
					}
					for _, path := range paths {
						if strings.Contains(out, path) {
							t.Fatalf("iqlog frame %q in %q", path, out)
						}
					}
				})
			}
		})
	}
}

func TestCallerAttributionEventAt(t *testing.T) {
	tests := []struct {
		name         string
		function     string
		file         string
		wantFunction string
		wantFile     string
		wantCaller   bool
	}{
		{"explicit caller", "app.fn", "app.go:12", "app.fn", "app.go:12", true},
		{"function only", "app.fn", "", "app.fn", "", true},
		{"file only", "", "app.go:12", "", "", false},
		{"empty caller", "", "", "", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &syncBuffer{}
			l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
			l.EventAt(LevelInfo, tc.function, tc.file).Msg("m")
			record := lastJSONMap(t, buf)
			fn, hasFunc := record["func"].(string)
			file, hasFile := record["file"].(string)
			// A non-empty function is required; a file alone is ignored.
			if hasFunc != tc.wantCaller || hasFile != tc.wantCaller {
				t.Fatalf("caller keys present func=%v file=%v record=%#v", hasFunc, hasFile, record)
			}
			if fn != tc.wantFunction || file != tc.wantFile {
				t.Fatalf("caller func=%q file=%q, want func=%q file=%q", fn, file, tc.wantFunction, tc.wantFile)
			}
		})
	}
}

//go:noinline
func callerDeepChain(l *Logger, remaining int) int {
	if remaining == 0 {
		_, _, line, _ := runtime.Caller(0)
		l.Info("deep")
		return line + 1
	}
	return callerDeepChain(l, remaining-1)
}

func TestCallerAttributionDeepStack(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	wantLine := callerDeepChain(l, maxCallerFrames+8)
	// Parsing the output proves that exhausting the bounded frame buffer did
	// not panic or leave a partial record.
	record := lastJSONMap(t, buf)
	fn, hasFunc := record["func"].(string)
	file, hasFile := record["file"].(string)
	if hasFunc != hasFile {
		t.Fatalf("partial caller func=%q file=%q record=%#v", fn, file, record)
	}
	if hasFunc {
		if !strings.HasSuffix(fn, ".callerDeepChain") {
			t.Fatalf("func %q is neither absent nor the expected outer recursive frame", fn)
		}
		wantFile := "caller_attribution_test.go:" + strconv.Itoa(wantLine)
		if file != wantFile {
			t.Fatalf("file %q is neither absent nor the expected outer recursive frame %q", file, wantFile)
		}
	}
	if hasFile && strings.Contains(file, "iqlog.go:") {
		t.Fatalf("iqlog frame %q", file)
	}
}

//go:noinline
func callerAttributionMultiline(l *Logger) int {
	_, _, line, _ := runtime.Caller(0)
	l.Info(
		"multiline")
	return line + 1
}

func TestCallerAttributionUsesCallSiteLine(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	wantLine := callerAttributionMultiline(l)
	record := lastJSONMap(t, buf)
	file, ok := record["file"].(string)
	if !ok {
		t.Fatalf("caller file not emitted: %#v", record)
	}
	want := "caller_attribution_test.go:" + strconv.Itoa(wantLine)
	if file != want {
		t.Fatalf("file %q want call site %q", file, want)
	}
}

func TestCallerAttributionAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("allocation counts are unstable under race")
	}
	zero := MustNew(Config{Format: FormatJSON, Writer: io.Discard, CallerDepth: 0})
	one := MustNew(Config{Format: FormatJSON, Writer: io.Discard, CallerDepth: 1})
	zeroAllocs := testing.AllocsPerRun(1000, func() { zero.InfoEvent().Str("k", "v").Msg("m") })
	oneAllocs := testing.AllocsPerRun(1000, func() { one.InfoEvent().Str("k", "v").Msg("m") })
	t.Logf("CallerDepth 0 allocations: %f", zeroAllocs)
	t.Logf("CallerDepth 1 allocations: %f", oneAllocs)
	if zeroAllocs != 0 {
		t.Fatalf("CallerDepth 0 allocations = %f, want 0", zeroAllocs)
	}
	// The stack scan uses runtime.CallersFrames because, as documented by
	// runtime.Callers, iterating raw PCs cannot account for inlining or adjust
	// return PCs to call PCs correctly. The measured cost is the CallersFrames
	// iterator plus inline-frame resolution; it must not exceed the 2.0 ceiling.
	const ceilingOneAllocs = 2.0
	if oneAllocs > ceilingOneAllocs {
		t.Fatalf("CallerDepth 1 allocations = %f, exceeds ceiling %f (measured)", oneAllocs, ceilingOneAllocs)
	}
}

// slogRecordWithLibraryPC builds a record whose program counter resolves to an
// iqlog method, exercising the handler's internal-frame fallback scan.
//
//go:noinline
func slogRecordWithLibraryPC() slog.Record {
	pc := reflect.ValueOf((*Logger).Info).Pointer()
	return slog.NewRecord(time.Time{}, slog.LevelInfo, "m", pc)
}

func TestSlogCallerFallsBackForInternalPC(t *testing.T) {
	buf := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: buf, CallerDepth: 1})
	if err := l.SlogHandler().Handle(context.Background(), slogRecordWithLibraryPC()); err != nil {
		t.Fatal(err)
	}
	record := lastJSONMap(t, buf)
	file, _ := record["file"].(string)
	fn, _ := record["func"].(string)
	if !strings.HasSuffix(fn, ".TestSlogCallerFallsBackForInternalPC") {
		t.Fatalf("fallback caller func %q, want test frame", fn)
	}
	if !strings.HasPrefix(file, "caller_attribution_test.go:") {
		t.Fatalf("fallback caller file %q, want caller_attribution_test.go", file)
	}
	if strings.Contains(file, "iqlog.go") {
		t.Fatalf("fallback reported iqlog frame: func=%q file=%q", fn, file)
	}
}

func TestCaptureCallerPCSkipsInfraFrame(t *testing.T) {
	pc := reflect.ValueOf(runtime.Callers).Pointer()
	var data callerData
	captureCallerPC(pc, 1, &data)
	if strings.Contains(string(data.callerFunc[:data.callerFuncLen]), "runtime") {
		t.Fatalf("runtime caller was emitted: func=%q file=%q", data.callerFunc[:data.callerFuncLen], data.callerFile[:data.callerFileLen])
	}
}

func TestPackageWrappersAndLegacyHelpers(t *testing.T) {
	old := Default()
	t.Cleanup(func() { SetDefault(old) })
	buf := &syncBuffer{}
	exitCodes := make([]int, 0, 2)
	l := MustNew(Config{Format: FormatJSON, Writer: buf, Level: LevelTrace, ExitFunc: func(code int) {
		exitCodes = append(exitCodes, code)
	}})
	SetDefault(l)

	cases := []struct {
		name string
		call func()
	}{
		{"trace", func() { Trace("trace") }}, {"tracef", func() { Tracef("trace%s", "f") }}, {"traceln", func() { Traceln("trace", "ln") }},
		{"debug", func() { Debug("debug") }}, {"debugf", func() { Debugf("debug%s", "f") }}, {"debugln", func() { Debugln("debug", "ln") }},
		{"info", func() { Info("info") }}, {"infof", func() { Infof("info%s", "f") }}, {"infoln", func() { Infoln("info", "ln") }},
		{"warn", func() { Warn("warn") }}, {"warnf", func() { Warnf("warn%s", "f") }}, {"warnln", func() { Warnln("warn", "ln") }},
		{"error", func() { Error("error") }}, {"errorf", func() { Errorf("error%s", "f") }}, {"errorln", func() { Errorln("error", "ln") }},
		{"fatal", func() { Fatal("fatal") }}, {"fatalf", func() { Fatalf("fatal%s", "f") }}, {"fatalln", func() { Fatalln("fatal", "ln") }},
		{"panic", func() { Panic("panic") }}, {"panicf", func() { Panicf("panic%s", "f") }}, {"panicln", func() { Panicln("panic", "ln") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if tc.name == "panic" || tc.name == "panicf" || tc.name == "panicln" {
					if recover() == nil {
						t.Fatal("panic wrapper did not panic")
					}
				}
			}()
			tc.call()
		})
	}

	Print("print")
	Printf("printf%s", "f")
	Println("println", "line")
	Logln(LevelInfo, "logln", "line")
	if got := WithError(context.Canceled); got == nil {
		t.Fatal("package WithError returned nil")
	}
	if got := WithFields(map[string]any{"field": "value"}).WithError(context.Canceled); got == nil {
		t.Fatal("WithFields/WithError returned nil")
	}
	if len(exitCodes) != 3 {
		t.Fatalf("fatal exit calls = %d, want 3", len(exitCodes))
	}
}

func TestCallerAttributionFramePredicates(t *testing.T) {
	cases := []struct {
		name         string
		function     string
		file         string
		wantInternal bool
		wantInfra    bool
	}{
		{"library frame", "github.com/iqhive/iqlog.Warnf", "compatibility.go", true, false},
		{"library method", "github.com/iqhive/iqlog.(*Logger).Print", "legacy.go", true, false},
		{"legacy v3", "bitbucket.org/iqhive/iqlog/v3.Warnf", "compat.go", true, false},
		{"legacy v4", "bitbucket.org/iqhive/iqlog/v4.X", "x.go", true, false},
		{"sibling with dash", "github.com/iqhive/iqlog-app.Foo", "main.go", false, false},
		{"sibling with suffix", "github.com/iqhive/iqlogx.Y", "y.go", false, false},
		{"in-package test helper", "github.com/iqhive/iqlog.helper", "caller_attribution_test.go", false, false},
		{"runtime", "runtime.Callers", "extern.go", false, true},
		{"stdlib log", "log.Printf", "log.go", false, true},
		{"stdlib slog", "log/slog.Info", "logger.go", false, true},
		{"stdlib slog method", "log/slog.(*Logger).Info", "logger.go", false, true},
		{"user module starting log", "log.example.com/x.Fn", "log.go", false, false},
		{"user module starting runtime", "runtime.example/x.Fn", "runtime.go", false, false},
		{"user package named log", "github.com/acme/log.Foo", "log.go", false, false},
		{"empty name", "", "", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isInternalFrame(tc.function, tc.file); got != tc.wantInternal {
				t.Errorf("isInternalFrame(%q, %q) = %v, want %v", tc.function, tc.file, got, tc.wantInternal)
			}
			if got := isInfraFrame(tc.function); got != tc.wantInfra {
				t.Errorf("isInfraFrame(%q) = %v, want %v", tc.function, got, tc.wantInfra)
			}
		})
	}
}
