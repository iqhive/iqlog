package iqlog

import (
	"bytes"
	"regexp"
	"testing"
)

func TestPreallocLineConsoleMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithPreallocLineInfo().Msg("Test message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO Test message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineConsoleMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithPreallocLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO string=value int=42 float=3.140000 Test message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineConsoleMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Msg("Test message")
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

	// expected := "INFO string=value int=42 float=3.140000 Test message\n"
	// if logOutput != expected {
	// 	t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	// }
}
func TestPreallocLineConsoleMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO string=value int=42 float=3.140000 Test message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

	// expected := "INFO string=value int=42 float=3.140000 Test message\n"
	// if logOutput != expected {
	// 	t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	// }
}

func TestPreallocLineConsoleMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithPreallocLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expected := "INFO Test message 42 string 3.140000\n"
	if logOutput != expected {
		t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	}
}

func TestPreallocLineConsoleMsgfWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestPreallocLineConsoleMsgfTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestPreallocLineConsoleMsgfVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestPreallocLineJSONMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithPreallocLineInfo().Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineJSONMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithPreallocLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","string":"value","int":42,"float":3.140000,"message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineJSONMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineJSONMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","string":"value","int":42,"float":3.140000,"message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineJSONMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithPreallocLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineJSONMsgWithVarf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithPreallocLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","string":"value","int":42,"float":3.140000,"message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineJSONMsgWithTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestPreallocLineJSONMsgWithVarsTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithPreallocLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","string":"value","int":42,"float":3.140000,"message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}
