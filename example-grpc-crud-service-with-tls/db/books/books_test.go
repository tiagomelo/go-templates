// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package books

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service-with-tls/db/books/models"
)

func TestList(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func() *sql.DB
		expectedOutput []*models.Book
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "title", "author", "pages"}).
						AddRow(1, "some title", "some author", 100).
						AddRow(2, "another title", "another author", 150))
				return db
			},
			expectedOutput: []*models.Book{
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
			},
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnError(errors.New("select error"))
				return db
			},
			expectedError: errors.New("listing books: select error"),
		},
		{
			name: "error on scan",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows([]string{"id", "title", "author", "pages"}).
					AddRow("invalid", "data", "types", "here")
				mock.ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)

				return db
			},
			expectedError: errors.New(`scanning book: sql: Scan error on column index 0, name "id": converting driver.Value type string ("invalid") to a int: invalid syntax`),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			output, err := List(context.TODO(), db)
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

func TestGetById(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func() *sql.DB
		expectedOutput *models.Book
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery(regexp.QuoteMeta(getByIdQuery)).WithArgs(1).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "title", "author", "pages"}).
						AddRow(1, "some title", "some author", 100))
				return db
			},
			expectedOutput: &models.Book{
				Id:     1,
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
		},
		{
			name: "no book found",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery(regexp.QuoteMeta(getByIdQuery)).WithArgs(1).
					WillReturnError(sql.ErrNoRows)
				return db
			},
			expectedError: errors.New("no book with id 1 found"),
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery(regexp.QuoteMeta(getByIdQuery)).WithArgs(1).
					WillReturnError(errors.New("select error"))
				return db
			},
			expectedError: errors.New("getting book with id 1: select error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			output, err := GetById(context.TODO(), db, 1)
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

func TestCreate(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func() *sql.DB
		input          *models.NewBook
		expectedOutput *models.NewBook
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(createQuery)).WithArgs("some title", "some author", 100).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db
			},
			input: &models.NewBook{
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
			expectedOutput: &models.NewBook{
				Id:     1,
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(createQuery)).WithArgs("some title", "some author", 100).
					WillReturnError(errors.New("insert error"))
				return db
			},
			input: &models.NewBook{
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
			expectedError: errors.New("inserting book: insert error"),
		},
		{
			name: "error, duplicate book",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(createQuery)).WithArgs("some title", "some author", 100).
					WillReturnError(sqlite3.Error{Code: sqlite3.ErrConstraint})
				return db
			},
			input: &models.NewBook{
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
			expectedError: errors.New(`book with title "some title" from author "some author" already exists`),
		},
		{
			name: "error on last insert id",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(createQuery)).WithArgs("some title", "some author", 100).
					WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("last insert id error")))
				return db
			},
			input: &models.NewBook{
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
			expectedError: errors.New("getting last insert id: last insert id error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			output, err := Create(context.TODO(), db, tc.input)
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

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func() *sql.DB
		input          *models.UpdatedBook
		expectedOutput *models.UpdatedBook
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WithArgs("some title", "some author", 100, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db
			},
			input: &models.UpdatedBook{
				Id:     1,
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
			expectedOutput: &models.UpdatedBook{
				Id:     1,
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WithArgs("some title", "some author", 100, 1).
					WillReturnError(errors.New("update error"))
				return db
			},
			input: &models.UpdatedBook{
				Id:     1,
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
			expectedError: errors.New("updating book with id 1: update error"),
		},
		{
			name: "error on rows affected",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WithArgs("some title", "some author", 100, 1).
					WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))
				return db
			},
			input: &models.UpdatedBook{
				Id:     1,
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
			expectedError: errors.New("checking affected rows: rows affected error"),
		},
		{
			name: "no book found",
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WithArgs("some title", "some author", 100, 1).
					WillReturnResult(sqlmock.NewResult(0, 0))
				return db
			},
			input: &models.UpdatedBook{
				Id:     1,
				Title:  "some title",
				Author: "some author",
				Pages:  100,
			},
			expectedError: errors.New("no book with id 1 found"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			output, err := Update(context.TODO(), db, tc.input)
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

func TestDeleteById(t *testing.T) {
	testCases := []struct {
		name          string
		input         int
		mockClosure   func() *sql.DB
		expectedError error
	}{
		{
			name:  "happy path",
			input: 1,
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(deleteByIdQuery)).WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				return db
			},
		},
		{
			name:  "error",
			input: 1,
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(deleteByIdQuery)).WithArgs(1).
					WillReturnError(errors.New("delete error"))
				return db
			},
			expectedError: errors.New("deleting book with id 1: delete error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			err := DeleteById(context.TODO(), db, tc.input)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
			}
		})
	}
}
