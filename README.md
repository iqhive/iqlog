# iqlog

A low-allocation structured logging library for Go 1.23+.

The primary API follows standard Go conventions: construct a concrete logger,
configure it with a value, and let consuming packages define any narrow
interfaces they need. Historical iqlog helper names remain available as thin,
deprecated adapters for gradual migration.

```go
logger, err := iqlog.New(iqlog.Config{
    Format:      iqlog.FormatJSON,
    Level:       iqlog.LevelInfo,
    Writer:      os.Stderr,
    IncludeTime: true,
    CallerDepth: 1,
})
if err != nil {
    return err
}

logger.Info("service started")
logger.Infof("listening on %s", ":8080")
logger.InfoEvent().
    Str("user", "alice").
    Int("requests", 42).
    Msg("request handled")
```

## Configuration

`Config{}` is valid and creates a synchronous console logger writing to stderr
at info level. `New` returns configuration or writer setup errors. `MustNew`
provides explicit panic-on-error construction for static application setup.

`Config` supports console or JSON output, levels, writers, timestamps and
layouts, caller depth, ANSI color, context field extraction, application/syslog
metadata, writer mode, buffering, overflow policy, and an injectable fatal exit
function.

`WriterMode` uses `WriterSync`, `WriterAsync`, or `WriterRing`. `WriterRing`
is a compatibility mode backed by the same channel queue as `WriterAsync`,
with ring-compatible defaults. `OverflowPolicy` uses `OverflowBlock`,
`OverflowDrop`, or `OverflowSync`. Dropped and
background writer errors are available from `LastWriteError` and `Flush`.
`Dropped()` reports the cumulative number of overflow-dropped records.

Package functions use `Default()`. Replace it atomically with
`SetDefault(logger)`. `SetDefault` does not close the previous logger.

For new code, prefer configuring complete values during construction instead
of calling setters one at a time. This produces one coherent configuration and
makes dependencies explicit.

## Logging

Package functions and `Logger` methods provide:

- `Trace`, `Debug`, `Info`, `Warn`, `Error`, `Panic`, and `Fatal`
- `...f` formatted and `...ln` println-style variants
- `Print`, `Printf`, and `Println`
- `Log`, `Logf`, and `Logln` for a dynamic `Level`
- `LogContext` and `LogContextf` for explicit context propagation

`Panic` writes and then panics with the formatted message. `Fatal` flushes,
writes when enabled, and invokes `Config.ExitFunc(1)`, which defaults to
`os.Exit`. Panic and fatal still terminate when their record is filtered.

Use `Enabled(level)` before constructing expensive arguments.

## Structured Events

`Logger.Event(level)` and the level-specific `InfoEvent`, `ErrorEvent`, and
related constructors return a single-use `*Event`. Package-level event
constructors use `Default()`.

Events support `Str`, `Int`, `Int64`, `Uint`, `Uint64`, `Float32`, `Float64`,
`Bool`, `Err`, `Any`, `Time`, `Duration`, `Bytes`, and `RawJSON`. Invalid raw
JSON is encoded as `null`. `Any` preserves JSON-compatible values in JSON mode
and uses `%v` in console mode.

Finish an event with `Msg`, `Msgs`, or `Msgf`. Finishing consumes it; subsequent
calls are ignored. Events are not safe for concurrent use.

`WithFields`, `WithError`, and `WithContext` return logger copies. Copies keep
the originating writer and immutable configuration snapshot. Persistent fields
are emitted deterministically. A configured `ContextExtractor` runs once for
enabled context-aware records; explicit context overrides stored context.

## Lifecycle

`Flush` waits for accepted queued records and returns errors accumulated since
the previous flush. `LastWriteError` remains a sticky diagnostic. `Close`
drains internal writer workers, is idempotent, and does not close caller-owned
underlying writers. After close, `Enabled` is false and reconfiguration or
flushing returns `ErrClosed`.

## Levels

`Level` implements `String`, `encoding.TextMarshaler`, and
`encoding.TextUnmarshaler`. `ParseLevel` accepts `trace`, `debug`, `info`,
`warn`/`warning`, `error`, `panic`, and `fatal`.

## Compatibility Helpers

Legacy convenience APIs remain supported and delegate to the primary API:

- `NewIQLogger(jsonMode)` and `NewGlobalIQLogger()` call `New(Config)`.
- `Init(...)` constructs and installs a default logger.
- `Warning`, `Warningf`, and `Warningln` call the corresponding `Warn` helper.
- `InfoWith` and the other `*With` helpers call the corresponding `*Event`
  constructor.
- Legacy configuration and writer methods update `Config` or install the
  requested writer through the current writer lifecycle implementation.
- `LevelPrint` aliases `LevelInfo`.
- `LogWithFields` and `LogFWithFields` compose `WithFields` with
  `LogContext`/`LogContextf`.

These helpers are intentionally adapters rather than a second implementation.
New documentation and examples use `New(Config)`, `InfoEvent`, `Warn`, and
explicit `LogContext` calls.

The historical context-first `Log(ctx, level, ...)` signature cannot coexist
with `Log(level, ...)` because Go does not support overloaded functions. The
idiomatic API keeps `Log(level, ...)` and names the context-aware operation
`LogContext(ctx, level, ...)`.

`SetNewLine` is retained as a compatibility no-op. Loggers always emit complete
records, and JSON remains newline-delimited.

## Development

```bash
go build ./...
go test -race -count=1 ./...
go vet ./...
```
