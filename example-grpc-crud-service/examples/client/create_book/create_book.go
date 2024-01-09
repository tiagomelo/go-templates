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
	conn, err := grpc.Dial(serverHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("failed to dial server: ", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := book.NewBookServiceClient(conn)
	newBook := &book.Book{
		Title:  "title",
		Author: "author",
		Pages:  100,
	}
	createdBook, err := client.CreateBook(ctx, &book.CreateBookRequest{
		Book: newBook,
	})
	if err != nil {
		fmt.Println("failed to create book: ", err)
	}
	fmt.Printf("created book: %v\n", createdBook)
}
