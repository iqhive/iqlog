//go:build !race

package iqlog

// raceEnabled reports whether the binary was built with -race. See the
// build-tagged counterpart of this file.
const raceEnabled = false
