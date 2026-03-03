package watchdog

import (
	"log"
	"time"

	"win-powerctl/internal/shutdown"
)

func Start(timeout time.Duration) {
	go func() {
		time.Sleep(timeout)

		log.Println("watchdog: escalating to force shutdown")
		if err := shutdown.ForceIfHung(); err != nil {
			log.Println("force shutdown failed:", err)
		}
	}()
}
