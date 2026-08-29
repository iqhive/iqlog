package main

import (
	"os"

	"github.com/iqhive/iqlog"
)

func main() {
	log := iqlog.MustNew(iqlog.Config{Writer: os.Stdout, Color: true})
	log.InfoEvent().
		Str("service", "checkout").
		Int("port", 8080).
		Bool("tls", true).
		Msg("listening")
}
