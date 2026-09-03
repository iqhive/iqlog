//go:build race

package iqlog

// raceEnabled reports whether the binary was built with -race. The race
// detector allocates for its own bookkeeping, so allocation-count assertions
// measure the detector rather than the logger and are skipped under it.
const raceEnabled = true
