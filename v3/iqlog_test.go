package iqlog_test

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"bitbucket.org/iqhive/iqlog/v3"
)

// TestBasicLogs ensures that each log level can produce output without error.
func TestBasicLogs(t *testing.T) {
	buf := &bytes.Buffer{}

	// Replace the internal writer for capturing logs (if your iqlog package supports it)
	originalWriter := iqlog.GetWriter() // Hypothetical method to get the current writer.
	defer iqlog.SetWriter(originalWriter)

	iqlog.SetWriter(buf)
	iqlog.SetDebugMode(true)

	iqlog.Info("Test Info")
	iqlog.Warn("Test Warn")
	iqlog.Error("Test Error")
	iqlog.Debug("Test Debug")
	iqlog.Debugf("Test Debugf %s", "test")

	logOutput := buf.String()

	if !strings.Contains(logOutput, "Test Info") {
		t.Errorf("Expected 'Test Info' in log output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "Test Warn") {
		t.Errorf("Expected 'Test Warn' in log output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "Test Error") {
		t.Errorf("Expected 'Test Error' in log output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "Test Debug") {
		t.Errorf("Expected 'Test Debug' in log output, got: %s", logOutput)
	}
}

// TestDebugModeOff ensures that Debug logs are suppressed when debug mode is disabled.
func TestDebugModeOff(t *testing.T) {
	buf := &bytes.Buffer{}
	originalWriter := iqlog.GetWriter()
	defer iqlog.SetWriter(originalWriter)

	iqlog.SetWriter(buf)
	iqlog.SetDebugMode(false)

	iqlog.Info("Info with DebugMode disabled")
	iqlog.Debug("Debug with DebugMode disabled")

	logOutput := buf.String()

	if !strings.Contains(logOutput, "Info with DebugMode disabled") {
		t.Errorf("Expected 'Info with DebugMode disabled' in log output, got: %s", logOutput)
	}
	if strings.Contains(logOutput, "Debug with DebugMode disabled") {
		t.Errorf("Did not expect 'Debug with DebugMode disabled' in log output when debug mode is off")
	}
}

// TestEnvironmentVariable checks if the logging behaves differently based on an environment variable (if supported).
func TestEnvironmentVariable(t *testing.T) {
	// Hypothetical example: if iqlog can read an env var "IQLOG_DEBUG" to auto-enable debug
	const envVarKey = "IQLOG_DEBUG"
	originalVal, hadVal := os.LookupEnv(envVarKey)

	// Temporarily set environment variable for this test
	os.Setenv(envVarKey, "1") // Suppose setting it to "1" forces debug on
	defer func() {
		if hadVal {
			os.Setenv(envVarKey, originalVal)
		} else {
			os.Unsetenv(envVarKey)
		}
	}()

	// Force re-read env config if your logger supports it. Example:
	// iqlog.ReloadConfigFromEnv() // Hypothetical

	buf := &bytes.Buffer{}
	originalWriter := iqlog.GetWriter()
	defer iqlog.SetWriter(originalWriter)

	iqlog.SetWriter(buf)

	// If the logger automatically reads environment variables, we expect debug to be on now.
	iqlog.Debug("Environment debug check")

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Environment debug check") {
		t.Errorf("Expected debug message due to environment variable, got: %s", logOutput)
	}
}

// TestSequentialLogs ensures that multiple consecutive logs don't conflict.
func TestSequentialLogs(t *testing.T) {
	buf := &bytes.Buffer{}
	originalWriter := iqlog.GetWriter()
	defer iqlog.SetWriter(originalWriter)

	iqlog.SetWriter(buf)
	iqlog.SetDebugMode(true)

	for i := 0; i < 5; i++ {
		iqlog.Info("Sequential log message:", i)
	}

	logOutput := buf.String()

	// We expect 5 lines containing the message
	count := strings.Count(logOutput, "Sequential log message:")
	if count != 5 {
		t.Errorf("Expected 5 sequential log messages, got %d", count)
	}
}

// TestConcurrentLogs checks if concurrent logging behaves as expected without races or corruption.
func TestConcurrentLogs(t *testing.T) {
	buf := &bytes.Buffer{}
	originalWriter := iqlog.GetWriter()
	defer iqlog.SetWriter(originalWriter)

	iqlog.SetWriter(buf)
	iqlog.SetDebugMode(true)

	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 10

	wg.Add(numGoroutines)
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			for m := 0; m < messagesPerGoroutine; m++ {
				iqlog.Info("ConcurrentLog", goroutineID, "message", m)
			}
		}(g)
	}
	wg.Wait()

	logOutput := buf.String()
	// Verify that we see the correct count of messages
	expectedCount := numGoroutines * messagesPerGoroutine
	count := strings.Count(logOutput, "ConcurrentLog")
	if count != expectedCount {
		t.Errorf("Expected %d log messages, found %d", expectedCount, count)
	}
}

// TestLogFormat uses a regex to validate the format of each log line if known or enforced by the iqlog package.
func TestLogFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	originalWriter := iqlog.GetWriter()
	defer iqlog.SetWriter(originalWriter)

	iqlog.SetWriter(buf)
	iqlog.SetDebugMode(true)

	iqlog.Info("Format check")
	logOutput := buf.String()

	// Example format check: [INFO]: Format check
	regexPattern := `(?m)^\[INFO\]:\s.*$`
	matched, err := regexp.MatchString(regexPattern, logOutput)
	if err != nil {
		t.Fatalf("Failed to compile regex pattern: %v", err)
	}
	if !matched {
		t.Errorf("Log output did not match expected format. Output:\n%s", logOutput)
	}
}
