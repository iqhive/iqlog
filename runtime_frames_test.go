package iqlog

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

// TestCallerBasic tests basic caller functionality
func TestCallerBasic(t *testing.T) {
	pc := Caller(0) // Current function
	if pc == 0 {
		t.Fatal("Caller returned zero PC")
	}

	name, file, line := pc.NameFileLine()

	// Verify we get reasonable results
	if !strings.Contains(name, "TestCallerBasic") {
		t.Errorf("Expected function name to contain 'TestCallerBasic', got: %s", name)
	}

	if !strings.Contains(file, "runtime_frames_test.go") {
		t.Errorf("Expected file to contain 'runtime_frames_test.go', got: %s", file)
	}

	if line <= 0 {
		t.Errorf("Expected positive line number, got: %d", line)
	}
}

// TestCallerDepth tests caller with different skip values
func TestCallerDepth(t *testing.T) {
	testFunc := func() PC {
		return Caller(1) // Skip this function to get the caller
	}

	pc := testFunc()
	name, file, line := pc.NameFileLine()

	// Should point to this test function, not testFunc
	if !strings.Contains(name, "TestCallerDepth") {
		t.Errorf("Expected function name to contain 'TestCallerDepth', got: %s", name)
	}

	if !strings.Contains(file, "runtime_frames_test.go") {
		t.Errorf("Expected file to contain 'runtime_frames_test.go', got: %s", file)
	}

	if line <= 0 {
		t.Errorf("Expected positive line number, got: %d", line)
	}
}

// TestFuncEntry tests function entry point capture
func TestFuncEntry(t *testing.T) {
	entry := FuncEntry(0)
	if entry == 0 {
		t.Fatal("FuncEntry returned zero PC")
	}

	// Entry should give us function info too
	name, file, line := entry.NameFileLine()

	if !strings.Contains(name, "TestFuncEntry") {
		t.Errorf("Expected function name to contain 'TestFuncEntry', got: %s", name)
	}

	if !strings.Contains(file, "runtime_frames_test.go") {
		t.Errorf("Expected file to contain 'runtime_frames_test.go', got: %s", file)
	}

	if line <= 0 {
		t.Errorf("Expected positive line number, got: %d", line)
	}
}

// TestCallers tests multiple frame capture
func TestCallers(t *testing.T) {
	pcs := Callers(0, 5) // Get up to 5 frames

	if len(pcs) == 0 {
		t.Fatal("Callers returned empty slice")
	}

	// First frame should be this function
	name, file, line := pcs[0].NameFileLine()

	if !strings.Contains(name, "TestCallers") {
		t.Errorf("Expected function name to contain 'TestCallers', got: %s", name)
	}

	if !strings.Contains(file, "runtime_frames_test.go") {
		t.Errorf("Expected file to contain 'runtime_frames_test.go', got: %s", file)
	}

	if line <= 0 {
		t.Errorf("Expected positive line number, got: %d", line)
	}
}

// TestCallersFill tests pre-allocated slice filling
func TestCallersFill(t *testing.T) {
	trace := make([]PC, 5)
	filled := CallersFill(0, trace)

	if len(filled) == 0 {
		t.Fatal("CallersFill returned empty slice")
	}

	// First frame should be this function
	name, file, line := filled[0].NameFileLine()

	if !strings.Contains(name, "TestCallersFill") {
		t.Errorf("Expected function name to contain 'TestCallersFill', got: %s", name)
	}

	if !strings.Contains(file, "runtime_frames_test.go") {
		t.Errorf("Expected file to contain 'runtime_frames_test.go', got: %s", file)
	}

	if line <= 0 {
		t.Errorf("Expected positive line number, got: %d", line)
	}
}

// TestFillCallerData tests the core caller data extraction
func TestFillCallerData(t *testing.T) {
	pc := Caller(0)
	var data callerData

	fillCallerData(pc, &data)

	// Check function name was captured
	if data.callerFuncLen == 0 {
		t.Error("Function name length is zero")
	}

	funcName := string(data.callerFunc[:data.callerFuncLen])
	if !strings.Contains(funcName, "TestFillCallerData") {
		t.Errorf("Expected function name to contain 'TestFillCallerData', got: %s", funcName)
	}

	// Check file name was captured
	if data.callerFileLen == 0 {
		t.Error("File name length is zero")
	}

	fileName := string(data.callerFile[:data.callerFileLen])
	if !strings.Contains(fileName, "runtime_frames_test.go") {
		t.Errorf("Expected file name to contain 'runtime_frames_test.go', got: %s", fileName)
	}
}

// TestPCFormatting tests PC string formatting
func TestPCFormatting(t *testing.T) {
	pc := Caller(0)

	// Test basic string formatting
	str := pc.String()
	if !strings.Contains(str, "runtime_frames_test.go:") {
		t.Errorf("Expected PC string to contain 'runtime_frames_test.go:', got: %s", str)
	}

	// Test format with different verbs
	testCases := []struct {
		format   string
		contains string
	}{
		{"%v", "runtime_frames_test.go:"},
		{"%+v", "runtime_frames_test.go"},
		{"%s", "TestPCFormatting"},
		{"%f", "runtime_frames_test.go"},
		{"%n", "TestPCFormatting"},
	}

	for _, tc := range testCases {
		result := fmt.Sprintf(tc.format, pc)
		if !strings.Contains(result, tc.contains) {
			t.Errorf("Format %s: expected to contain '%s', got: %s", tc.format, tc.contains, result)
		}
	}
}

// TestPCsFormatting tests multiple PC formatting
func TestPCsFormatting(t *testing.T) {
	pcs := Callers(0, 3)

	// Test basic string formatting
	str := pcs.String()
	if !strings.Contains(str, "runtime_frames_test.go:") {
		t.Errorf("Expected PCs string to contain 'runtime_frames_test.go:', got: %s", str)
	}

	// Test custom format
	formatted := pcs.FormatString("+")
	if !strings.Contains(formatted, "runtime_frames_test.go") {
		t.Errorf("Expected formatted string to contain 'runtime_frames_test.go', got: %s", formatted)
	}
}

// TestKeepNumDirs tests directory path truncation
func TestKeepNumDirs(t *testing.T) {
	testCases := []struct {
		input    string
		numDirs  int
		expected string
	}{
		{"/a/b/c/d/file.go", 2, "c/d/file.go"},
		{"/a/b/c/file.go", 1, "c/file.go"},
		{"/a/file.go", 2, "/a/file.go"},
		{"file.go", 1, "file.go"},
	}

	for _, tc := range testCases {
		result := KeepNumDirs(tc.input, tc.numDirs)
		if result != tc.expected {
			t.Errorf("KeepNumDirs(%q, %d) = %q, expected %q",
				tc.input, tc.numDirs, result, tc.expected)
		}
	}
}

// TestLoggerCallerDepthIntegration tests caller depth integration with logging
func TestLoggerCallerDepthIntegration(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.SetCallerDepth(1)
	logger.SetUseColour(false)

	// Helper function to log at different depth
	logFromHelper := func() {
		logger.Info("Test message from helper")
	}

	logFromHelper()

	output := buf.String()

	// Should show the test function in caller info (since CallerDepth=1)
	if !strings.Contains(output, "TestLoggerCallerDepthIntegration") {
		t.Errorf("Expected output to contain 'TestLoggerCallerDepthIntegration', got: %s", output)
	}
}

// TestLoggerWithCallerInfo tests caller info integration in structured logging
func TestLoggerWithCallerInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.SetCallerDepth(1)
	logger.SetUseColour(false)

	logger.WithByteSliceLineInfo().Msg("Test with caller info")

	output := buf.String()

	// Should include caller information - in this case, it captures the test runner
	// which is normal behavior for Go testing framework
	if !strings.Contains(output, "testing.") {
		t.Errorf("Expected output to contain caller info with 'testing.', got: %s", output)
	}
}

// TestLoggerJSONCallerInfo tests caller info in JSON mode
func TestLoggerJSONCallerInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true) // JSON mode
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.SetCallerDepth(1)

	logger.WithByteSliceLineInfo().Msg("Test JSON caller info")

	output := buf.String()

	// Should include function and file in JSON
	if !strings.Contains(output, `"func":`) {
		t.Errorf("Expected JSON output to contain '\"func\":', got: %s", output)
	}

	if !strings.Contains(output, `"file":`) {
		t.Errorf("Expected JSON output to contain '\"file\":', got: %s", output)
	}
}

// TestCallerDataBoundaries tests edge cases for caller data
func TestCallerDataBoundaries(t *testing.T) {
	pc := Caller(0)
	var data callerData

	fillCallerData(pc, &data)

	// Verify lengths are within bounds
	if data.callerFuncLen > callerDataMaxLen {
		t.Errorf("Function name length %d exceeds max %d", data.callerFuncLen, callerDataMaxLen)
	}

	if data.callerFileLen > callerDataMaxLen {
		t.Errorf("File name length %d exceeds max %d", data.callerFileLen, callerDataMaxLen)
	}

	// Verify no buffer overflow
	funcName := string(data.callerFunc[:data.callerFuncLen])
	fileName := string(data.callerFile[:data.callerFileLen])

	if len(funcName) != int(data.callerFuncLen) {
		t.Errorf("Function name length mismatch")
	}

	if len(fileName) != int(data.callerFileLen) {
		t.Errorf("File name length mismatch")
	}
}

// TestFunctionsToSkip tests frame skipping functionality
func TestFunctionsToSkip(t *testing.T) {
	// Verify the skip list contains expected entries
	expectedSkips := []string{"/iqlog/", "runtime.", "testing."}

	for _, expected := range expectedSkips {
		found := false
		for _, skip := range FunctionsToSkip {
			if skip == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected skip pattern '%s' not found in FunctionsToSkip", expected)
		}
	}
}

// TestPCZeroValue tests behavior with zero PC values
func TestPCZeroValue(t *testing.T) {
	var pc PC = 0

	name, file, line := pc.NameFileLine()

	// Zero PC should return empty/zero values gracefully
	if name != "" {
		t.Errorf("Expected empty name for zero PC, got: %s", name)
	}

	if file != "" {
		t.Errorf("Expected empty file for zero PC, got: %s", file)
	}

	if line != 0 {
		t.Errorf("Expected zero line for zero PC, got: %d", line)
	}
}

// TestCallerConsistency tests that caller results are consistent
func TestCallerConsistency(t *testing.T) {
	// Call Caller multiple times and ensure consistent results
	pc1 := Caller(0)
	pc2 := Caller(0)

	name1, file1, line1 := pc1.NameFileLine()
	name2, file2, line2 := pc2.NameFileLine()

	if name1 != name2 {
		t.Errorf("Inconsistent function names: %s vs %s", name1, name2)
	}

	if file1 != file2 {
		t.Errorf("Inconsistent file names: %s vs %s", file1, file2)
	}

	// Line numbers might be slightly different due to sequential calls
	if abs(line1-line2) > 5 {
		t.Errorf("Line numbers too different: %d vs %d", line1, line2)
	}
}

// TestDeepCallStack tests caller capture through multiple function levels
func TestDeepCallStack(t *testing.T) {
	level3 := func() PCs {
		return Callers(0, 5)
	}

	level2 := func() PCs {
		return level3()
	}

	level1 := func() PCs {
		return level2()
	}

	pcs := level1()

	if len(pcs) < 4 {
		t.Fatalf("Expected at least 4 frames, got %d", len(pcs))
	}

	// Verify the call stack order
	names := make([]string, len(pcs))
	for i, pc := range pcs {
		name, _, _ := pc.NameFileLine()
		names[i] = filepath.Base(name)
	}

	// Should show level3, level2, level1, TestDeepCallStack in order
	expectedFuncs := []string{"func1", "func2", "func3", "TestDeepCallStack"}
	for i, expected := range expectedFuncs {
		if i >= len(names) {
			break
		}
		if !strings.Contains(names[i], expected) {
			t.Errorf("Frame %d: expected to contain '%s', got '%s'", i, expected, names[i])
		}
	}
}

// TestRuntimeFramesGo123Compatibility tests Go version specific frame handling
func TestRuntimeFramesGo123Compatibility(t *testing.T) {
	// This test ensures our frame handling works across Go versions
	pc := Caller(0)

	// Basic functionality should work regardless of Go version
	name, file, line := pc.NameFileLine()

	if name == "" || file == "" || line == 0 {
		t.Error("Frame handling failed - possible Go version compatibility issue")
	}

	// Test formatting which uses internal frame structures
	formatted := fmt.Sprintf("%v", pc)
	if formatted == "" {
		t.Error("PC formatting failed - possible Go version compatibility issue")
	}
}

// Helper function for TestCallerConsistency
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestFillCallerDataTrimsMainModule verifies that the main module path is
// stripped from the caller function name, leaving a module-relative path.
func TestFillCallerDataTrimsMainModule(t *testing.T) {
	pc := Caller(0)
	var data callerData
	fillCallerData(pc, &data)

	funcName := string(data.callerFunc[:data.callerFuncLen])

	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Skip("build info unavailable")
	}

	prefixDot := info.Main.Path + "."
	prefixSlash := info.Main.Path + "/"
	if strings.HasPrefix(funcName, prefixDot) || strings.HasPrefix(funcName, prefixSlash) {
		t.Errorf("caller function name still contains main module prefix: %s", funcName)
	}
}

// Benchmark tests for performance verification
func BenchmarkCaller(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Caller(0)
	}
}

func BenchmarkFillCallerData(b *testing.B) {
	pc := Caller(0)
	var data callerData

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fillCallerData(pc, &data)
	}
}

func BenchmarkCallers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Callers(0, 5)
	}
}

func BenchmarkPCString(b *testing.B) {
	pc := Caller(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pc.String()
	}
}

func BenchmarkPCFormat(b *testing.B) {
	pc := Caller(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%v", pc)
	}
}
