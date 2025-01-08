// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service-with-tls/api/proto/gen/book"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func run() error {
	ctx := context.Background()
	const serverHost = "localhost:4444"

	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return errors.Wrap(err, "loading CA's certificate")
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return errors.New("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return errors.Wrap(err, "loading client's certificate and private key")
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}
	conn, err := grpc.NewClient(serverHost, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	if err != nil {
		return errors.Wrap(err, "dialing")
	}

	// Create the client
	client := book.NewBookServiceClient(conn)

	books, err := client.GetAllBooks(ctx, &book.GetAllBooksRequest{})
	if err != nil {
		return errors.Wrap(err, "gettingh all books")
	}
	fmt.Printf("%v\n", books)

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
