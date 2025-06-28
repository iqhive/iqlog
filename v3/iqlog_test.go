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
	logger := iqlog.NewIQLogger(false)

	logger.SetWriter(buf)
	logger.SetDebugMode(true)

	logger.Info("Test Info")
	logger.Warn("Test Warn")
	logger.Error("Test Error")
	logger.Debug("Test Debug")
	logger.Debugf("Test Debugf %s", "test")

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
	logger := iqlog.NewIQLogger(false)

	logger.SetWriter(buf)
	logger.SetDebugMode(false)

	logger.Info("Info with DebugMode disabled")
	logger.Debug("Debug with DebugMode disabled")

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
	originalVal, _ := os.LookupEnv(envVarKey)

	// Temporarily set environment variable for this test
	os.Setenv(envVarKey, "1") // Suppose setting it to "1" forces debug on
	defer func() {
		os.Setenv(envVarKey, originalVal)
	}()

	// Force re-read env config if your logger supports it. Example:
	// iqlog.ReloadConfigFromEnv() // Hypothetical

	buf := &bytes.Buffer{}
	logger := iqlog.NewIQLogger(false)

	logger.SetWriter(buf)

	// If the logger automatically reads environment variables, we expect debug to be on now.
	logger.Debug("Environment debug check")
	// logger.Flush()
	// time.Sleep(1 * time.Second)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Environment debug check") {
		t.Errorf("Expected debug message due to environment variable, got: %s", logOutput)
	}
}

// TestSequentialLogs ensures that multiple consecutive logs don't conflict.
func TestSequentialLogs(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := iqlog.NewIQLogger(false)

	logger.SetWriter(buf)
	logger.SetDebugMode(true)

	for i := 0; i < 5; i++ {
		logger.Infof("Sequential log message: %d", i)
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
	logger := iqlog.NewIQLogger(false)

	logger.SetWriter(buf)
	logger.SetDebugMode(true)

	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 10

	var syncLock sync.RWMutex
	syncLock.Lock()

	wg.Add(numGoroutines)
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			syncLock.RLock()
			defer syncLock.RUnlock()
			for m := 0; m < messagesPerGoroutine; m++ {
				logger.Info("ConcurrentLog", goroutineID, "message", m)
			}
		}(g)
	}
	syncLock.Unlock()
	wg.Wait()

	logger.Flush()

	logOutput := buf.String()
	// Verify that we see the correct count of messages
	expectedCount := numGoroutines * messagesPerGoroutine
	count := strings.Count(logOutput, "ConcurrentLog")
	if count != expectedCount {
		t.Errorf("Expected %d log messages, found %d", expectedCount, count)
	}
}

// TestLogFormatStringInfo uses a regex to validate the format of each log line if known or enforced by the iqlog package.
func TestLogFormatStringInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := iqlog.NewIQLogger(false)

	logger.SetWriter(buf)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.SetUseColour(false)
	logger.Info("Format check")
	logOutput := buf.String()

	// Example format check: [INFO]: Format check
	regexPattern := `(?m)^INFO \w.*\n$`
	matched, err := regexp.MatchString(regexPattern, logOutput)
	if err != nil {
		t.Fatalf("Failed to compile regex pattern: %v", err)
	}
	if !matched {
		t.Errorf("Log output did not match expected format. Output:\n|%s|", logOutput)
	}
}

// TestLogFormatJSONInfo checks json output
func TestLogFormatJSONInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := iqlog.NewIQLogger(true)

	logger.SetWriter(buf)
	logger.SetDebugMode(true)

	logger.Info("Format check")
	logOutput := buf.String()

	// Example format check: [INFO]: Format check
	// re := regexp.MustCompile(`^\{"time":"[A-Za-z]{3}\s\d{1,2}\s\d{2}:\d{2}:\d{2}\.\d+","level":\d+,"message":"[^"]+"\}\n$`)
	re := regexp.MustCompile(`^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Format check"\}\n$`)
	if !re.MatchString(logOutput) {
		t.Errorf("Log output did not match expected format. Output:\n|%s|", logOutput)
	}
}

// TestLogFormatStringWarn uses a regex to validate the format of each log line if known or enforced by the iqlog package.
func TestLogFormatStringWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := iqlog.NewIQLogger(false)

	logger.SetWriter(buf)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.SetUseColour(false)
	logger.Warn("Format check")
	logOutput := buf.String()

	// Example format check: [WARN]: Format check
	regexPattern := `(?m)^WARN \w.*\n$`
	matched, err := regexp.MatchString(regexPattern, logOutput)
	if err != nil {
		t.Fatalf("Failed to compile regex pattern: %v", err)
	}
	if !matched {
		t.Errorf("Log output did not match expected format. Output:\n|%s|", logOutput)
	}
}

// TestLogFormatJSONWarn checks json output
func TestLogFormatJSONWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := iqlog.NewIQLogger(true)

	logger.SetWriter(buf)
	logger.SetDebugMode(true)

	logger.Warn("Format check")
	logOutput := buf.String()

	// Example format check: [INFO]: Format check
	// re := regexp.MustCompile(`^\{"time":"[A-Za-z]{3}\s\d{1,2}\s\d{2}:\d{2}:\d{2}\.\d+","level":\d+,"message":"[^"]+"\}\n$`)
	re := regexp.MustCompile(`^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"warn","message":"Format check"\}\n$`)
	if !re.MatchString(logOutput) {
		t.Errorf("Log output did not match expected format. Output:\n|%s|", logOutput)
	}
}
