package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/iqhive/iqlog"
)

// slowWriter adds a delay to every write so the async queue backs up.
type slowWriter struct {
	out io.Writer
}

func (sw *slowWriter) Write(p []byte) (int, error) {
	time.Sleep(80 * time.Millisecond)
	return sw.out.Write(p)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: demo-queue <block|drop|sync>")
		os.Exit(1)
	}

	policyStr := os.Args[1]
	var policy iqlog.OverflowPolicy
	switch policyStr {
	case "block":
		policy = iqlog.OverflowBlock
		fmt.Println("WRITER ASYNC // OVERFLOW BLOCK")
	case "drop":
		policy = iqlog.OverflowDrop
		fmt.Println("WRITER ASYNC // OVERFLOW DROP")
	case "sync":
		policy = iqlog.OverflowSync
		fmt.Println("WRITER ASYNC // OVERFLOW SYNC")
	default:
		fmt.Fprintf(os.Stderr, "unknown policy: %s\n", policyStr)
		os.Exit(1)
	}
	fmt.Println()

	// Small buffer so we can overflow it easily; slow writer so it backs up
	log := iqlog.MustNew(iqlog.Config{
		Writer:         &slowWriter{out: os.Stdout},
		Format:         iqlog.FormatConsole,
		IncludeTime:    false,
		WriterMode:     iqlog.WriterAsync,
		BufferSize:     3,
		OverflowPolicy: policy,
	})

	start := time.Now()
	var wg sync.WaitGroup
	for i := 1; i <= 8; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.InfoEvent().Int("rec", i).Msg("queued")
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	_ = log.Flush()

	dropped := log.Dropped()
	fmt.Printf("Flush() ====> written, dropped=%d, elapsed=%v\n", dropped, elapsed.Round(time.Millisecond))
}
