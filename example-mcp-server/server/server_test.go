// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-templates/example-mcp-server/jsonrpc"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func parseResponses(t *testing.T, buf *bytes.Buffer) []jsonrpc.Response {
	t.Helper()
	var responses []jsonrpc.Response
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var resp jsonrpc.Response
		require.NoError(t, json.Unmarshal(line, &resp))
		responses = append(responses, resp)
	}
	return responses
}

func resultMap(t *testing.T, resp jsonrpc.Response) map[string]any {
	t.Helper()
	m, ok := resp.Result.(map[string]any)
	require.True(t, ok, "expected result to be a map, got %T", resp.Result)
	return m
}

const (
	initRequest      = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}`
	initNotification = `{"jsonrpc":"2.0","method":"notifications/initialized"}`
)

func initHandshake() string {
	return initRequest + "\n" + initNotification + "\n"
}

func registerEchoTool(s *Server) {
	s.RegisterTool(ToolDefinition{
		Name:        "echo",
		Description: "Echoes input.",
		InputSchema: map[string]any{"type": "object"},
	}, func(ctx context.Context, args json.RawMessage) (any, error) {
		return map[string]any{"echo": string(args)}, nil
	})
}

type errWriter struct{}

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write error")
}

type errReader struct{}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestRun_Initialize(t *testing.T) {
	out := &bytes.Buffer{}
	s := New(strings.NewReader(initRequest+"\n"), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)

	resp := responses[0]
	require.Nil(t, resp.Error)
	require.Equal(t, float64(1), resp.ID)

	result := resultMap(t, resp)
	require.Equal(t, ProtocolVersion, result["protocolVersion"])

	serverInfo, ok := result["serverInfo"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "go-mcp-server", serverInfo["name"])

	capabilities, ok := result["capabilities"].(map[string]any)
	require.True(t, ok)
	tools, ok := capabilities["tools"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, false, tools["listChanged"])
}

func TestRun_InitializeWithoutParams(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.Nil(t, responses[0].Error)

	result := resultMap(t, responses[0])
	require.Equal(t, ProtocolVersion, result["protocolVersion"])
}

func TestRun_InitializeWithInvalidParams(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":"invalid"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.NotNil(t, responses[0].Error)
	require.Equal(t, jsonrpc.InvalidParams, responses[0].Error.Code)
}

func TestRun_Ping(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.Nil(t, responses[0].Error)
	require.Equal(t, float64(1), responses[0].ID)
}

func TestRun_ToolsList(t *testing.T) {
	input := initHandshake() +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	// Register two tools to verify sorting.
	s.RegisterTool(ToolDefinition{
		Name:        "zebra",
		Description: "Z tool.",
		InputSchema: map[string]any{"type": "object"},
	}, func(ctx context.Context, args json.RawMessage) (any, error) {
		return nil, nil
	})
	s.RegisterTool(ToolDefinition{
		Name:        "alpha",
		Description: "A tool.",
		InputSchema: map[string]any{"type": "object"},
	}, func(ctx context.Context, args json.RawMessage) (any, error) {
		return nil, nil
	})

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	// 1 = initialize response, 2 = tools/list response (notifications/initialized has no response).
	require.Len(t, responses, 2)

	toolsResp := responses[1]
	require.Nil(t, toolsResp.Error)

	result := resultMap(t, toolsResp)
	toolsList, ok := result["tools"].([]any)
	require.True(t, ok)
	require.Len(t, toolsList, 2)

	first := toolsList[0].(map[string]any)
	second := toolsList[1].(map[string]any)
	require.Equal(t, "alpha", first["name"])
	require.Equal(t, "zebra", second["name"])
}

func TestRun_ToolsCall_Success(t *testing.T) {
	input := initHandshake() +
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"msg":"hi"}}}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())
	registerEchoTool(s)

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 2) // initialize + tools/call

	callResp := responses[1]
	require.Nil(t, callResp.Error)
	require.Equal(t, float64(3), callResp.ID)

	result := resultMap(t, callResp)
	require.Equal(t, false, result["isError"])
	require.NotNil(t, result["content"])
	require.NotNil(t, result["structuredContent"])
}

func TestRun_ToolsCall_HandlerError(t *testing.T) {
	input := initHandshake() +
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"fail","arguments":{}}}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())
	s.RegisterTool(ToolDefinition{
		Name:        "fail",
		InputSchema: map[string]any{"type": "object"},
	}, func(ctx context.Context, args json.RawMessage) (any, error) {
		return nil, errors.New("tool failed")
	})

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 2)

	callResp := responses[1]
	// Handler errors are returned as isError in the result, not as a JSON-RPC error.
	require.Nil(t, callResp.Error)
	result := resultMap(t, callResp)
	require.Equal(t, true, result["isError"])

	content := result["content"].([]any)
	first := content[0].(map[string]any)
	require.Equal(t, "tool failed", first["text"])
}

func TestRun_ToolsCall_InvalidParams(t *testing.T) {
	input := initHandshake() +
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":"bad"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 2)

	require.NotNil(t, responses[1].Error)
	require.Equal(t, jsonrpc.InvalidParams, responses[1].Error.Code)
}

func TestRun_ToolsCall_UnknownTool(t *testing.T) {
	input := initHandshake() +
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"nope","arguments":{}}}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 2)

	require.NotNil(t, responses[1].Error)
	require.Equal(t, jsonrpc.InvalidParams, responses[1].Error.Code)
	require.Contains(t, responses[1].Error.Message, "unknown tool")
}

func TestRun_ToolsCall_MarshalError(t *testing.T) {
	input := initHandshake() +
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"bad","arguments":{}}}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())
	s.RegisterTool(ToolDefinition{
		Name:        "bad",
		InputSchema: map[string]any{"type": "object"},
	}, func(ctx context.Context, args json.RawMessage) (any, error) {
		// Channels cannot be marshaled to JSON.
		return make(chan int), nil
	})

	err := s.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to indent json response")
}

func TestRun_ToolsListBeforeInit(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.NotNil(t, responses[0].Error)
	require.Equal(t, jsonrpc.InvalidRequest, responses[0].Error.Code)
	require.Contains(t, responses[0].Error.Message, "not received notifications/initialized")
}

func TestRun_ToolsCallBeforeInit(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{}}}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())
	registerEchoTool(s)

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.NotNil(t, responses[0].Error)
	require.Equal(t, jsonrpc.InvalidRequest, responses[0].Error.Code)
}

func TestRun_InvalidJSON(t *testing.T) {
	input := "not json\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.NotNil(t, responses[0].Error)
	require.Equal(t, jsonrpc.ParseError, responses[0].Error.Code)
}

func TestRun_InvalidJSONRPCVersion(t *testing.T) {
	input := `{"jsonrpc":"1.0","id":1,"method":"ping"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.NotNil(t, responses[0].Error)
	require.Equal(t, jsonrpc.InvalidRequest, responses[0].Error.Code)
	require.Contains(t, responses[0].Error.Message, "jsonrpc must be 2.0")
}

func TestRun_UnknownMethodWithID(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"unknown/method"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.NotNil(t, responses[0].Error)
	require.Equal(t, jsonrpc.MethodNotFound, responses[0].Error.Code)
}

func TestRun_UnknownNotification(t *testing.T) {
	// Notifications have no ID — the server should silently ignore unknown ones.
	input := `{"jsonrpc":"2.0","method":"unknown/notification"}` + "\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 0)
}

func TestRun_EmptyLines(t *testing.T) {
	input := "\n\n" + `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n\n"
	out := &bytes.Buffer{}
	s := New(strings.NewReader(input), out, discardLogger())

	err := s.Run(context.Background())
	require.NoError(t, err)

	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.Nil(t, responses[0].Error)
}

func TestRun_ContextCanceled(t *testing.T) {
	r, w := io.Pipe()
	defer w.Close()
	out := &bytes.Buffer{}
	ctx, cancel := context.WithCancel(context.Background())

	s := New(r, out, discardLogger())

	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	cancel()

	err := <-done
	require.ErrorIs(t, err, context.Canceled)
}

func TestRun_ScannerError(t *testing.T) {
	// Provide a valid line first, then an errReader to trigger scanner error.
	input := io.MultiReader(
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`+"\n"),
		&errReader{},
	)
	out := &bytes.Buffer{}
	s := New(input, out, discardLogger())

	err := s.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading input")

	// The ping response should still have been written before the error.
	responses := parseResponses(t, out)
	require.Len(t, responses, 1)
	require.Nil(t, responses[0].Error)
}

func TestRun_WriteError_ParseError(t *testing.T) {
	input := "not json\n"
	s := New(strings.NewReader(input), &errWriter{}, discardLogger())

	err := s.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal request")
}

func TestRun_WriteError_InvalidVersion(t *testing.T) {
	input := `{"jsonrpc":"1.0","id":1,"method":"ping"}` + "\n"
	s := New(strings.NewReader(input), &errWriter{}, discardLogger())

	err := s.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid jsonrpc version")
}

func TestRun_WriteError_RequireInitialized(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	s := New(strings.NewReader(input), &errWriter{}, discardLogger())

	err := s.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to handle request")
}

func TestRun_WriteError_FailedToMarshalJSON(t *testing.T) {
	originalJsonMarshal := jsonMarshal
	defer func() { jsonMarshal = originalJsonMarshal }()
	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("marshal error")
	}
	input := `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"
	s := New(strings.NewReader(input), &errWriter{}, discardLogger())

	err := s.Run(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "marshal error")
}
