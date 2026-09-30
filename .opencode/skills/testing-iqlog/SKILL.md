---
name: testing-iqlog
description: How to runtime-verify the iqlog Go logging library (builders, JSON mode, writers)
---

# Testing iqlog

Pure Go library, no app/UI - verify with `go build ./...`, `go vet ./...`,
`go test -race -count=1 ./...`, plus small adversarial programs.

Performance is a release requirement. For changes touching logging hot paths,
also run the `bench/` competitor suite with `-benchmem -count=5`. Disabled and
normal typed events must remain allocation-free. Optional features must not add
work to the default path; gate them before caller lookup, context extraction,
escaping, formatting, or synchronization work begins.

- `go vet ./...` should pass cleanly.
- Build adversarial programs in a SEPARATE module with
  `replace github.com/iqhive/iqlog => /path/to/checkout` (copy the repo's go.sum).
  Re-usable harnesses live in /home/ubuntu/repos/iqlog-verify (injection, numkeys, misc,
  ringload, readme) if that box still exists.
- The supported builder is `*Event`, reached with `logger.Event(level)` or
  `logger.InfoEvent()` and related level helpers.
- JSON message key is `"message"` (not `"msg"`); caller fields are `"func"` and `"file"`.
- `Event` is the only structured builder implementation. Test it in console
  and JSON modes, including caller capture and non-finite values.
- For queued-writer tests, use a mutex-guarded counting writer that splits on '\n'
  and json.Unmarshals every line; run with `go run -race`, then poll until the consumer
  goroutine drains (line count == expected) with a deadline to detect deadlocks.
- `FatalEvent().Msg()` invokes `Config.ExitFunc(1)` after flushing; inject the
  exit function in unit tests.
- The `log/slog` bridge is `(*Logger).SlogHandler()` and the package-level
  `SlogHandler()`. Drive it through a `*slog.Logger` with typed `LogAttrs` so
  slog's own argument boxing does not show up as handler allocations;
  `TestSlogHandlerHandleDoesNotAllocate` must stay at zero, and
  `testing/slogtest` must pass. Caller output for slog records follows
  `CallerDepth`, and the record's own time is used, not `Config.Now`.
- No secrets needed.
