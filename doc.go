// Package iqlog is a fast, low-allocation structured logger.
//
// A [Logger] is built from a [Config] and writes either coloured console
// lines or newline-delimited JSON. Records are built with typed events:
//
//	log := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: os.Stdout})
//	log.InfoEvent().Str("service", "checkout").Int("port", 8080).Msg("listening")
//
// Disabled events and enabled events made of typed fields do not allocate.
// Caller capture, context extraction, custom timestamps, and colour are
// optional and cost nothing until they are switched on. CallerDepth N reports
// the N-th frame at or above the library boundary, where 1 is the direct
// caller; if fewer frames exist, no caller is emitted. The bounded scan examines
// at most 32 frames, so deeper callers are reported as absent. iqlog-family frames,
// including the legacy bitbucket.org/iqhive/iqlog/v3 wrapper, are never
// reported. [Logger.EventAt] requires a non-empty function; a file without a
// function is ignored, and empty function and file values suppress the caller.
//
// Package-level functions such as [Info] and [InfoEvent] write through the
// logger returned by [Default], which [SetDefault] replaces.
//
// Code that speaks log/slog, such as libraries and the standard log package
// once slog.SetDefault has been called, can write through a logger via
// [Logger.SlogHandler]. A record's program counter is used when present;
// zero-PC records, including log-bridge and hand-built records, fall back to a
// live-stack scan. See [Logger.SlogHandler] for the complete mapping rules.
package iqlog
