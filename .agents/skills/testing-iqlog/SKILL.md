---
name: testing-iqlog
description: How to runtime-verify the iqlog Go logging library (builders, JSON mode, writers)
---

# Testing iqlog

Pure Go library, no app/UI — verify with `go build ./...`, `go vet ./...`,
`go test -race -count=1 ./...`, plus small adversarial programs.

- `go vet` intentionally reports two unsafe warnings (loc_fmt.go reflect.SliceHeader,
  loc_name_file_line_unsafe.go unsafe.Pointer). These are expected; anything else is a regression.
- Build adversarial programs in a SEPARATE module with
  `replace github.com/iqhive/iqlog => /path/to/checkout` (copy the repo's go.sum).
  Re-usable harnesses live in /home/ubuntu/repos/iqlog-verify (injection, numkeys, misc,
  ringload, readme) if that box still exists.
- The six line builders are reached via exported methods on `*logger`
  (from `iqlog.NewIQLogger(jsonMode bool)`): `WithByteSliceLineInfo()`,
  `WithBufferLineInfo()`, `WithBufferLineNLInfo()`, `WithPreallocLineInfo()`,
  `WithPreallocLine2Info()`, `WithVarStackInfo()` — each returns a fluent builder with
  `Str/Int/Int64/Float32/Float64/Bool/Any` and `Msg/Msgs/Msgf`.
- JSON message key is `"message"` (not `"msg"`); caller fields are `"func"` and `"file"`.
- Only the bytesliceLine `With*` path captures callers; preallocLine/preallocLine2/varStack
  With* builders never populate callerData (pre-existing design).
- Watch out: varStack has its own float formatting in `varstack_write.go` that may bypass
  fixes made in `formatting.go` (e.g. non-finite float fallback) — test varStack floats
  separately from the shared fast paths.
- For ring-buffer/writer tests, use a mutex-guarded counting writer that splits on '\n'
  and json.Unmarshals every line; run with `go run -race`, then poll until the consumer
  goroutine drains (line count == expected) with a deadline to detect deadlocks.
- `WithByteSliceLineFatal().Msg()` calls os.Exit(1) — test it in a subprocess and assert
  exit code 1 AND that the line reached stderr first.
- No secrets needed.
