//go:build windows

package iqlog

import "golang.org/x/sys/windows"

// consoleSupportsColor reports whether fd is a console that will interpret
// ANSI colour sequences. A console handle alone is not enough: a classic
// conhost session renders the escapes as text unless virtual-terminal
// processing is enabled on the output handle, so it is switched on here,
// once, the way other colour-aware loggers do. If the console refuses, the
// output stays plain.
func consoleSupportsColor(fd uintptr) bool {
	h := windows.Handle(fd)
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return false
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}
	return windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}
