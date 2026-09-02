//go:build !darwin && !windows

package iqlog

// prepareNativeLog is available on macOS (os_log) and Windows (Event Log)
// only.
func prepareNativeLog(Config) (nativeLogWriter, error) {
	return nil, errNativeLogUnsupported
}
