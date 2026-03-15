// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package web

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Param retrieves a path parameter from the URL of an HTTP request.
func Param(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}

// BookIdPathParam extracts the 'id' parameter from the path of an HTTP request
// and converts it to an integer.
func BookIdPathParam(r *http.Request) (int, error) {
	idParam := Param(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return 0, ErrInvalidBookId
	}
	return id, nil
}
