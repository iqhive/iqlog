package iqlog

import (
	"io"
	"os"
)

type fileDescriptor interface {
	Fd() uintptr
}

// terminalWriter reports whether writer is a terminal that should receive
// ANSI colour. The environment conventions are honoured first; the
// descriptor probe is platform-specific because a Windows console only
// interprets ANSI sequences once virtual-terminal processing is enabled on
// it.
func terminalWriter(writer io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if writer == nil {
		writer = os.Stderr
	}
	fd, ok := writer.(fileDescriptor)
	return ok && consoleSupportsColor(fd.Fd())
}
