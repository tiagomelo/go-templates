// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tiagomelo/go-templates/example-grpc-crud-service-with-jwt/auth"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service-with-jwt/config"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service/api/proto/gen/book"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("failed to read config: ", err)
		os.Exit(1)
	}
	authSvc := auth.NewService(cfg.JwtKey)
	token, err := authSvc.IssueJWTToken("123456")
	if err != nil {
		fmt.Println("failed to issue JWT token: ", err)
		os.Exit(1)
	}
	md := metadata.Pairs("authorization", token, "user_id", "123456")
	ctx = metadata.NewOutgoingContext(ctx, md)

	const serverHost = "localhost:4444"
	conn, err := grpc.NewClient(serverHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("failed to dial server: ", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := book.NewBookServiceClient(conn)
	books, err := client.GetAllBooks(ctx, &book.GetAllBooksRequest{})
	if err != nil {
		fmt.Println("failed to get all books: ", err)
		os.Exit(1)
	}
	fmt.Printf("books: %v\n", books)
}
