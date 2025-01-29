package iqlog

import (
	"bytes"
	"regexp"
	"testing"
)

func TestByteSliceLineConsoleMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = false
	logger.WithByteSliceLineInfo().Msg("Test message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO Test message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineConsoleMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = false
	logger.WithByteSliceLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO string=value int=42 float=3.140000 Test message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineConsoleMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Msg("Test message")
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
func TestByteSliceLineConsoleMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
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

func TestByteSliceLineConsoleMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = false
	logger.WithByteSliceLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expected := "INFO Test message 42 string 3.140000\n"
	if logOutput != expected {
		t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	}
}

func TestByteSliceLineConsoleMsgfWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestByteSliceLineConsoleMsgfTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestByteSliceLineConsoleMsgfVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestByteSliceLineJSONMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = false
	logger.WithByteSliceLineInfo().Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = false
	logger.WithByteSliceLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","string":"value","int":42,"float":3.140000,"message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","string":"value","int":42,"float":3.140000,"message":"Test message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = false
	logger.WithByteSliceLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithVarf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = false
	logger.WithByteSliceLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","string":"value","int":42,"float":3.140000,"message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithVarsTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime = true
	logger.WithByteSliceLineInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","string":"value","int":42,"float":3.140000,"message":"Test message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}
