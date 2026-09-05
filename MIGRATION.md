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

## From log/slog

Code that already uses `log/slog` keeps working unchanged through
`slog.SetDefault(slog.New(log.SlogHandler()))`. When rewriting hot paths onto
typed events, the calls map like this:

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
