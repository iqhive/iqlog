package iqlog

import (
	"bytes"
	"regexp"
	"testing"
)

func TestBufferLineConsoleMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithBufferLineInfo().Msg("Test message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO Test message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineConsoleMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithBufferLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO string=value int=42 float=3.140000 Test message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineConsoleMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Msg("Test message")
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
func TestBufferLineConsoleMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
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

func TestBufferLineConsoleMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithBufferLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expected := "INFO Test message 42 string 3.140000\n"
	if logOutput != expected {
		t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	}
}

func TestBufferLineConsoleMsgfWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestBufferLineConsoleMsgfTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestBufferLineConsoleMsgfVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestBufferLineJSONMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithBufferLineInfo().Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineJSONMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithBufferLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","string":"value","int":42,"float":3.140000,"message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineJSONMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineJSONMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","string":"value","int":42,"float":3.140000,"message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineJSONMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithBufferLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineJSONMsgWithVarf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithBufferLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","string":"value","int":42,"float":3.140000,"message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineJSONMsgWithTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestBufferLineJSONMsgWithVarsTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithBufferLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","string":"value","int":42,"float":3.140000,"message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}
