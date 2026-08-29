package iqlog

import (
	"bytes"
	"regexp"
	"testing"
)

func TestByteSliceLineConsoleMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(false)
	logger.InfoEvent().Msg("Test 1 message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO Test 1 message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineConsoleMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(false)
	logger.InfoEvent().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test 2 message")
	logOutput := buf.String()

	expectedRegexStr := `^INFO string=value int=42 float=3.14 Test 2 message\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineConsoleMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Msg("Test 3 message")
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
func TestByteSliceLineConsoleMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test 4 message")
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO string=value int=42 float=3.14 Test 4 message\n$`
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
	logger := MustNew(Config{IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(false)
	logger.InfoEvent().Msgf("Test 5 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expected := "INFO Test 5 message 42 string 3.140000\n"
	if logOutput != expected {
		t.Errorf("Expected |%q|, got |%q|", expected, logOutput)
	}
}

func TestByteSliceLineConsoleMsgfWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Msgf("Test 6 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test 6 message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestByteSliceLineConsoleMsgfTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Msgf("Test 7 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test 7 message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestByteSliceLineConsoleMsgfVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Msgf("Test 8 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()

	expectedRegexStr := `^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\] INFO Test 8 message 42 string 3.140000\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}

}

func TestByteSliceLineJSONMsg(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{Format: FormatJSON, IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(false)
	logger.InfoEvent().Msg("Test 9 message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"INFO","message":"Test 9 message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithVars(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{Format: FormatJSON, IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(false)
	logger.InfoEvent().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test 10 message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"INFO","string":"value","int":42,"float":3.14,"message":"Test 10 message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{Format: FormatJSON, IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Msg("Test 11 message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z","level":"INFO","message":"Test 11 message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithVarsTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{Format: FormatJSON, IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msg("Test 12 message")
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z","level":"INFO","string":"value","int":42,"float":3.14,"message":"Test 12 message"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{Format: FormatJSON, IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(false)
	logger.InfoEvent().Msgf("Test 13 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"INFO","message":"Test 13 message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithVarf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{Format: FormatJSON, IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(false)
	logger.InfoEvent().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test 14 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"level":"INFO","string":"value","int":42,"float":3.14,"message":"Test 14 message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{Format: FormatJSON, IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Msgf("Test 15 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z","level":"INFO","message":"Test 15 message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}

func TestByteSliceLineJSONMsgWithVarsTimestampf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := MustNew(Config{Format: FormatJSON, IncludeTime: true})
	logger.setWriter(buf)
	logger.setLevel(LevelDebug)
	logger.setIncludeTime(true)
	logger.InfoEvent().Str("string", "value").Int("int", 42).Float32("float", 3.14).Msgf("Test 16 message %d %s %f", 42, "string", 3.14)
	logOutput := buf.String()
	expectedRegexStr := `^\{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z","level":"INFO","string":"value","int":42,"float":3.14,"message":"Test 16 message 42 string 3.140000"\}\n$`
	expected := regexp.MustCompile(expectedRegexStr)
	if !expected.MatchString(logOutput) {
		t.Errorf("Expected |%q|, got |%q|", expectedRegexStr, logOutput)
	}
}
