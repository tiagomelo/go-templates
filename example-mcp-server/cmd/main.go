// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/tiagomelo/go-templates/example-mcp-server/server"
	"github.com/tiagomelo/go-templates/example-mcp-server/tools"
)

func run(ctx context.Context, logger *slog.Logger) error {
	mcpServer := server.New(os.Stdin, os.Stdout, logger)
	tools.RegisterDefaultTools(mcpServer)

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Create a cancellable context so we can signal the server to stop.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for JSON-RPC requests.
	go func() {
		serverErrors <- mcpServer.Run(ctx)
	}()

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.WithMessage(err, "MCP server error")
	case sig := <-shutdown:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
		cancel()
		return nil
	}
}

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(log.Writer(), nil))
	if err := run(ctx, logger); err != nil {
		logger.Error("error", slog.Any("err", err))
		os.Exit(1)
	}
}
