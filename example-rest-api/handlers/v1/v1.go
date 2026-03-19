// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package v1

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tiagomelo/go-templates/example-rest-api/handlers/v1/hello"
	"github.com/tiagomelo/go-templates/example-rest-api/middleware"
)

// Config holds the configuration for the API routes, including the logger.
type Config struct {
	Log *slog.Logger
}

// Routes sets up the API routes for the application and applies middleware.
func Routes(c *Config) *mux.Router {
	router := mux.NewRouter()
	initializeRoutes(router)
	router.Use(
		func(h http.Handler) http.Handler {
			return middleware.Logger(c.Log, h)
		},
		middleware.Compress,
		middleware.PanicRecovery,
	)
	return router
}

// initializeRoutes sets up the API routes for the application.
func initializeRoutes(router *mux.Router) {
	helloHandlers := hello.New()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/v1/hello", helloHandlers.SayHello).Methods(http.MethodGet)
}
