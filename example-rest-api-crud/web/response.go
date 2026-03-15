// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package web

import (
	"encoding/json"
	"net/http"
)

// RespondWithError responds a json with an error message
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJson(w, code, map[string]string{"error": message})
}

// RespondWithJson responds a json with an error message
func RespondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithStatus responds with an HTTP status code
func RespondWithStatus(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}
