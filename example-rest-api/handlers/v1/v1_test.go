// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package v1_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-templates/example-rest-api/handlers"
)

func TestV1SayHello(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	apiMux := handlers.NewApiMux(&handlers.ApiMuxConfig{
		Log: log,
	})

	testServer := httptest.NewServer(apiMux)
	defer testServer.Close()

	expectedOutput := `{"message":"Hello, World!"}`
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/hello", testServer.URL))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b))
}
