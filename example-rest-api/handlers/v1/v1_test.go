// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package v1_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-templates/example-rest-api/db"
	"github.com/tiagomelo/go-templates/example-rest-api/handlers"
)

var (
	testDb     *sql.DB
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	const sqliteDbFile = "../../db/booksRestApiTest.db"
	var err error
	testDb, err = db.ConnectToSqlite(sqliteDbFile)
	if err != nil {
		fmt.Println("error when connecting to the test database:", err)
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	apiMux := handlers.NewApiMux(&handlers.ApiMuxConfig{
		Db:  testDb,
		Log: log,
	})
	testServer = httptest.NewServer(apiMux)
	defer testServer.Close()
	exitVal := m.Run()
	if err := os.Remove(sqliteDbFile); err != nil {
		fmt.Println("error when deleting test database:", err)
		os.Exit(1)
	}
	os.Exit(exitVal)
}

func TestV1Create(t *testing.T) {
	input := `{"title":"some title","author":"some author","pages":100}`
	expectedOutput := `{"id":1,"title":"some title","author":"some author","pages":100}`
	resp, err := http.Post(testServer.URL+"/api/v1/book", "application/json", bytes.NewBuffer([]byte(input)))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b))
}

func TestV1GetById(t *testing.T) {
	bookId := 1
	expectedOutput := `{"id":1,"title":"some title","author":"some author","pages":100}`
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/book/%d", testServer.URL, bookId))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b))
}

func TestV1List(t *testing.T) {
	expectedOutput := `[{"id":1,"title":"some title","author":"some author","pages":100}]`
	resp, err := http.Get(testServer.URL + "/api/v1/books")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b))
}

func TestV1Update(t *testing.T) {
	bookId := "1"
	input := `{"title":"new title","author":"new author","pages":150}`
	expectedOutput := `{"id":1,"title":"new title","author":"new author","pages":150}`
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/book/%s", testServer.URL, bookId), bytes.NewBuffer([]byte(input)))
	require.NoError(t, err)
	vars := map[string]string{
		"id": bookId,
	}
	req = mux.SetURLVars(req, vars)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b))
}

func TestV1DeleteById(t *testing.T) {
	bookId := "1"
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/book/%s", testServer.URL, bookId), nil)
	require.NoError(t, err)
	vars := map[string]string{
		"id": bookId,
	}
	req = mux.SetURLVars(req, vars)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}
