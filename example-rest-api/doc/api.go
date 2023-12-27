package doc

import "github.com/tiagomelo/go-templates/example-rest-api/db/books/models"

// swagger:route GET /v1/books books List
// List all books.
// ---
// responses:
//		200: listBooksResponse
//		500: description: internal server error

// swagger:response listBooksResponse
type ListBooksdResponseWrapper struct {
	// in:body
	Body []models.Book
}

// swagger:route GET /v1/book/{id} book GetById
// Get a book by its id.
// ---
// responses:
//		200: getBookByIdResponse
//		400: description: invalid id
//		404: description: book not found
//		500: description: internal server error

// swagger:parameters GetById
type GetBookByIdParamWrapper struct {
	// in:path
	Id int
}

// swagger:response getBookByIdResponse
type GetBookByIdResponseWrapper struct {
	// in:body
	Body models.Book
}

// swagger:route POST /v1/book book Create
// Create a book.
// ---
// responses:
//		201: createBookResponse
//		400: description: missing required fields
//		500: description: internal server error

// swagger:response createBookResponse
type CreateBookResponseWrapper struct {
	// in:body
	Body models.NewBook
}

// swagger:route PUT /v1/book/{id} book Update
// Update a book.
// ---
// responses:
//		200: updateBookResponse
//		400: description: invalid id
//		400: description: missing required fields
//		404: description: book not found
//		500: description: internal server error

// swagger:parameters Update
type UpdateBookByIdParamWrapper struct {
	// in:path
	Id int
}

// swagger:response updateBookResponse
type UpdateBookResponseWrapper struct {
	// in:body
	Body models.UpdatedBook
}

// swagger:route DELETE /v1/book/{id} book DeleteById
// Delete a book by its id.
// ---
// responses:
//		204: description: success
//		400: description: invalid id
//		500: description: internal server error

// swagger:parameters DeleteById
type DeleteBookByIdParamWrapper struct {
	// in:path
	Id int
}
