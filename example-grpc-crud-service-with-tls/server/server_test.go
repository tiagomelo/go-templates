// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package server

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service-with-tls/api/proto/gen/book"
	bookErrors "github.com/tiagomelo/go-templates/example-grpc-crud-service-with-tls/db/books/errors"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service-with-tls/db/books/models"
	"google.golang.org/grpc/credentials"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name          string
		mockTlsCreds  func() (credentials.TransportCredentials, error)
		expectedError error
	}{
		{
			name: "happy path",
			mockTlsCreds: func() (credentials.TransportCredentials, error) {
				return nil, nil
			},
		},
		{
			name: "error",
			mockTlsCreds: func() (credentials.TransportCredentials, error) {
				return nil, errors.New("some error")
			},
			expectedError: errors.New("loading TLS creds: some error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tlsCreds = tc.mockTlsCreds
			logger := log.New(io.Discard, "", 0)
			s, err := New(logger, nil)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.NotNil(t, s)
			}
		})
	}
}

func TestGetAllBooks(t *testing.T) {
	testCases := []struct {
		name           string
		mockListBooks  func(ctx context.Context, db *sql.DB) ([]*models.Book, error)
		expectedOutput *book.GetAllBooksResponse
		expectedError  error
	}{
		{
			name: "happy path",
			mockListBooks: func(ctx context.Context, db *sql.DB) ([]*models.Book, error) {
				return []*models.Book{
					{
						Id:     1,
						Title:  "title",
						Author: "author",
						Pages:  100,
					},
					{
						Id:     2,
						Title:  "another title",
						Author: "another author",
						Pages:  150,
					},
				}, nil
			},
			expectedOutput: &book.GetAllBooksResponse{
				Books: []*book.Book{
					{
						Id:     1,
						Title:  "title",
						Author: "author",
						Pages:  100,
					},
					{
						Id:     2,
						Title:  "another title",
						Author: "another author",
						Pages:  150,
					},
				},
			},
		},
		{
			name: "error",
			mockListBooks: func(ctx context.Context, db *sql.DB) ([]*models.Book, error) {
				return nil, errors.New("list books error")
			},
			expectedError: errors.New("rpc error: code = Internal desc = getting all books: list books error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			listBooks = tc.mockListBooks
			logger := log.New(io.Discard, "", 0)
			s := &server{
				logger: logger,
			}
			output, err := s.GetAllBooks(context.TODO(), &book.GetAllBooksRequest{})
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestGetBook(t *testing.T) {
	testCases := []struct {
		name            string
		mockGetBookById func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error)
		expectedOutput  *book.Book
		expectedError   error
	}{
		{
			name: "happy path",
			mockGetBookById: func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error) {
				return &models.Book{
					Id:     1,
					Title:  "title",
					Author: "author",
					Pages:  100,
				}, nil
			},
			expectedOutput: &book.Book{
				Id:     1,
				Title:  "title",
				Author: "author",
				Pages:  100,
			},
		},
		{
			name: "does not exist",
			mockGetBookById: func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error) {
				return nil, &bookErrors.ErrBookNotFound{Id: 1}
			},
			expectedError: errors.New("rpc error: code = NotFound desc = no book with id 1 found"),
		},
		{
			name: "error",
			mockGetBookById: func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error) {
				return nil, errors.New("get book error")
			},
			expectedError: errors.New("rpc error: code = Internal desc = getting book with id 1: get book error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getBookById = tc.mockGetBookById
			logger := log.New(io.Discard, "", 0)
			s := &server{
				logger: logger,
			}
			output, err := s.GetBook(context.TODO(), &book.GetBookRequest{Id: 1})
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestCreateBook(t *testing.T) {
	testCases := []struct {
		name           string
		input          *book.CreateBookRequest
		mockCreateBook func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error)
		expectedOutput *book.Book
		expectedError  error
	}{
		{
			name: "happy path",
			input: &book.CreateBookRequest{
				Book: &book.Book{
					Title:  "title",
					Author: "author",
					Pages:  100,
				},
			},
			mockCreateBook: func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
				return &models.NewBook{
					Id:     1,
					Title:  "title",
					Author: "author",
					Pages:  100,
				}, nil
			},
			expectedOutput: &book.Book{
				Id:     1,
				Title:  "title",
				Author: "author",
				Pages:  100,
			},
		},
		{
			name: "invalid input",
			input: &book.CreateBookRequest{
				Book: &book.Book{},
			},
			expectedError: errors.New(`rpc error: code = InvalidArgument desc = [{"field":"title","error":"title is a required field"},{"field":"author","error":"author is a required field"},{"field":"pages","error":"pages is a required field"}]`),
		},
		{
			name: "already exists",
			input: &book.CreateBookRequest{
				Book: &book.Book{
					Title:  "title",
					Author: "author",
					Pages:  100,
				},
			},
			mockCreateBook: func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
				return nil, &bookErrors.ErrDuplicateBook{
					Title:  "title",
					Author: "author",
				}
			},
			expectedError: errors.New(`rpc error: code = AlreadyExists desc = book with title "title" from author "author" already exists`),
		},
		{
			name: "error",
			input: &book.CreateBookRequest{
				Book: &book.Book{
					Title:  "title",
					Author: "author",
					Pages:  100,
				},
			},
			mockCreateBook: func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
				return nil, errors.New("create book error")
			},
			expectedError: errors.New("rpc error: code = Internal desc = create book error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createBook = tc.mockCreateBook
			logger := log.New(io.Discard, "", 0)
			s := &server{
				logger: logger,
			}
			output, err := s.CreateBook(context.TODO(), tc.input)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestUpdateBook(t *testing.T) {
	testCases := []struct {
		name           string
		input          *book.UpdateBookRequest
		mockUpdateBook func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error)
		expectedOutput *book.Book
		expectedError  error
	}{
		{
			name: "happy path",
			input: &book.UpdateBookRequest{
				Book: &book.Book{
					Id:     1,
					Title:  "new title",
					Author: "new author",
					Pages:  150,
				},
			},
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return &models.UpdatedBook{
					Id:     1,
					Title:  "new title",
					Author: "new author",
					Pages:  150,
				}, nil
			},
			expectedOutput: &book.Book{
				Id:     1,
				Title:  "new title",
				Author: "new author",
				Pages:  150,
			},
		},
		{
			name: "invalid input",
			input: &book.UpdateBookRequest{
				Book: &book.Book{},
			},
			expectedError: errors.New(`rpc error: code = InvalidArgument desc = [{"field":"id","error":"id is a required field"},{"field":"title","error":"title is a required field"},{"field":"author","error":"author is a required field"},{"field":"pages","error":"pages is a required field"}]`),
		},
		{
			name: "does not exist",
			input: &book.UpdateBookRequest{
				Book: &book.Book{
					Id:     1,
					Title:  "new title",
					Author: "new author",
					Pages:  150,
				},
			},
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return nil, &bookErrors.ErrBookNotFound{Id: 1}
			},
			expectedError: errors.New("rpc error: code = NotFound desc = no book with id 1 found"),
		},
		{
			name: "error",
			input: &book.UpdateBookRequest{
				Book: &book.Book{
					Id:     1,
					Title:  "new title",
					Author: "new author",
					Pages:  150,
				},
			},
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return nil, errors.New("update book error")
			},
			expectedError: errors.New("rpc error: code = Internal desc = update book error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updateBook = tc.mockUpdateBook
			logger := log.New(io.Discard, "", 0)
			s := &server{
				logger: logger,
			}
			output, err := s.UpdateBook(context.TODO(), tc.input)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestDeleteBook(t *testing.T) {
	testCases := []struct {
		name           string
		input          *book.DeleteBookRequest
		mockDeleteBook func(ctx context.Context, db *sql.DB, bookId int) error
		expectedOutput *book.DeleteBookResponse
		expectedError  error
	}{
		{
			name: "happy path",
			input: &book.DeleteBookRequest{
				Id: int32(1),
			},
			mockDeleteBook: func(ctx context.Context, db *sql.DB, bookId int) error {
				return nil
			},
			expectedOutput: &book.DeleteBookResponse{
				Id: 1,
			},
		},
		{
			name: "error",
			input: &book.DeleteBookRequest{
				Id: int32(1),
			},
			mockDeleteBook: func(ctx context.Context, db *sql.DB, bookId int) error {
				return errors.New("delete book error")
			},
			expectedError: errors.New("rpc error: code = Internal desc = delete book error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteBook = tc.mockDeleteBook
			logger := log.New(io.Discard, "", 0)
			s := &server{
				logger: logger,
			}
			output, err := s.DeleteBook(context.TODO(), tc.input)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}
