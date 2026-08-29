package main

import (
	"fmt"
	"os"
	"time"

	"github.com/iqhive/iqlog"
)

func main() {
	fmt.Println("FORMAT A // CONSOLE")
	fmt.Println()

	consoleLog := iqlog.MustNew(iqlog.Config{
		Writer:      os.Stdout,
		Color:       true,
		IncludeTime: false,
	})
	consoleLog.DebugEvent().Str("cache", "warm").Msg("cache probe")
	consoleLog.InfoEvent().Int("port", 8080).Msg("listening")
	consoleLog.WarnEvent().Int("remaining", 2).Msg("connection pool low")
	consoleLog.ErrorEvent().Str("error", "timeout").Msg("upstream failed")

	fmt.Println()
	fmt.Println("-- INPUT --")
	fmt.Println()

	fmt.Println("FORMAT B // JSON")
	fmt.Println()

	jsonLog := iqlog.MustNew(iqlog.Config{
		Format: iqlog.FormatJSON,
		Writer: os.Stdout,
	})
	jsonLog.DebugEvent().Str("cache", "warm").Msg("cache probe")
	jsonLog.InfoEvent().Int("port", 8080).Msg("listening")
	jsonLog.WarnEvent().Int("remaining", 2).Msg("connection pool low")
	jsonLog.ErrorEvent().Str("error", "timeout").Msg("upstream failed")

	fmt.Println()
	fmt.Println("ONE EVENT // TWO ENCODINGS")

	// Keep output visible
	time.Sleep(100 * time.Millisecond)
}
