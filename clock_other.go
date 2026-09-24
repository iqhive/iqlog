//go:build !(linux && amd64)

package iqlog

import "time"

// wallClock returns the wall clock as Unix seconds and microseconds. Only
// linux/amd64 has a single-read path (clock_linux_amd64.go); everywhere else
// syscall.Gettimeofday is absent or a raw system call, so the reading is
// time.Now's own, truncated to the microsecond the fixed-width encoders
// carry, and those encoders produce the same bytes they would from the full
// time.Time.
func wallClock() (sec, usec int64) {
	now := time.Now()
	return now.Unix(), int64(now.Nanosecond()) / 1000
}
