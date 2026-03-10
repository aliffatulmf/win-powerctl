package timeout

import (
	"context"
	"log"
	"time"

	"win-powerctl/internal/shutdown"
)

func Start(ctx context.Context, timeout time.Duration) {
	go func() {
		select {
		case <-time.After(timeout):
			log.Println("timeout: escalating to force shutdown")
			if err := shutdown.ForceIfHung(); err != nil {
				log.Printf("force shutdown failed: %v", err)
			}
		case <-ctx.Done():
			log.Println("timeout: cancelled")
		}
	}()
}
