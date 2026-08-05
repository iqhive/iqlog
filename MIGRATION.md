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
