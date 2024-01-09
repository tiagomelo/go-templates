// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package books

import (
	"context"
	"database/sql"

	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	bookErrors "github.com/tiagomelo/go-templates/example-grpc-crud-service/db/books/errors"
	"github.com/tiagomelo/go-templates/example-grpc-crud-service/db/books/models"
)

// SQL queries as constants for CRUD operations on the 'books' table.
const (
	listQuery = `
	SELECT id, title, author, pages
	FROM books 
	`

	getByIdQuery = `
	SELECT id, title, author, pages
	FROM books
	WHERE id = $1
	`

	createQuery = `
	INSERT INTO books (title, author, pages)
	VALUES ($1, $2, $3)
	`

	updateQuery = `
	UPDATE books
	SET title = $1, author = $2, pages = $3
	WHERE id = $4
	`

	deleteByIdQuery = `
	DELETE FROM books
	WHERE id = $1
	`
)

// List retrieves all books from the database.
func List(ctx context.Context, db *sql.DB) ([]*models.Book, error) {
	rows, err := db.QueryContext(ctx, listQuery)
	if err != nil {
		return nil, errors.Wrap(err, "listing books")
	}
	defer rows.Close()
	books := []*models.Book{}
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Pages); err != nil {
			return nil, errors.Wrap(err, "scanning book")
		}
		books = append(books, &book)
	}
	return books, nil
}

// GetById retrieves a book by its ID.
func GetById(ctx context.Context, db *sql.DB, bookId int) (*models.Book, error) {
	row := db.QueryRowContext(ctx, getByIdQuery, bookId)
	var book models.Book
	if err := row.Scan(&book.Id, &book.Title, &book.Author, &book.Pages); err != nil {
		if err == sql.ErrNoRows {
			return nil, &bookErrors.ErrBookNotFound{Id: bookId}
		}
		return nil, errors.Wrapf(err, "getting book with id %d", bookId)
	}
	return &book, nil
}

// Create adds a new book record to the database.
func Create(ctx context.Context, db *sql.DB, newBook *models.NewBook) (*models.NewBook, error) {
	result, err := db.ExecContext(ctx, createQuery, newBook.Title, newBook.Author, newBook.Pages)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return nil, &bookErrors.ErrDuplicateBook{Title: newBook.Title, Author: newBook.Author}
		}
		return nil, errors.Wrap(err, "inserting book")
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "getting last insert id")
	}
	newBook.Id = int(id)
	return newBook, nil
}

// Update modifies an existing book record.
func Update(ctx context.Context, db *sql.DB, book *models.UpdatedBook) (*models.UpdatedBook, error) {
	result, err := db.ExecContext(ctx, updateQuery, book.Title, book.Author, book.Pages, book.Id)
	if err != nil {
		return nil, errors.Wrapf(err, "updating book with id %d", book.Id)
	}
	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "checking affected rows")
	}
	if rowsUpdated == 0 {
		return nil, &bookErrors.ErrBookNotFound{Id: book.Id}
	}
	return book, nil
}

// DeleteById removes a book record by its ID.
func DeleteById(ctx context.Context, db *sql.DB, bookId int) error {
	if _, err := db.ExecContext(ctx, deleteByIdQuery, bookId); err != nil {
		return errors.Wrapf(err, "deleting book with id %d", bookId)
	}
	return nil
}
