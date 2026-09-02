//go:build darwin && !cgo

package iqlog

// prepareNativeLog requires cgo to reach os_log; builds with CGO_ENABLED=0
// cannot create the writer.
func prepareNativeLog(Config) (nativeLogWriter, error) {
	return nil, errNativeLogCGO
}
