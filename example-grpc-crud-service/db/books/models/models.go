// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package models

// Book represents the model for a book record.
type Book struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Pages  int    `json:"pages"`
}

// NewBook is used to create a new book record.
type NewBook struct {
	Id     int    `json:"id"`
	Title  string `json:"title" validate:"required"`
	Author string `json:"author" validate:"required"`
	Pages  int    `json:"pages" validate:"required,gt=0"`
}

// UpdateBook is used to update a book record.
type UpdatedBook struct {
	Id     int    `json:"id" validate:"required"`
	Title  string `json:"title" validate:"required"`
	Author string `json:"author" validate:"required"`
	Pages  int    `json:"pages" validate:"required,gt=0"`
}
