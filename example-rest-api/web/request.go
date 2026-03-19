// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Param retrieves a path parameter from the URL of an HTTP request.
func Param(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}
