package iqlog

import (
	"bytes"
	"regexp"
	"testing"
)

func TestVarStackConsoleMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithVarStackInfo().Msg("Test 1 message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO Test 1 message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackConsoleMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithVarStackInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test 2 message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO string=value int=42 float=3.140000 Test 2 message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackConsoleMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Msg("Test 3 message")
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test 3 message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

	// expected := "INFO string=value int=42 float=3.140000 Test message\n"
	// if logOutput != expected {
	// 	t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	// }
}
func TestVarStackConsoleMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test 4 message")
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO string=value int=42 float=3.140000 Test 4 message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

	// expected := "INFO string=value int=42 float=3.140000 Test message\n"
	// if logOutput != expected {
	// 	t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	// }
}

func TestVarStackConsoleMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithVarStackInfo().Msgf("Test 5 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expected := "INFO Test 5 message 42 string 3.140000\n"
	if logOutput != expected {
		t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	}
}

func TestVarStackConsoleMsgfWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Msgf("Test 6 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test 6 message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestVarStackConsoleMsgfTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Msgf("Test 7 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test 7 message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestVarStackConsoleMsgfVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(false)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Msgf("Test 8 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test 8 message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestVarStackJSONMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithVarStackInfo().Msg("Test 9 message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","message":"Test 9 message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackJSONMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithVarStackInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test 10 message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","string":"value","int":42,"float":3.140000,"message":"Test 10 message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackJSONMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Msg("Test 11 message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Test 11 message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackJSONMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test 12 message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","string":"value","int":42,"float":3.140000,"message":"Test 12 message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackJSONMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithVarStackInfo().Msgf("Test 13 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","message":"Test 13 message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackJSONMsgWithVarf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(false)
	logger.WithVarStackInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test 14 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"info","string":"value","int":42,"float":3.140000,"message":"Test 14 message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackJSONMsgWithTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Msgf("Test 15 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","message":"Test 15 message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestVarStackJSONMsgWithVarsTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewIQLogger(true)
	logger.SetWriter(buf)
	logger.SetDebugMode(true)
	logger.IncludeTime.Store(true)
	logger.WithVarStackInfo().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test 16 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}","level":"info","string":"value","int":42,"float":3.140000,"message":"Test 16 message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}
