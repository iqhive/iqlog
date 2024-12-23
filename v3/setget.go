package iqlog

import (
	"io"
	"log/syslog"
	"os"
	"strings"
)

func (l *logger) SetWriter(w io.Writer) {
	l.out = w
}

func (l *logger) GetWriter() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}

func SetWriter(w io.Writer) {
	GlobalLogger.SetWriter(w)
}

func GetWriter() io.Writer {
	return GlobalLogger.GetWriter()
}

// SetApplicationName sets the application name on the GlobalLogger
func SetApplicationName(applicationName string) {
	GlobalLogger.SetApplicationName(applicationName)
}
func (l *logger) SetApplicationName(name string) { l.applicationName = name }

// SetSyslogHost sets the syslog host on the GlobalLogger
func SetSyslogHost(host string) {
	GlobalLogger.SetSyslogHost(host)
}

// SetDebugMode sets the syslog host on the GlobalLogger
func SetDebugMode(debugMode bool) {
	GlobalLogger.SetDebugMode(debugMode)
}
func (l *logger) SetDebugMode(d bool) { l.debug = d }

// SetNewLine sets the new line on the GlobalLogger
func SetNewLine(newLine bool) {
	GlobalLogger.SetNewLine(newLine)
}
func (l *logger) SetNewLine(d bool) {
	l.newLine = d
}

// SetCaptureCallers sets the capture callers on the GlobalLogger
func SetCaptureCallers(captureCallers bool) {
	GlobalLogger.SetCaptureCallers(captureCallers)
}
func (l *logger) SetCaptureCallers(d bool) { l.captureCallers = d }

func SetUseColour(useColour bool) {
	GlobalLogger.SetUseColour(useColour)
}
func (l *logger) SetUseColour(d bool) { l.useColour = d }

// SetJSONMode sets the JSON mode on the GlobalLogger
func SetJSONMode(jsonMode bool) {
	GlobalLogger.SetJSONMode(jsonMode)
}
func (l *logger) SetJSONMode(d bool) {
	l.jsonMode = d
	l.newLine = true
	l.useColour = false
}

func (l *logger) SetSyslogHost(newhost string) {
	l.syslogHost = newhost
	if newhost != "" && !strings.Contains(newhost, ":") {
		// make sure we have a (UDP) port in the host definition
		newhost = newhost + ":514"
	}
	if l.syslogHost == newhost && newhost == "" {
		// no change
		l.Debugf("Syslog host not changed to (%s) - already set to that", newhost)
		return
	}
	if newhost == "" {
		l.Infof("Log output changed to StdErr", newhost)
		l.out = os.Stderr
		return
	}
	newSyslog, syslogErr := syslog.Dial("udp", newhost, syslog.LOG_DAEMON|syslog.LOG_INFO, l.applicationName)
	if syslogErr == nil && newSyslog != nil {
		l.out = newSyslog
	} else {
		l.Errorf("ERROR: Unable to init syslog to (%s): %v", newhost, syslogErr)
		os.Exit(1)
	}
}

func (r *LogRecord) SetField(i int, key string, val any) {
	if i < 0 || i >= len(r.Fields) {
		return
	}
	r.Fields[i].KeyLen = safeStringCopy(&r.Fields[i].Key, key)

	r.Fields[i].Quote = false

	switch v := val.(type) {
	case nil:
		r.Fields[i].VLen = safeStringCopy(&r.Fields[i].VStr, "null")
	case bool:
		if v {
			r.Fields[i].VLen = safeStringCopy(&r.Fields[i].VStr, "true")
		} else {
			r.Fields[i].VLen = safeStringCopy(&r.Fields[i].VStr, "false")
		}
	case int:
		r.Fields[i].VLen = writeIntDecimal(r.Fields[i].VStr[:], int64(v))
	case int64:
		r.Fields[i].VLen = writeIntDecimal(r.Fields[i].VStr[:], v)
	case float64:
		n := fastFloatFill(r.Fields[i].VStr[:], v, 6)
		r.Fields[i].VLen = n
	case float32:
		n := fastFloatFill(r.Fields[i].VStr[:], float64(v), 6)
		r.Fields[i].VLen = n
	case string:
		r.Fields[i].Quote = true
		r.Fields[i].VLen = safeStringCopy(&r.Fields[i].VStr, v)
	default:
		r.Fields[i].Quote = true
		r.Fields[i].VLen = safeStringCopy(&r.Fields[i].VStr, "unsupported_type")
	}

	r.Fields[i].Used = true
}
