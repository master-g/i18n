package wksig

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var (
	// InterruptChan as signal to control goroutine to finish
	InterruptChan = make(chan struct{})
	// interruptSignals defines system signals to catch
	interruptSignals = []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM}
)

// Start waiting UNIX system signal
func Start() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, interruptSignals...)
	<-c
	close(InterruptChan)
}

// StartWithContext waiting UNIX system signal with context
func StartWithContext(ctx context.Context) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, interruptSignals...)

	select {
	case <-c:
	case <-ctx.Done():
	}
	close(InterruptChan)
}

// StartWithStopChannel waiting UNIX system signal with a stop channel
func StartWithStopChannel(stop chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, interruptSignals...)

	select {
	case <-c:
	case <-stop:
	}
	close(InterruptChan)
}
