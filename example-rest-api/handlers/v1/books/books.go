// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package books

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/tiagomelo/go-templates/example-rest-api/db/books"
	"github.com/tiagomelo/go-templates/example-rest-api/db/books/models"
	"github.com/tiagomelo/go-templates/example-rest-api/validate"
	"github.com/tiagomelo/go-templates/example-rest-api/web"
)

// handlers struct holds a database connection.
type handlers struct {
	db *sql.DB
}

// New initializes a new instance of handlers with a database connection.
func New(db *sql.DB) *handlers {
	return &handlers{
		db: db,
	}
}

// For ease of unit testing.
var (
	listBooks   = books.List
	getBookById = books.GetById
	createBook  = books.Create
	updateBook  = books.Update
	deleteBook  = books.DeleteById

	// jsonDecode decodes a JSON request body into a given struct.
	jsonDecode = func(r io.Reader, v any) error {
		return json.NewDecoder(r).Decode(v)
	}
)

// List handles the HTTP request to list all books.
func (h *handlers) List(w http.ResponseWriter, r *http.Request) {
	books, err := listBooks(r.Context(), h.db)
	if err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	web.RespondWithJson(w, http.StatusOK, books)
}

// GetById handles the HTTP request to retrieve a book by its ID.
func (h *handlers) GetById(w http.ResponseWriter, r *http.Request) {
	bookId, err := web.BookIdPathParam(r)
	if err != nil {
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	book, err := getBookById(r.Context(), h.db, bookId)
	if err != nil {
		var bookNotFoundErr *books.ErrBookNotFound
		if errors.As(err, &bookNotFoundErr) {
			web.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	web.RespondWithJson(w, http.StatusOK, book)
}

// Create handles the HTTP request to create a new book.
func (h *handlers) Create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var newBook models.NewBook
	if err := jsonDecode(r.Body, &newBook); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validate.Check(newBook); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	book, err := createBook(r.Context(), h.db, &newBook)
	if err != nil {
		var duplicateBookErr *books.ErrDuplicateBook
		if errors.As(err, &duplicateBookErr) {
			web.RespondWithError(w, http.StatusConflict, err.Error())
			return
		}
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	web.RespondWithJson(w, http.StatusCreated, book)
}

// Update handles the HTTP request to update an existing book.
func (h *handlers) Update(w http.ResponseWriter, r *http.Request) {
	bookId, err := web.BookIdPathParam(r)
	if err != nil {
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()
	var updatedBook models.UpdatedBook
	if err := jsonDecode(r.Body, &updatedBook); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	updatedBook.Id = bookId
	if err := validate.Check(updatedBook); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	book, err := updateBook(r.Context(), h.db, &updatedBook)
	if err != nil {
		var bookNotFoundErr *books.ErrBookNotFound
		if errors.As(err, &bookNotFoundErr) {
			web.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	web.RespondWithJson(w, http.StatusOK, book)
}

// DeleteById handles the HTTP request to delete a book by its ID.
func (h *handlers) DeleteById(w http.ResponseWriter, r *http.Request) {
	bookId, err := web.BookIdPathParam(r)
	if err != nil {
		web.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := deleteBook(r.Context(), h.db, bookId); err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	web.RespondWithStatus(w, http.StatusNoContent)
}
