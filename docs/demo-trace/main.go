package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/iqhive/iqlog"
)

type requestIDKey struct{}

func main() {
	log := iqlog.MustNew(iqlog.Config{
		Writer: os.Stdout,
		Format: iqlog.FormatJSON,
		ContextExtractor: func(ctx context.Context) map[string]any {
			if id, ok := ctx.Value(requestIDKey{}).(string); ok {
				return map[string]any{"request_id": id}
			}
			return nil
		},
	})

	fmt.Println("CONTEXT TRACE")
	fmt.Println()

	// Request 1
	fmt.Println("--> GET /orders/42  x-request-id: req-7f3a")
	ctx1 := context.WithValue(context.Background(), requestIDKey{}, "req-7f3a")
	log.LogContext(ctx1, iqlog.LevelInfo, "request accepted")
	log.WithContext(ctx1).InfoEvent().Int("status", 200).Msg("response sent")
	fmt.Println()

	// Request 2 interleaves
	fmt.Println("--> GET /users/7  x-request-id: req-b912")
	ctx2 := context.WithValue(context.Background(), requestIDKey{}, "req-b912")
	log.LogContext(ctx2, iqlog.LevelInfo, "request accepted")
	fmt.Println()

	// Failure on request 1
	fmt.Println("--> GET /orders/42 upstream fails")
	log.WithContext(ctx1).ErrorEvent().Str("error", "upstream timeout").Msg("upstream failed")
	fmt.Println()

	fmt.Println("CONTEXT IN // CORRELATED FIELDS OUT")

	time.Sleep(100 * time.Millisecond)
}
