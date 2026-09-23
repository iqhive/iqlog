//go:build !windows

package iqlog

import "golang.org/x/term"

// consoleSupportsColor reports whether fd is a terminal.
func consoleSupportsColor(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}
