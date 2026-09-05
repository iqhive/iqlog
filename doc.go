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
// optional and cost nothing until they are switched on.
//
// Package-level functions such as [Info] and [InfoEvent] write through the
// logger returned by [Default], which [SetDefault] replaces.
//
// Code that speaks log/slog, such as libraries and the standard log package
// once slog.SetDefault has been called, can write through a logger via
// [Logger.SlogHandler]; see that method for the mapping rules.
package iqlog
