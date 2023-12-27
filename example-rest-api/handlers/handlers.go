// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package handlers

import (
	"database/sql"
	"log/slog"

	"github.com/gorilla/mux"
	v1 "github.com/tiagomelo/go-templates/example-rest-api/handlers/v1"
)

// ApiMuxConfig struct holds the configuration for the API.
type ApiMuxConfig struct {
	Db  *sql.DB
	Log *slog.Logger
}

// NewApiMux creates and returns a new mux.Router configured with version 1 (v1) routes.
func NewApiMux(c *ApiMuxConfig) *mux.Router {
	return v1.Routes(&v1.Config{
		Db:  c.Db,
		Log: c.Log,
	})
}
