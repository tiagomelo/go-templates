// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tiagomelo/go-templates/example-grpc-crud-service/api/proto/gen/book"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
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
	}
	fmt.Printf("%v\n", books)
}
