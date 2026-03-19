// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package hello

import (
	"net/http"

	"github.com/tiagomelo/go-templates/example-rest-api/web"
)

// handlers struct holds any dependencies for the handlers.
type handlers struct{}

// New initializes a new instance of handlers.
func New() *handlers {
	return &handlers{}
}

// SayHello handles the HTTP request to say hello.
func (h *handlers) SayHello(w http.ResponseWriter, r *http.Request) {
	web.RespondWithJson(w, http.StatusOK, map[string]string{"message": "Hello, World!"})
}
