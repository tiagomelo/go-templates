// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
)

// Logger is a middleware that logs the start and end of each HTTP request along with
// some additional information.
func Logger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()
		log.Info("request started",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remoteaddr", r.RemoteAddr),
		)
		next.ServeHTTP(w, r)
		log.Info("request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remoteaddr", r.RemoteAddr),
			slog.Duration("since", time.Since(start)),
		)
	})
}

// Compress is a middleware that applies compression to HTTP responses.
func Compress(next http.Handler) http.Handler {
	return handlers.CompressHandler(next)
}

// PanicRecovery is a middleware that recovers from panics in the application,
// preventing the server from crashing and logging the stack trace.
func PanicRecovery(next http.Handler) http.Handler {
	return handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(next)
}
