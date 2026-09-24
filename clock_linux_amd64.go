//go:build linux && amd64

package iqlog

import (
	"syscall"
	"time"
)

// wallClock returns CLOCK_REALTIME as Unix seconds and microseconds from a
// single reading of the clock.
//
// syscall.Gettimeofday on linux/amd64 is one call through
// runtime.vdsoGettimeofdaySym ($GOROOT/src/syscall/asm_linux_amd64.s), the
// kernel's __vdso_gettimeofday, with the system call taken only when no vDSO
// is mapped. It reads the same clock as clock_gettime(CLOCK_REALTIME) and
// reports it truncated to the microsecond, which is all the fixed-width
// timestamps carry, so every stamp is exactly the wall clock at the instant
// of the read. time.Now makes two vDSO calls, CLOCK_REALTIME and then
// CLOCK_MONOTONIC, around a switch to the g0 stack, and a timestamp never
// uses the monotonic half; on the benchmark host this reading costs about
// 30ns against time.Now's 58ns. gettimeofday is declared //go:noescape, so
// the Timeval stays on the stack and the read allocates nothing.
//
// The other Linux architectures generate Gettimeofday as a raw system call,
// which is far slower than time.Now, so they take the fallback in
// clock_other.go.
func wallClock() (sec, usec int64) {
	var tv syscall.Timeval
	if syscall.Gettimeofday(&tv) != nil {
		// gettimeofday cannot fail for a valid pointer; if it ever did the
		// record would still be stamped from the wall clock
		now := time.Now()
		return now.Unix(), int64(now.Nanosecond()) / 1000
	}
	return tv.Sec, tv.Usec
}
