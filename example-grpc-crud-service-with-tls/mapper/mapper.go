// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

// Package mapper provides functions to convert between
// gRPC data models and database models for book entities.
package mapper

import (
	"github.com/tiagomelo/go-templates/example-grpc-crud-service-with-tls/api/proto/gen/book"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service-with-tls/db/books/models"
)

// NewBookDbModel converts a Book protobuf message to a NewBook database model.
// It is used when creating a new book entry in the database.
func NewBookDbModel(book *book.Book) *models.NewBook {
	return &models.NewBook{
		Title:  book.GetTitle(),
		Author: book.GetAuthor(),
		Pages:  int(book.GetPages()),
	}
}

// UpdatedBookDbModel converts a Book protobuf message to an UpdatedBook database model.
// It is used when updating an existing book entry in the database.
func UpdatedBookDbModel(book *book.Book) *models.UpdatedBook {
	return &models.UpdatedBook{
		Id:     int(book.GetId()),
		Title:  book.GetTitle(),
		Author: book.GetAuthor(),
		Pages:  int(book.GetPages()),
	}
}

// BookProto converts a Book database model to a Book protobuf message.
// It is typically used when sending book data back to the client.
func BookProto(dbBook *models.Book) *book.Book {
	return &book.Book{
		Id:     int32(dbBook.Id),
		Title:  dbBook.Title,
		Author: dbBook.Author,
		Pages:  int32(dbBook.Pages),
	}
}

// BookProtoList converts a list of Book database models to a slice of Book protobuf messages.
// It is used when sending multiple book records back to the client.
func BookProtoList(dbBooks []*models.Book) []*book.Book {
	bookProtoList := []*book.Book{}
	for _, dbBook := range dbBooks {
		bookProto := &book.Book{
			Id:     int32(dbBook.Id),
			Title:  dbBook.Title,
			Author: dbBook.Author,
			Pages:  int32(dbBook.Pages),
		}
		bookProtoList = append(bookProtoList, bookProto)
	}
	return bookProtoList
}
