// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package server

import (
	"context"
	"database/sql"
	"log"

	"github.com/pkg/errors"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service/api/proto/gen/book"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service/db/books"
	bookErrors "github.com/tiagomelo/go-templates/example-grpc-crud-service/db/books/errors"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service/mapper"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service/validate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// For ease of unit testing.
var (
	listBooks   = books.List
	getBookById = books.GetById
	createBook  = books.Create
	updateBook  = books.Update
	deleteBook  = books.DeleteById
)

// server implements BookServiceServer.
type server struct {
	book.UnimplementedBookServiceServer
	GrpcSrv *grpc.Server

	logger *log.Logger
	db     *sql.DB
}

// New creates and returns a new server instance.
// It initializes the gRPC server and registers the BookService.
func New(logger *log.Logger, db *sql.DB) *server {
	grpServer := grpc.NewServer()
	srv := &server{
		GrpcSrv: grpServer,
		logger:  logger,
		db:      db,
	}
	book.RegisterBookServiceServer(grpServer, srv)
	return srv
}

// GetAllBooks handles the GetAllBooks gRPC call.
// It retrieves all books from the database and returns them.
func (s *server) GetAllBooks(ctx context.Context, in *book.GetAllBooksRequest) (*book.GetAllBooksResponse, error) {
	books, err := listBooks(ctx, s.db)
	if err != nil {
		s.logger.Println(err)
		return nil, status.Error(codes.Internal, errors.Wrap(err, "getting all books").Error())
	}
	return &book.GetAllBooksResponse{
		Books: mapper.BookProtoList(books),
	}, nil
}

// GetBook handles the GetBook gRPC call.
// It retrieves a single book by its ID and returns it.
func (s *server) GetBook(ctx context.Context, in *book.GetBookRequest) (*book.Book, error) {
	book, err := getBookById(ctx, s.db, int(in.GetId()))
	if err != nil {
		var errNotFound *bookErrors.ErrBookNotFound
		if errors.As(err, &errNotFound) {
			s.logger.Println(err)
			return nil, status.Error(codes.NotFound, errNotFound.Error())
		}
		s.logger.Println(err)
		return nil, status.Error(codes.Internal, errors.Wrapf(err, "getting book with id %d", in.GetId()).Error())
	}
	return mapper.BookProto(book), nil
}

// CreateBook handles the CreateBook gRPC call.
// It creates a new book record in the database.
func (s *server) CreateBook(ctx context.Context, in *book.CreateBookRequest) (*book.Book, error) {
	newBook := mapper.NewBookDbModel(in.Book)
	if err := validate.Check(newBook); err != nil {
		s.logger.Printf("create book validation error: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	createdBook, err := createBook(ctx, s.db, newBook)
	if err != nil {
		var errDuplicateBook *bookErrors.ErrDuplicateBook
		if errors.As(err, &errDuplicateBook) {
			s.logger.Println(errDuplicateBook)
			return nil, status.Error(codes.AlreadyExists, errDuplicateBook.Error())
		}
		s.logger.Printf("error when creating book: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	in.Book.Id = int32(createdBook.Id)
	return in.Book, nil
}

// UpdateBook handles the UpdateBook gRPC call.
// It updates an existing book record in the database.
func (s *server) UpdateBook(ctx context.Context, in *book.UpdateBookRequest) (*book.Book, error) {
	updatedBook := mapper.UpdatedBookDbModel(in.Book)
	if err := validate.Check(updatedBook); err != nil {
		s.logger.Printf("update book validation error: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, err := updateBook(ctx, s.db, updatedBook)
	if err != nil {
		var errBookNotFound *bookErrors.ErrBookNotFound
		if errors.As(err, &errBookNotFound) {
			s.logger.Println(errBookNotFound)
			return nil, status.Error(codes.NotFound, errBookNotFound.Error())
		}
		s.logger.Printf("error when updating book: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return in.Book, nil
}

// DeleteBook handles the DeleteBook gRPC call.
// It deletes a book record from the database by its ID.
func (s *server) DeleteBook(ctx context.Context, in *book.DeleteBookRequest) (*book.DeleteBookResponse, error) {
	if err := deleteBook(ctx, s.db, int(in.GetId())); err != nil {
		s.logger.Printf("error when deleting book with id %d: %v", in.GetId(), err)
		return nil, status.Error(codes.Internal, err.Error())

	}
	return &book.DeleteBookResponse{
		Id: in.GetId(),
	}, nil
}
