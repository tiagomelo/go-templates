// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package books

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-templates/example-rest-api/db/books"
	"github.com/tiagomelo/go-templates/example-rest-api/db/books/models"
)

func TestList(t *testing.T) {
	testCases := []struct {
		name               string
		mockListBooks      func(ctx context.Context, db *sql.DB) ([]*models.Book, error)
		expectedOutput     string
		expectedStatusCode int
	}{
		{
			name: "happy path",
			mockListBooks: func(ctx context.Context, db *sql.DB) ([]*models.Book, error) {
				return []*models.Book{
					{
						Id:     1,
						Title:  "some title",
						Author: "some author",
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
			expectedOutput:     `[{"id":1,"title":"some title","author":"some author","pages":100},{"id":2,"title":"another title","author":"another author","pages":150}]`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "error",
			mockListBooks: func(ctx context.Context, db *sql.DB) ([]*models.Book, error) {
				return nil, errors.New("list error")
			},
			expectedOutput:     `{"error":"list error"}`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			listBooks = tc.mockListBooks
			req, err := http.NewRequest(http.MethodGet, "books", nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			h := New(nil)
			handler := http.HandlerFunc((h).List)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatusCode, rr.Code)
			require.Equal(t, tc.expectedOutput, rr.Body.String())
		})
	}
}

func TestGetById(t *testing.T) {
	testCases := []struct {
		name               string
		bookId             string
		mockGetBookById    func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error)
		expectedOutput     string
		expectedStatusCode int
	}{
		{
			name:   "happy path",
			bookId: "1",
			mockGetBookById: func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error) {
				return &models.Book{
					Id:     1,
					Title:  "some title",
					Author: "some author",
					Pages:  100,
				}, nil
			},
			expectedOutput:     `{"id":1,"title":"some title","author":"some author","pages":100}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "invalid book id",
			bookId: "invalidId",
			mockGetBookById: func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error) {
				return nil, nil
			},
			expectedOutput:     `{"error":"invalid book id"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "book not found",
			bookId: "1",
			mockGetBookById: func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error) {
				return nil, &books.ErrBookNotFound{Id: 1}
			},
			expectedOutput:     `{"error":"no book with id 1 found"}`,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:   "error",
			bookId: "1",
			mockGetBookById: func(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error) {
				return nil, errors.New("GetById error")
			},
			expectedOutput:     `{"error":"GetById error"}`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		getBookById = tc.mockGetBookById
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/book/%s", tc.bookId), nil)
			require.NoError(t, err)
			vars := map[string]string{
				"id": tc.bookId,
			}
			req = mux.SetURLVars(req, vars)
			rr := httptest.NewRecorder()
			h := New(nil)
			handler := http.HandlerFunc((h).GetById)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatusCode, rr.Code)
			require.Equal(t, tc.expectedOutput, rr.Body.String())
		})
	}
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		name               string
		input              string
		mockJsonDecode     func(r io.Reader, v any) error
		mockCreateBook     func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error)
		expectedOutput     string
		expectedStatusCode int
	}{
		{
			name:  "happy path",
			input: `{"title":"some title","author":"some author","pages":100}`,
			mockCreateBook: func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
				return &models.NewBook{
					Id:     1,
					Title:  "some title",
					Author: "some author",
					Pages:  100,
				}, nil
			},
			expectedOutput:     `{"id":1,"title":"some title","author":"some author","pages":100}`,
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:  "error on decoding payload",
			input: ``,
			mockJsonDecode: func(r io.Reader, v any) error {
				return errors.New("decode error")
			},
			mockCreateBook: func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
				return nil, nil
			},
			expectedOutput:     `{"error":"decode error"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:  "validation error",
			input: `{}`,
			mockCreateBook: func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
				return nil, nil
			},
			expectedOutput:     `{"error":"[{\"field\":\"title\",\"error\":\"title is a required field\"},{\"field\":\"author\",\"error\":\"author is a required field\"},{\"field\":\"pages\",\"error\":\"pages is a required field\"}]"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:  "duplicate book",
			input: `{"title":"some title","author":"some author","pages":100}`,
			mockCreateBook: func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
				return nil, &books.ErrDuplicateBook{Title: "some title", Author: "some author"}
			},
			expectedOutput:     `{"error":"book with title \"some title\" from author \"some author\" already exists"}`,
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:  "error",
			input: `{"title":"some title","author":"some author","pages":100}`,
			mockCreateBook: func(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
				return nil, errors.New("create error")
			},
			expectedOutput:     `{"error":"create error"}`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	originalJsonDecode := jsonDecode
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createBook = tc.mockCreateBook
			if tc.mockJsonDecode != nil {
				jsonDecode = tc.mockJsonDecode
			}
			defer func() {
				jsonDecode = originalJsonDecode
			}()
			req, err := http.NewRequest(http.MethodPost, "book", bytes.NewBuffer([]byte(tc.input)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			h := New(nil)
			handler := http.HandlerFunc((h).Create)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatusCode, rr.Code)
			require.Equal(t, tc.expectedOutput, rr.Body.String())
		})
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name               string
		bookId             string
		input              string
		mockJsonDecode     func(r io.Reader, v any) error
		mockUpdateBook     func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error)
		expectedOutput     string
		expectedStatusCode int
	}{
		{
			name:   "happy path",
			bookId: "1",
			input:  `{"title":"some title","author":"some author","pages":100}`,
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return &models.UpdatedBook{
					Id:     1,
					Title:  "some title",
					Author: "some author",
					Pages:  100,
				}, nil
			},
			expectedOutput:     `{"id":1,"title":"some title","author":"some author","pages":100}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "invalid book id",
			bookId: "invalidId",
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return nil, nil
			},
			expectedOutput:     `{"error":"invalid book id"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "error on decoding payload",
			bookId: "1",
			input:  ``,
			mockJsonDecode: func(r io.Reader, v any) error {
				return errors.New("decode error")
			},
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return nil, nil
			},
			expectedOutput:     `{"error":"decode error"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "validation error",
			bookId: "1",
			input:  `{}`,
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return nil, nil
			},
			expectedOutput:     `{"error":"[{\"field\":\"title\",\"error\":\"title is a required field\"},{\"field\":\"author\",\"error\":\"author is a required field\"},{\"field\":\"pages\",\"error\":\"pages is a required field\"}]"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "book not found",
			bookId: "1",
			input:  `{"title":"some title","author":"some author","pages":100}`,
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return nil, &books.ErrBookNotFound{Id: 1}
			},
			expectedOutput:     `{"error":"no book with id 1 found"}`,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:   "error",
			bookId: "1",
			input:  `{"title":"some title","author":"some author","pages":100}`,
			mockUpdateBook: func(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
				return nil, errors.New("update error")
			},
			expectedOutput:     `{"error":"update error"}`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	originalJsonDecode := jsonDecode
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updateBook = tc.mockUpdateBook
			if tc.mockJsonDecode != nil {
				jsonDecode = tc.mockJsonDecode
			}
			defer func() {
				jsonDecode = originalJsonDecode
			}()
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/book/%s", tc.bookId), bytes.NewBuffer([]byte(tc.input)))
			require.NoError(t, err)
			vars := map[string]string{
				"id": tc.bookId,
			}
			req = mux.SetURLVars(req, vars)
			rr := httptest.NewRecorder()
			h := New(nil)
			handler := http.HandlerFunc((h).Update)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatusCode, rr.Code)
			require.Equal(t, tc.expectedOutput, rr.Body.String())
		})
	}
}

func TestDeleteById(t *testing.T) {
	testCases := []struct {
		name               string
		bookId             string
		mockDeleteBook     func(ctx context.Context, db *sql.DB, bookId int) error
		expectedOutput     string
		expectedStatusCode int
	}{
		{
			name:   "happy path",
			bookId: "1",
			mockDeleteBook: func(ctx context.Context, db *sql.DB, bookId int) error {
				return nil
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:   "invalid book id",
			bookId: "invalidId",
			mockDeleteBook: func(ctx context.Context, db *sql.DB, bookId int) error {
				return nil
			},
			expectedOutput:     `{"error":"invalid book id"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "error",
			bookId: "1",
			mockDeleteBook: func(ctx context.Context, db *sql.DB, bookId int) error {
				return errors.New("delete error")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedOutput:     `{"error":"delete error"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteBook = tc.mockDeleteBook
			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/book/%s", tc.bookId), nil)
			require.NoError(t, err)
			vars := map[string]string{
				"id": tc.bookId,
			}
			req = mux.SetURLVars(req, vars)
			rr := httptest.NewRecorder()
			h := New(nil)
			handler := http.HandlerFunc((h).DeleteById)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatusCode, rr.Code)
			require.Equal(t, tc.expectedOutput, rr.Body.String())
		})
	}
}
