// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service/db"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service/server"
)

type options struct {
	Port int `short:"p" long:"port" description:"server's port" required:"true"`
}

func run(logger *log.Logger, serverPort int) error {
	logger.Println("main: initializing gRPC server")
	defer logger.Println("main: Completed")

	// =========================================================================
	// Database support

	const sqlitePath = "db/booksGrpcService.db"
	db, err := db.ConnectToSqlite(sqlitePath)
	if err != nil {
		return errors.Wrap(err, "connecting to database")
	}

	// =========================================================================
	// Listener init

	port := fmt.Sprintf(":%d", serverPort)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return errors.Wrap(err, "tcp listening")
	}

	// =========================================================================
	// Server init

	srv := server.New(logger, db)

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		logger.Printf("main: gRPC server listening on %s", port)
		serverErrors <- srv.GrpcSrv.Serve(lis)
	}()

	// =========================================================================
	// Shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		logger.Println("main: received signal for shutdown: ", sig)
		srv.GrpcSrv.Stop()
	}
	return nil
}

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	logger := log.New(os.Stdout, "BOOKS GRPC SERVER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if err := run(logger, opts.Port); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
