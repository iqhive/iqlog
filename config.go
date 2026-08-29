package iqlog

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

var errNilWriter = errors.New("iqlog: nil writer")

const (
	defaultConsoleTimestampLayout = "2006-01-02T15:04:05.000000"
)

var (
	ErrClosed       = errors.New("iqlog: logger closed")
	ErrWriteDropped = errors.New("iqlog: write dropped because buffer is full")
)

type WriterMode uint8

const (
	WriterSync WriterMode = iota
	WriterAsync
	WriterRing
)

type OverflowPolicy uint8

const (
	OverflowBlock OverflowPolicy = iota
	OverflowDrop
	OverflowSync
)

// JSONTimeMode controls timestamp output for JSON records.
type JSONTimeMode uint8

const (
	// JSONTimeUTC emits the default fixed-width UTC timestamp using the fast path.
	JSONTimeUTC JSONTimeMode = iota
	// JSONTimeDisabled omits the timestamp from JSON records.
	JSONTimeDisabled
	// JSONTimeCustom formats JSON timestamps with Config.TimestampLayout.
	JSONTimeCustom
)

// Format selects the record encoding.
type Format uint8

const (
	FormatConsole Format = iota
	FormatJSON
)

// Config configures a Logger. The zero value uses synchronous console output
// to stderr at info level.
type Config struct {
	Format Format
	Level  Level
	Writer io.Writer
	// ConcurrentWriter skips serialization when Writer supports concurrent writes.
	ConcurrentWriter bool
	// EscapeFieldNames enables JSON escaping for dynamic field names. It is
	// optional because trusted identifier-style keys are substantially faster.
	EscapeFieldNames bool
	IncludeTime      bool
	// DisableTime disables timestamps when Format is FormatJSON. JSON output
	// includes timestamps by default.
	DisableTime     bool
	// TimestampLayout is passed to time.Format when JSONTimeMode is
	// JSONTimeCustom. It also controls console timestamps.
	TimestampLayout string
	// JSONTimeMode selects UTC, disabled, or custom JSON timestamps. The zero
	// value is the high-performance UTC mode.
	JSONTimeMode    JSONTimeMode
	CallerDepth     int
	// Color forces ANSI colors for console output. Terminal writers are
	// detected automatically unless DisableColor is set.
	Color bool
	// DisableColor disables ANSI colors, including automatic terminal colors.
	DisableColor     bool
	ApplicationName  string
	SyslogHost       string
	ContextExtractor func(context.Context) map[string]any
	WriterMode       WriterMode
	BufferSize       int
	OverflowPolicy   OverflowPolicy
	ExitFunc         func(int)
	Now              func() time.Time
}

type loggerConfig struct {
	format           Format
	level            Level
	includeTime      bool
	timestampLayout  string
	jsonTimeMode     JSONTimeMode
	callerDepth      int
	color            bool
	applicationName  string
	syslogHost       string
	contextExtractor func(context.Context) map[string]any
	exitFunc         func(int)
	now              func() time.Time
	writerMode       WriterMode
	bufferSize       int
	overflowPolicy   OverflowPolicy
	concurrentWriter bool
	escapeFieldNames bool
}

func normalizeConfig(cfg Config) Config {
	if cfg.Format == FormatJSON {
		if cfg.DisableTime {
			cfg.JSONTimeMode = JSONTimeDisabled
		}
		cfg.IncludeTime = cfg.JSONTimeMode != JSONTimeDisabled
	}
	if cfg.Format == FormatConsole && !cfg.DisableColor && terminalWriter(cfg.Writer) {
		cfg.Color = true
	}
	if cfg.DisableColor {
		cfg.Color = false
	}
	if cfg.Level == LevelUnknown {
		cfg.Level = LevelInfo
	}
	if cfg.Writer == nil {
		cfg.Writer = os.Stderr
	}
	if cfg.Format == FormatJSON && cfg.JSONTimeMode == JSONTimeCustom {
		// Custom layouts are passed directly to time.Time.AppendFormat.
	} else if cfg.TimestampLayout == "" {
		cfg.TimestampLayout = defaultConsoleTimestampLayout
	} else {
		cfg.TimestampLayout = normalizeTimestampLayout(cfg.TimestampLayout)
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 1000
	}
	if cfg.WriterMode == WriterRing && cfg.OverflowPolicy == OverflowBlock {
		cfg.OverflowPolicy = OverflowSync
	}
	if cfg.ExitFunc == nil {
		cfg.ExitFunc = os.Exit
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return cfg
}

func validateConfig(cfg Config) error {
	if cfg.Format != FormatConsole && cfg.Format != FormatJSON {
		return errors.New("iqlog: invalid format")
	}
	if cfg.CallerDepth < 0 {
		return errors.New("iqlog: caller depth must not be negative")
	}
	if cfg.BufferSize < 0 {
		return errors.New("iqlog: buffer size must not be negative")
	}
	if cfg.WriterMode > WriterRing {
		return errors.New("iqlog: invalid writer mode")
	}
	if cfg.OverflowPolicy > OverflowSync {
		return errors.New("iqlog: invalid overflow policy")
	}
	if cfg.JSONTimeMode > JSONTimeCustom {
		return errors.New("iqlog: invalid JSON time mode")
	}
	if cfg.Format == FormatJSON && cfg.JSONTimeMode == JSONTimeCustom && cfg.TimestampLayout == "" {
		return errors.New("iqlog: custom JSON time mode requires a timestamp layout")
	}
	return nil
}

func nilWriter(w io.Writer) bool {
	if w == nil {
		return true
	}
	v := reflect.ValueOf(w)
	return v.Kind() == reflect.Pointer && v.IsNil()
}

func (l *Logger) updateConfig(update func(*loggerConfig)) {
	for {
		old := l.config.Load()
		next := *old
		update(&next)
		if l.config.CompareAndSwap(old, &next) {
			return
		}
	}
}

func (l *Logger) setConfig(input Config) error {
	if err := validateConfig(input); err != nil {
		return err
	}
	if input.Writer != nil && nilWriter(input.Writer) {
		return errNilWriter
	}
	cfg := normalizeConfig(input)
	configured := cfg.Writer
	if cfg.SyslogHost != "" {
		var err error
		configured, err = prepareSyslog(cfg.SyslogHost, cfg.ApplicationName)
		if err != nil {
			return err
		}
	}
	if nilWriter(configured) {
		return errNilWriter
	}
	active := configured
	switch cfg.WriterMode {
	case WriterAsync:
		active = newAsyncWriter(configured, cfg.BufferSize, cfg.OverflowPolicy, l.queueError)
	case WriterRing:
		active = newAsyncWriter(configured, cfg.BufferSize, cfg.OverflowPolicy, l.queueError)
	}
	l.writer.mu.Lock()
	if l.writer.closed.Load() {
		l.writer.mu.Unlock()
		if async, ok := active.(*asyncWriter); ok {
			_ = async.Close()
		}
		return ErrClosed
	}
	old := l.writer.active.Load()
	concurrent := cfg.ConcurrentWriter || configured == io.Discard || cfg.WriterMode != WriterSync
	l.writer.active.Store(&outputState{out: active, configured: configured, concurrent: concurrent})
	l.config.Store(cfg.snapshot())
	l.level.Store(int32(cfg.Level))
	l.writer.mu.Unlock()
	if async, ok := old.out.(*asyncWriter); ok {
		return async.Close()
	}
	return nil
}

func (l *Logger) queueError(err error) {
	if errors.Is(err, ErrWriteDropped) {
		l.writer.dropped.Add(1)
	}
	l.recordWriteErr(err)
}

// Config returns a coherent copy of the logger's current configuration.
func (l *Logger) Config() Config {
	cfg := l.config.Load()
	w := l.writer.active.Load().configured
	return Config{
		Format: cfg.format, Level: cfg.level, Writer: w,
		ConcurrentWriter: cfg.concurrentWriter,
		EscapeFieldNames: cfg.escapeFieldNames,
		IncludeTime:      cfg.includeTime, DisableTime: cfg.format == FormatJSON && cfg.jsonTimeMode == JSONTimeDisabled, TimestampLayout: cfg.timestampLayout, JSONTimeMode: cfg.jsonTimeMode,
		CallerDepth: cfg.callerDepth, Color: cfg.color, DisableColor: !cfg.color,
		ApplicationName: cfg.applicationName, SyslogHost: cfg.syslogHost,
		ContextExtractor: cfg.contextExtractor, ExitFunc: cfg.exitFunc, Now: cfg.now,
		WriterMode: cfg.writerMode, BufferSize: cfg.bufferSize, OverflowPolicy: cfg.overflowPolicy,
	}
}

// SetConfig atomically replaces formatting configuration and then replaces the writer.
func (l *Logger) SetConfig(cfg Config) error { return l.setConfig(cfg) }

func normalizeTimestampLayout(layout string) string {
	switch strings.ToLower(layout) {
	case "seconds":
		return "2006-01-02T15:04:05Z07:00"
	case "milliseconds":
		return "2006-01-02T15:04:05.000Z07:00"
	case "microseconds":
		return "2006-01-02T15:04:05.000000Z07:00"
	case "nanoseconds":
		return time.RFC3339Nano
	default:
		return layout
	}
}

func (cfg Config) snapshot() *loggerConfig {
	return &loggerConfig{
		format:           cfg.Format,
		level:            cfg.Level,
		includeTime:      cfg.IncludeTime,
		timestampLayout:  cfg.TimestampLayout,
		jsonTimeMode:     cfg.JSONTimeMode,
		callerDepth:      cfg.CallerDepth,
		color:            cfg.Color,
		applicationName:  cfg.ApplicationName,
		syslogHost:       cfg.SyslogHost,
		contextExtractor: cfg.ContextExtractor,
		exitFunc:         cfg.ExitFunc,
		now:              cfg.Now,
		writerMode:       cfg.WriterMode,
		bufferSize:       cfg.BufferSize,
		overflowPolicy:   cfg.OverflowPolicy,
		concurrentWriter: cfg.ConcurrentWriter,
		escapeFieldNames: cfg.EscapeFieldNames,
	}
}

func appendTimestamp(dst []byte, now time.Time, cfg *loggerConfig, jsonMode bool) []byte {
	if jsonMode {
		if cfg.jsonTimeMode == JSONTimeUTC {
			return appendDefaultJSONTimestamp(dst, now.UTC())
		}
		dst = append(dst, `{"time":"`...)
		dst = now.AppendFormat(dst, cfg.timestampLayout)
		return append(dst, `",`...)
	}
	if cfg.timestampLayout == defaultConsoleTimestampLayout {
		dst = append(dst, make([]byte, 29)...)
		addTimeConsoleInPlaceCopy(now, dst[len(dst)-29:])
		return dst
	}
	dst = append(dst, '[')
	dst = now.AppendFormat(dst, cfg.timestampLayout)
	return append(dst, ']', ' ')
}

func appendDefaultJSONTimestamp(dst []byte, now time.Time) []byte {
	start := len(dst)
	dst = append(dst, make([]byte, 37)...)
	addTimeJSONInPlaceCopy(now, dst[start:])
	dst = dst[:len(dst)-2]
	return append(dst, 'Z', '"', ',')
}
