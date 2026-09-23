package iqlog

import (
	"context"
	"errors"
	"io"
	"net"
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
	// JSONTimeDisabled omits the timestamp from JSON records.
	JSONTimeDisabled JSONTimeMode = iota
	// JSONTimeUTC emits the default fixed-width UTC timestamp using the fast path.
	JSONTimeUTC
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
	// EscapeFieldNames forces JSON escaping for every dynamic field name.
	// It is not needed for correctness: keys are always scanned and escaped
	// when they contain a byte that would break the JSON string. Set it only
	// to skip that scan's branch for keys already known to need escaping.
	EscapeFieldNames bool
	// IncludeTime controls console timestamps. JSON timestamps are
	// controlled by JSONTimeMode.
	IncludeTime bool
	// TimestampLayout is passed to time.Format when JSONTimeMode is
	// JSONTimeCustom. It also controls console timestamps.
	TimestampLayout string
	// JSONTimeMode controls timestamp output for JSON records. The zero
	// value (JSONTimeDisabled) omits the timestamp; set JSONTimeUTC or
	// JSONTimeCustom to opt in.
	JSONTimeMode JSONTimeMode
	CallerDepth  int
	// Color forces ANSI colors for console output. Terminal writers are
	// detected automatically unless DisableColor is set; Config reports the
	// forced value, not the detected one, so a round-trip does not turn a
	// detected terminal into a forced colour on a later, redirected writer.
	Color bool
	// DisableColor disables ANSI colors, including automatic terminal colors.
	DisableColor    bool
	ApplicationName string
	SyslogHost      string
	// NativeLog routes output to the platform-native system log: os_log on
	// macOS (requires cgo) and the Event Log on Windows. ApplicationName
	// becomes the os_log subsystem / event source name. When set it takes
	// precedence over Writer and SyslogHost. The feature is opt-in; the
	// default path does no native-log work.
	NativeLog        bool
	ContextExtractor func(context.Context) map[string]any
	WriterMode       WriterMode
	BufferSize       int
	OverflowPolicy   OverflowPolicy
	ExitFunc         func(int)
	Now              func() time.Time
}

type loggerConfig struct {
	format          Format
	level           Level
	includeTime     bool
	timestampLayout string
	jsonTimeMode    JSONTimeMode
	callerDepth     int
	// color is the effective setting the console encoder reads: forced by
	// colorForced, or detected on the destination when nothing overrides it.
	// colorForced is what Config.Color asked for and what Config() reports.
	color            bool
	colorForced      bool
	disableColor     bool
	applicationName  string
	syslogHost       string
	nativeLog        bool
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
		cfg.IncludeTime = cfg.JSONTimeMode != JSONTimeDisabled
	}
	cfg.Level = clampLevel(cfg.Level)
	if cfg.Writer == nil {
		cfg.Writer = os.Stderr
	}
	// The convenience names are resolved for every mode; a real Go layout
	// passes through normalizeTimestampLayout unchanged. Only the default is
	// mode-specific: a custom JSON timestamp has no default layout (validation
	// requires one), while the console default is the fixed-width fast path.
	if cfg.TimestampLayout != "" {
		cfg.TimestampLayout = normalizeTimestampLayout(cfg.TimestampLayout)
	} else if !(cfg.Format == FormatJSON && cfg.JSONTimeMode == JSONTimeCustom) {
		cfg.TimestampLayout = defaultConsoleTimestampLayout
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
	return w == nil || isNilPointer(w)
}

// isNilPointer reports whether v is a non-nil interface holding a nil
// pointer. Calling a method through such a value dereferences nil inside the
// method, so the encoders treat it as the nil it is. The reflect calls do not
// allocate for a pointer value, and every caller checks for an untyped nil
// first so the common path pays nothing.
func isNilPointer(v any) bool {
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Pointer && rv.IsNil()
}

// effectiveColor reports whether console records should carry ANSI colour
// when dest is the destination: DisableColor wins, a forced Color is
// honoured, and otherwise a terminal destination is detected. Detection is
// only meaningful for the writer actually in use, so callers pass the syslog
// connection or native handle rather than Config.Writer when those are
// selected; neither is a terminal, so the record stays plain.
func (cfg *loggerConfig) effectiveColor(dest io.Writer) bool {
	if cfg.disableColor {
		return false
	}
	return cfg.colorForced || (cfg.format == FormatConsole && terminalWriter(dest))
}

// updateConfig applies update to a fresh copy of the configuration snapshot
// and publishes it. The level fast path is republished from the same snapshot
// under cfgMu so a concurrent updateConfig or setConfig cannot leave
// Logger.level and loggerConfig.level disagreeing.
func (l *Logger) updateConfig(update func(*loggerConfig)) {
	l.cfgMu.Lock()
	next := *l.config.Load()
	update(&next)
	l.config.Store(&next)
	l.level.Store(int32(next.level))
	l.cfgMu.Unlock()
}

func (l *Logger) setConfig(input Config) error {
	if err := validateConfig(input); err != nil {
		return err
	}
	if input.Writer != nil && nilWriter(input.Writer) {
		return errNilWriter
	}
	cfg := normalizeConfig(input)
	// A queue wrapper handed back in through GetWriter is unwrapped: it is
	// about to be retired below, and the mode requested by cfg decides whether
	// the destination underneath gets a fresh queue.
	configured := unwrapAsync(cfg.Writer)
	if cfg.SyslogHost != "" && !cfg.NativeLog {
		var err error
		configured, err = prepareSyslog(cfg.SyslogHost, cfg.ApplicationName)
		if err != nil {
			return err
		}
	}
	var native nativeLogWriter
	if cfg.NativeLog {
		var err error
		native, err = prepareNativeLog(cfg)
		if err != nil {
			return err
		}
		configured = native
	}
	if nilWriter(configured) {
		return errNilWriter
	}
	active := configured
	switch cfg.WriterMode {
	case WriterAsync:
		active = newAsyncWriter(configured, cfg.BufferSize, cfg.OverflowPolicy, l.queueError, &l.writer.writeMu)
	case WriterRing:
		active = newAsyncWriter(configured, cfg.BufferSize, cfg.OverflowPolicy, l.queueError, &l.writer.writeMu)
	}
	snapshot := cfg.snapshot()
	// Colour is detected on the destination actually written to, so a
	// process started from a terminal does not colour its syslog datagrams.
	snapshot.color = snapshot.effectiveColor(configured)
	l.writer.mu.Lock()
	if l.writer.closed.Load() {
		l.writer.mu.Unlock()
		_ = retireWriter(active, nil)
		return ErrClosed
	}
	old := l.writer.active.Load()
	concurrent := cfg.ConcurrentWriter || configured == io.Discard || cfg.WriterMode != WriterSync
	l.writer.active.Store(&outputState{out: active, configured: configured, concurrent: concurrent, native: native})
	l.cfgMu.Lock()
	l.config.Store(snapshot)
	l.level.Store(int32(cfg.Level))
	l.cfgMu.Unlock()
	// The previous writer is drained while writer.mu is still held, so a
	// concurrent Flush observes the new writer only after the records accepted
	// by the old one have been written.
	err := retireWriter(old.out, configured)
	l.writer.mu.Unlock()
	return err
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
		IncludeTime:      cfg.includeTime, TimestampLayout: cfg.timestampLayout, JSONTimeMode: cfg.jsonTimeMode,
		CallerDepth: cfg.callerDepth, Color: cfg.colorForced, DisableColor: cfg.disableColor,
		ApplicationName: cfg.applicationName, SyslogHost: cfg.syslogHost, NativeLog: cfg.nativeLog,
		ContextExtractor: cfg.contextExtractor, ExitFunc: cfg.exitFunc, Now: cfg.now,
		WriterMode: cfg.writerMode, BufferSize: cfg.bufferSize, OverflowPolicy: cfg.overflowPolicy,
	}
}

// SetConfig atomically replaces formatting configuration and then replaces the writer.
func (l *Logger) SetConfig(cfg Config) error { return l.setConfig(cfg) }

// syslogAddr gives a bare host the default syslog port. Bare IPv6 literals
// contain colons themselves, so the check is a real host:port split rather
// than a search for ':'.
func syslogAddr(host string) string {
	if host == "" {
		return host
	}
	if _, _, err := net.SplitHostPort(host); err == nil {
		return host
	}
	return net.JoinHostPort(strings.Trim(host, "[]"), "514")
}

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
		color:            cfg.Color && !cfg.DisableColor,
		colorForced:      cfg.Color,
		disableColor:     cfg.DisableColor,
		applicationName:  cfg.ApplicationName,
		syslogHost:       cfg.SyslogHost,
		nativeLog:        cfg.NativeLog,
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
		start := len(dst)
		dst = now.AppendFormat(dst, cfg.timestampLayout)
		// the layout is caller-supplied and lands inside a JSON string
		dst = escapeJSONTail(dst, start)
		return append(dst, `",`...)
	}
	if cfg.timestampLayout == defaultConsoleTimestampLayout {
		start := len(dst)
		dst = append(dst, make([]byte, 29)...)
		if addTimeConsoleInPlaceCopy(now, dst[start:]) == 0 {
			return appendConsoleTimestampSlow(dst[:start], now)
		}
		return dst
	}
	dst = append(dst, '[')
	start := len(dst)
	dst = now.AppendFormat(dst, cfg.timestampLayout)
	// the layout is caller-supplied and lands in the console envelope, which
	// the whole-line scan does not cover
	dst = escapeConsoleTail(dst, start)
	return append(dst, ']', ' ')
}

func appendDefaultJSONTimestamp(dst []byte, now time.Time) []byte {
	start := len(dst)
	dst = append(dst, make([]byte, 37)...)
	if addTimeJSONInPlaceCopy(now, dst[start:]) == 0 {
		return appendJSONTimestampSlow(dst[:start], now)
	}
	dst = dst[:len(dst)-2]
	return append(dst, 'Z', '"', ',')
}

// defaultJSONTimestampLayout is what the fixed-width JSON fast path emits,
// for the fallback that handles years outside 0-9999.
const defaultJSONTimestampLayout = "2006-01-02T15:04:05.000000Z07:00"

// appendConsoleTimestampSlow is the default console timestamp for a year the
// fixed-width encoder cannot represent: the same layout, through the general
// formatter, which widens or signs the year as needed.
func appendConsoleTimestampSlow(dst []byte, now time.Time) []byte {
	dst = append(dst, '[')
	dst = now.AppendFormat(dst, defaultConsoleTimestampLayout)
	return append(dst, ']', ' ')
}

// appendJSONTimestampSlow is the JSONTimeUTC timestamp for a year the
// fixed-width encoder cannot represent. now is already in UTC, so the zone
// suffix renders as Z.
func appendJSONTimestampSlow(dst []byte, now time.Time) []byte {
	dst = append(dst, `{"time":"`...)
	dst = now.AppendFormat(dst, defaultJSONTimestampLayout)
	return append(dst, '"', ',')
}
