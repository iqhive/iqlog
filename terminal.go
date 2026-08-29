package iqlog

import (
	"io"
	"os"

	"golang.org/x/term"
)

type fileDescriptor interface {
	Fd() uintptr
}

func terminalWriter(writer io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if writer == nil {
		writer = os.Stderr
	}
	fd, ok := writer.(fileDescriptor)
	return ok && term.IsTerminal(int(fd.Fd()))
}
