package iqlog

import (
	"io"
	"log/syslog"
	"os"
	"strings"

	"bitbucket.org/iqhive/iqlog/v3/ringbuffer"
)

type ringWriter struct {
	ringBuffer *ringbuffer.RingBuffer[[]byte]
	writer     io.Writer
}

func (rw *ringWriter) Start() {
	go func() {
		for {
			val, ok := rw.ringBuffer.DequeueBlocking()
			if !ok {
				// Nothing available, maybe sleep or continue
				// fmt.Println("Nothing available on ring!")
				// os.Exit(1)
				continue
			}
			rw.writer.Write(val)
		}
	}()
}

func (rw *ringWriter) Write(p []byte) (n int, err error) {
	rw.ringBuffer.Enqueue(p)
	return len(p), nil
}

func (l *logger) SetWriter(w io.Writer) {
	// option 1 - plain old writer
	l.out = w
}

func (l *logger) SetAsyncWriter(w io.Writer) {
	// // option 1 - plain old writer
	// l.out = w

	// option 2 - async writer, which should be ok, but its really not
	l.out = newAsyncWriter(w, 1000)
}

func (l *logger) SetRingbufferWriter(w io.Writer) {
	// option 3 - ring writer, which should be better for non-stop loggings, lets see
	rw := &ringWriter{
		ringBuffer: ringbuffer.NewRingBuffer[[]byte](10000),
		writer:     w,
	}

	rw.Start()

	l.out = rw
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
func (l *logger) SetDebugMode(d bool) {
	if d {
		l.Level = LevelDebug
	} else {
		l.Level = LevelInfo
	}
}

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
	useColour = useColour
}
func (l *logger) SetUseColour(d bool) { useColour = d }

// SetJSONMode sets the JSON mode on the GlobalLogger
func SetJSONMode(jsonMode bool) {
	GlobalLogger.SetJSONMode(jsonMode)
}
func (l *logger) SetJSONMode(isJSONmode bool) {
	if isJSONmode {
		l.jsonMode = isJSONmode
		l.newLine = true
		l.IncludeTime = true
	} else {
		l.jsonMode = isJSONmode
		l.newLine = true
		l.IncludeTime = false
	}
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
		l.Info("Log output changed to StdErr", newhost)
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
