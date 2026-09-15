# Migration

The preferred API is configuration-driven, but historical convenience helpers
remain available as deprecated adapters.

| Historical API | Preferred API |
| --- | --- |
| `NewIQLogger(false)` | `New(Config{Format: FormatConsole})` or `MustNew(...)` |
| `NewIQLogger(true)` | `New(Config{Format: FormatJSON})` or `MustNew(...)` |
| `NewGlobalIQLogger()` | `MustNew(Config{IncludeTime: true, CallerDepth: 1})` |
| `GlobalLogger` | `Default()` / `SetDefault()` |
| `InfoWith()` | `InfoEvent()` |
| `Warning(...)` | `Warn(...)` |
| `SetWriter(w)` | `New(Config{Writer: w})` or `SetConfig` |
| `SetLevel(level)` | `New(Config{Level: level})` or `SetConfig` |
| `Log(ctx, level, msg)` | `LogContext(ctx, level, msg)` |
| `LogWithFields(...)` | `WithFields(fields).LogContext(...)` |

The old context-first `Log` and `Logf` signatures are the only common helpers
that cannot be retained under the same names, because Go has no function
overloading. All non-conflicting aliases route to the same current
implementation.

## Caller attribution

`CallerDepth = N` reports the N-th frame at or above the library boundary;
`N=1` is the direct caller of the logging entry point. If fewer frames exist,
no caller is emitted. iqlog never reports its own package-family frames,
including the legacy `bitbucket.org/iqhive/iqlog/v3` wrapper, so applications
using that shim are attributed to application code.

`EventAt` emits caller output only when its function argument is non-empty. A
file without a function is ignored; empty function and file arguments suppress
caller output.

## From log/slog

Code that already uses `log/slog` keeps working unchanged through
`slog.SetDefault(slog.New(log.SlogHandler()))`. When caller capture is enabled,
a record's own program counter is used when present. Records without one,
including records from the `log` bridge and hand-built records, fall back to a
live-stack scan. This changes the previous behavior: zero-PC slog records are
no longer caller-less. The nonzero-PC path reports the record's call site
directly and does not use `CallerDepth` values above 1 to select a later frame.
If a hand-built zero-PC record is forwarded through a wrapper handler or
middleware, the scan attributes it to that wrapper's frame.
When rewriting hot paths onto typed events, the calls map like this:

| log/slog | iqlog |
| --- | --- |
| `slog.New(handler)` | `iqlog.New(Config{...})` |
| `slog.Info(msg, "k", v)` | `log.InfoEvent().Str("k", v).Msg(msg)` |
| `slog.Error(msg, "err", err)` | `log.ErrorEvent().Err(err).Msg(msg)` |
| `logger.With("k", v)` | `log.WithFields(map[string]any{"k": v})` |
| `logger.WithGroup("g")` | Prefix keys yourself: `Str("g.k", v)` |
| `slog.InfoContext(ctx, msg)` | `log.LogContext(ctx, iqlog.LevelInfo, msg)` with a `ContextExtractor` |
| `HandlerOptions.Level` / `slog.LevelVar` | `Config.Level`, `SetConfig` |
| `HandlerOptions.AddSource` | `Config.CallerDepth` |
| `slog.LevelDebug - 4` and below | `LevelTrace` |
| `slog.LevelError + n` | `LevelError`; use `Fatal` or `Panic` explicitly to terminate |
