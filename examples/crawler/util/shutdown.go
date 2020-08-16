package util

import (
	"context"
	"os"
	"os/signal"
)

// Trap Ctrl+C and call cancel on the context
func ShutdownCtx() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	//defer func() {
	//	signal.Stop(sigChan)
	//	cancel()
	//}()

	go func() {
		select {
		case <-sigChan:
			cancel()
		}
	}()
	return ctx, cancel
}
