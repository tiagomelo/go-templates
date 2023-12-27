// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package v1

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tiagomelo/go-templates/example-rest-api/handlers/v1/books"
	"github.com/tiagomelo/go-templates/example-rest-api/middleware"
)

// Config struct holds the database connection and logger.
type Config struct {
	Db  *sql.DB
	Log *slog.Logger
}

// Routes initializes and returns a new router with configured routes.
func Routes(c *Config) *mux.Router {
	router := mux.NewRouter()
	initializeRoutes(c.Db, router)
	router.Use(
		func(h http.Handler) http.Handler {
			return middleware.Logger(c.Log, h)
		},
		middleware.Compress,
		middleware.PanicRecovery,
	)
	return router
}

// initializeRoutes sets up the routes for book operations.
func initializeRoutes(db *sql.DB, router *mux.Router) {
	booksHandlers := books.New(db)
	router.HandleFunc("/v1/book", booksHandlers.Create).Methods(http.MethodPost)
	router.HandleFunc("/v1/book/{id}", booksHandlers.Update).Methods(http.MethodPut)
	router.HandleFunc("/v1/book/{id}", booksHandlers.GetById).Methods(http.MethodGet)
	router.HandleFunc("/v1/book/{id}", booksHandlers.DeleteById).Methods(http.MethodDelete)
	router.HandleFunc("/v1/books", booksHandlers.List).Methods(http.MethodGet)
}
