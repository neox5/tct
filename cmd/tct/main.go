package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/neox5/tct/internal/app"
	"github.com/neox5/tct/internal/generator"
	"github.com/neox5/tct/internal/handler"
	"github.com/neox5/tct/internal/metrics"
	"github.com/neox5/tct/internal/server"
	"github.com/neox5/tct/internal/version"
)

func main() {
	// Handle --version flag
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("tct", version.String())
		os.Exit(0)
	}

	// Initialize application
	app, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialization failed: %v\n", err)
		os.Exit(1)
	}

	app.Logger.Info("starting tct", "version", version.String(), "mode", app.Mode)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run mode-specific logic
	var runErr error
	switch app.Mode {
	case "sender":
		runErr = runSender(ctx, app)
	case "receiver":
		runErr = runReceiver(ctx, app)
	default:
		fmt.Fprintf(os.Stderr, "invalid mode: %s\n", app.Mode)
		os.Exit(1)
	}

	if runErr != nil && runErr != context.Canceled {
		app.Logger.Error("runtime error", "error", runErr)
		os.Exit(1)
	}

	app.Logger.Info("shutdown complete")
}

// runSender starts the sender mode: HTTP server for observability + request generator.
func runSender(ctx context.Context, app *app.App) error {
	m := metrics.NewSenderMetrics()

	// Start HTTP server for observability
	srv := server.New(app.Config.SenderPort, app.Logger)
	srv.RegisterCommonRoutes(handler.Healthz, handler.Readyz)

	// Run server in background
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- srv.Start(ctx)
	}()

	// Run generator (blocks until context cancelled)
	generatorDone := make(chan error, 1)
	go func() {
		generatorDone <- generator.Run(ctx, app.Config, app.Logger, m)
	}()

	// Wait for either to complete
	select {
	case err := <-serverDone:
		return err
	case err := <-generatorDone:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// runReceiver starts the receiver mode: HTTP server with /inbox endpoint.
func runReceiver(ctx context.Context, app *app.App) error {
	m := metrics.NewReceiverMetrics()

	// Start HTTP server
	srv := server.New(app.Config.ReceiverPort, app.Logger)
	srv.RegisterCommonRoutes(handler.Healthz, handler.Readyz)
	srv.RegisterHandler("POST /inbox", handler.InboxHandler(app.Config, app.Logger, m))

	return srv.Start(ctx)
}
