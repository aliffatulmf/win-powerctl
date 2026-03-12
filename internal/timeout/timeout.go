package timeout

import (
	"context"
	"time"

	"win-powerctl/internal/logger"
	"win-powerctl/internal/shutdown"
)

func Start(ctx context.Context, timeout time.Duration) {
	go func() {
		select {
		case <-time.After(timeout):
			logger.Info("timeout", "escalating to force shutdown")
			if err := shutdown.ForceIfHung(); err != nil {
				logger.Error("timeout", "force shutdown failed", "error", err)
			}
		case <-ctx.Done():
			logger.Info("timeout", "cancelled")
		}
	}()
}
