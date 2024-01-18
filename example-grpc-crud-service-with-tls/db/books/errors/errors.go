// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package errors

import "fmt"

// ErrBookNotFound represents an error when a book is not found in the database.
type ErrBookNotFound struct {
	Id int
}

func (e ErrBookNotFound) Error() string {
	return fmt.Sprintf("no book with id %d found", e.Id)
}

// ErrDuplicateBook represents an error when a book with the same title and author already exists.
type ErrDuplicateBook struct {
	Title  string
	Author string
}

func (e ErrDuplicateBook) Error() string {
	return fmt.Sprintf(`book with title "%s" from author "%s" already exists`, e.Title, e.Author)
}
