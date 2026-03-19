// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package hello

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSayHello(t *testing.T) {
	t.Run("should return a greeting message", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "hello", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		h := New()
		handler := http.HandlerFunc((h).SayHello)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, `{"message":"Hello, World!"}`, rr.Body.String())
	})
}
