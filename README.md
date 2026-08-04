# iqlog

A fast, low-allocation structured logging library for Go, built around pooled
line builders and hand-rolled formatting to minimise garbage-collector
pressure on hot logging paths.

```go
import "github.com/iqhive/iqlog"
```

Requires Go 1.23+.

## Quick start

```go
package main

import "github.com/iqhive/iqlog"

func main() {
    // One-call setup of the global logger: application name, optional
    // syslog host ("" = log to stderr), and debug mode.
    iqlog.Init("myapp", "", false)

    iqlog.Info("service started")
    iqlog.Infof("listening on %s", ":8080")

    // Structured fields via the fluent builder API
    iqlog.InfoWith().
        Str("user", "alice").
        Int("requests", 42).
        Bool("admin", true).
        Msg("request handled")
}
```

Instance loggers can be created independently of the global logger:

```go
log := iqlog.NewIQLogger(true) // true = JSON output mode
log.Info("hello")
log.WithByteSliceLineInfo().Str("k", "v").Msg("structured")
```

## Features

### Log levels

Levels, in increasing severity: `LevelTrace`, `LevelDebug`,
`LevelInfo`/`LevelPrint`, `LevelWarn`, `LevelError`, `LevelPanic`,
`LevelFatal`. Messages below the logger's configured `Level` are discarded
cheaply (a shared no-op builder is returned, so filtered lines allocate
nothing).

Global helpers exist for every level: `Trace`, `Debug`, `Info`, `Warn`,
`Error`, `Panic`, `Fatal`, each with an `...f` formatting variant
(`Infof`, `Errorf`, ...) and a fluent `...With` variant (`InfoWith`,
`ErrorWith`, ...). The same methods are available on logger instances.

Note: `Fatal`/`Panic` via the fluent `getWith` path log at the respective
level but do not terminate the process. The dedicated per-builder
`...Fatal`/`...Panic` helpers (and their `With...` variants) call
`os.Exit(1)` after the message has been written.

The level can be changed at any time with `SetLevel(Level)` and read with
the `Level()` method; level, JSON mode, timestamp inclusion, newline mode,
and caller depth are stored atomically, so they can be reconfigured safely
while other goroutines are logging.

### Output formats

- **Console mode** (default): `LEVEL key=value ... message`, with optional
  ANSI colour for the level prefix and caller info. Colour is enabled
  automatically when stderr is a terminal (never on Windows) and can be
  overridden with `SetUseColour(bool)`. Carriage returns and newlines
  embedded in messages or values are escaped to `\r`/`\n`, so untrusted
  input cannot forge additional console log lines.
- **JSON mode**: one JSON object per line, e.g.
  `{"time":"...","level":"info","user":"alice","message":"request handled"}`.
  Enabled via `NewIQLogger(true)` or `SetJSONMode(true)`. All string field
  names, string values, and messages are JSON-escaped, so untrusted input
  cannot corrupt or inject log records.

### Structured fields

The fluent builders support the following field types: `Str`, `Int`,
`Int64`, `Float32`, `Float64`, `Bool`, and `Any` (which renders the value
with `%v` and logs it as a string). Finish a line with:

- `Msg(msg)` — plain message
- `Msgs(msg, args...)` — message followed by space-separated arguments
- `Msgf(format, args...)` — `fmt.Sprintf`-style formatting

### Timestamps

Timestamps are on by default (`IncludeTime`) and are encoded with
allocation-free hand-written formatters. `TimestampFormat` supports RFC3339
with seconds, milliseconds (default), microseconds, or nanoseconds.

### Caller capture

With `SetCallerDepth(n)` (n > 0), each line includes the calling function
and file:line, resolved through an optimised runtime-based lookup with a
PC cache (much cheaper than `runtime.Caller`). The main module path prefix
is trimmed from function names for readability. The global logger created
by `Init`/`NewGlobalIQLogger` enables caller capture at depth 1. All line
builders (including `preallocLine`, `preallocLine2`, and `bufferLineNL`)
emit `func` and `file` fields when caller capture is enabled.

### Output writers

- `SetWriter(w io.Writer)` — synchronous writes (default: stderr).
  Writes are serialised with an internal mutex.
- `SetAsyncWriter(w io.Writer)` — writes go through a buffered channel and
  a background goroutine; `Flush()` blocks until everything queued so far
  has been written (safe to call while other goroutines are logging).
- `SetRingbufferWriter(w io.Writer)` — writes are enqueued on a lock-free
  MPMC ring buffer (capacity 10,000) consumed by a background goroutine.
  If the ring is full the write falls back to a synchronous write, so no
  log lines are dropped.

Replacing the writer (via any `Set*Writer` call) closes the previous
async/ring wrapper: queued lines are drained to its underlying writer and
its background goroutine exits, so writers can be swapped repeatedly
without leaking goroutines.

Write errors from the output writer are recorded and can be inspected with
`LastWriteError()` (also available as a method on logger instances).

### Syslog (non-Windows)

`SetSyslogHost("host[:port]")` switches output to UDP syslog
(`LOG_DAEMON`); the port defaults to 514. Passing `""` switches back to
stderr. If the syslog connection cannot be established, an error is logged
and the current writer is kept. On Windows, `SetSyslogHost` is a no-op.

### Global configuration

- `Init(applicationName, syslogHost, debugMode)` — one-call setup
- `SetApplicationName(string)` — used as the syslog tag
- `SetDebugMode(bool)` — toggles between `LevelDebug` and `LevelInfo`
- `SetJSONMode(bool)`, `SetNewLine(bool)`, `SetUseColour(bool)`
- `SetLevel(Level)` — set the minimum level directly
- `SetCallerDepth(int)`, `SetWriter(io.Writer)`, `GetWriter()`
- `LastWriteError()` — most recent output writer error, if any
- `Flush()` — flush the async writer, if one is in use
- Environment: setting `IQLOG_DEBUG=true` enables debug level for newly
  created loggers.

### Context and legacy compatibility

- `WithContext(ctx)` returns a copy of the global logger carrying the
  context (copies share the writer lock with the original).
- `Log(ctx, level, msg)`, `Logf(ctx, level, format, args...)`, and
  `LogWithFields(ctx, level, fields, msg, args...)` provide a
  level-parameterised API (fields are a `map[string]any`; iteration order
  is not deterministic).
- `Print`, `Printf`, `Println` log at info level (Viper-style interface).
- `WithFields(map[string]any)` returns a handler with per-level methods
  (`Info`, `Errorf`, ...) that attach the fields to every line.

### Performance design

Log lines are built by pooled builders (`sync.Pool`) using pre-sized byte
slices, hand-written integer/float/timestamp encoders, and a bounded line
length (`maxLineLen = 1024` for the fixed-buffer strategies). The
repository contains several alternative line-builder implementations
(`bufferLine`, `bufferLineNL`, `preallocLine`, `preallocLine2`, `varStack`)
that were used to benchmark different allocation strategies; the
byte-slice builder (`ByteSliceLine*` / `With*` / the global level helpers)
is the primary, recommended path. Benchmarks live in `iqlog_test.go`.

## Limitations

- `WithGroup` and slog-style attribute grouping are currently no-ops; the
  `log/slog` handler integration is not yet enabled.
- Fixed-buffer strategies (`preallocLine`, `preallocLine2`) truncate lines
  longer than 1024 bytes. In JSON mode a truncated line is repaired so the
  emitted record is still valid JSON (trailing partial fields may be
  dropped).
- Fields in `LogWithFields`/`WithFields` are stringified with `%v`.
- Syslog output uses UDP only and is unavailable on Windows.

## Development

```bash
go test -race -count=1 ./...
go vet ./...
```

`go vet` reports two known `unsafe`/`reflect.SliceHeader` warnings in
`loc_fmt.go` and `loc_name_file_line_unsafe.go`; these are intentional
(performance-critical runtime introspection).
