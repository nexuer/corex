package corex

import (
	"context"
	"os"
	"os/signal"

	"github.com/nexuer/log"
)

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
// Only one of SetupSignalContext and SetupSignalHandler should be called, and only can
// be called once.
func SetupSignalHandler() <-chan struct{} {
	return SetupSignalContext().Done()
}

// SetupSignalContext is same as SetupSignalHandler, but a context.Context is returned.
// Only one of SetupSignalContext and SetupSignalHandler should be called, and only can
// be called once.
func SetupSignalContext() context.Context {
	shutdownHandler := make(chan os.Signal, 2)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(shutdownHandler, shutdownSignals...)
	go func() {
		sig := <-shutdownHandler
		log.Infof("signal: received %s, shutting down gracefully", sig)
		cancel()
		sig = <-shutdownHandler
		log.Warnf("signal: received %s again, forcing immediate exit", sig)
		os.Exit(1) // second signal. Exit directly.
	}()
	return ctx
}
