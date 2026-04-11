// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

// Package server implements a JSON-RPC 2.0 server that follows the MCP protocol specification.
// It allows registering tools with their definitions and handlers, and processes incoming requests
// to list and call those tools. The server also handles the initialization handshake and provides
// logging of all interactions.
//
// See https://modelcontextprotocol.io/specification/2025-11-25 for the MCP protocol specification.
package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/tiagomelo/go-templates/example-mcp-server/jsonrpc"
)

// ProtocolVersion is the version of the MCP protocol implemented by this server.
const ProtocolVersion = "2025-06-18"

// jsonMarshal is a variable to allow overriding json.Marshal in tests.
var jsonMarshal = json.Marshal

// ToolDefinition defines a tool that can be called by the client.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ToolHandler is a function that implements the logic of a tool.
// It receives the tool arguments as raw JSON and returns any result or an error.
type ToolHandler func(ctx context.Context, arguments json.RawMessage) (any, error)

// Server implements the MCP protocol over JSON-RPC 2.0.
type Server struct {
	in  io.Reader
	out io.Writer

	mu          sync.RWMutex
	tools       map[string]ToolHandler
	definitions map[string]ToolDefinition
	initialized bool
	handlers    map[string]func(context.Context, jsonrpc.Request) error
	logger      *slog.Logger
}

// New creates a new MCP server that reads requests from in and writes responses to out.
func New(in io.Reader, out io.Writer, logger *slog.Logger) *Server {
	s := &Server{
		in:          in,
		out:         out,
		tools:       make(map[string]ToolHandler),
		definitions: make(map[string]ToolDefinition),
		logger:      logger,
	}

	s.handlers = map[string]func(context.Context, jsonrpc.Request) error{
		"initialize":                s.handleInitialize,
		"notifications/initialized": s.handleInitializedNotification,
		"ping":                      s.handlePing,
		"tools/list":                s.handleToolsList,
		"tools/call":                s.handleToolsCall,
	}

	return s
}

// RegisterTool registers a tool with the server.
// It can be called at any time before or after initialization.
func (s *Server) RegisterTool(def ToolDefinition, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.definitions[def.Name] = def
	s.tools[def.Name] = handler
}

// ListToolNames returns a sorted list of registered tool names.
func (s *Server) ListToolNames() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.definitions))
	for name := range s.definitions {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Run starts the server and processes incoming requests until
// the context is canceled or an error occurs.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("mcp server started", slog.String("protocolVersion", ProtocolVersion))
	defer s.logger.Info("mcp server stopped")

	scanner := bufio.NewScanner(s.in)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	// scanCh delivers lines from stdin so we can select on ctx.Done().
	scanCh := make(chan []byte)
	scanErr := make(chan error, 1)
	go func() {
		defer close(scanCh)
		for scanner.Scan() {
			line := make([]byte, len(scanner.Bytes()))
			copy(line, scanner.Bytes())
			scanCh <- line
		}
		scanErr <- scanner.Err()
	}()

	for {
		var line []byte
		select {
		case <-ctx.Done():
			return ctx.Err()
		case l, ok := <-scanCh:
			if !ok {
				if err := <-scanErr; err != nil {
					return fmt.Errorf("reading input: %w", err)
				}
				return nil
			}
			line = l
		}

		if len(line) == 0 {
			continue
		}

		var req jsonrpc.Request
		if err := json.Unmarshal(line, &req); err != nil {
			if err := s.writeResponse(jsonrpc.Response{
				JSONRPC: "2.0",
				Error: &jsonrpc.Error{
					Code:    jsonrpc.ParseError,
					Message: fmt.Sprintf("parse error: %v", err),
				},
			}); err != nil {
				return errors.WithMessage(err, "failed to unmarshal request")
			}
			continue
		}

		if req.JSONRPC != "2.0" {
			if err := s.writeResponse(jsonrpc.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &jsonrpc.Error{
					Code:    jsonrpc.InvalidRequest,
					Message: "jsonrpc must be 2.0",
				},
			}); err != nil {
				return fmt.Errorf("invalid jsonrpc version: %s", req.JSONRPC)
			}
			continue
		}

		if err := s.handleRequest(ctx, req); err != nil {
			if errors.Is(err, errNotInitialized) {
				continue
			}
			return errors.WithMessage(err, "failed to handle request")
		}
	}
}

// handleRequest dispatches the request to the appropriate handler based on the method.
func (s *Server) handleRequest(ctx context.Context, req jsonrpc.Request) error {
	handler, ok := s.handlers[req.Method]
	if ok {
		return handler(ctx, req)
	}

	if req.ID == nil {
		s.logger.Debug("ignoring unknown notification", slog.String("method", req.Method))
		return nil
	}

	s.logger.Warn("unknown method",
		slog.String("method", req.Method),
		slog.Any("id", req.ID),
	)

	return s.writeResponse(jsonrpc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Error: &jsonrpc.Error{
			Code:    jsonrpc.MethodNotFound,
			Message: fmt.Sprintf("method not found: %s", req.Method),
		},
	})
}

// handleInitialize processes the initialize request and
// responds with the server capabilities and instructions.
func (s *Server) handleInitialize(ctx context.Context, req jsonrpc.Request) error {
	s.logger.Info("initialize request received", slog.Any("id", req.ID))

	type initializeParams struct {
		ProtocolVersion string         `json:"protocolVersion"`
		Capabilities    map[string]any `json:"capabilities"`
		ClientInfo      map[string]any `json:"clientInfo"`
	}

	var params initializeParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.writeResponse(jsonrpc.Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &jsonrpc.Error{
					Code:    jsonrpc.InvalidParams,
					Message: fmt.Sprintf("invalid initialize params: %v", err),
				},
			})
		}
	}

	return s.writeResponse(jsonrpc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": ProtocolVersion,
			"capabilities": map[string]any{
				"tools": map[string]any{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]any{
				"name":    "go-mcp-server",
				"version": "0.1.0",
			},
			"instructions": fmt.Sprintf("This MCP server provides the following tools: [%s]", strings.Join(s.ListToolNames(), ", ")),
		},
	})
}

// handleInitializedNotification processes the notifications/initialized notification
// and marks the server as initialized.
func (s *Server) handleInitializedNotification(ctx context.Context, req jsonrpc.Request) error {
	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()
	s.logger.Info("initialized notification received")
	return nil
}

// handlePing processes the ping request and responds with an empty result.
func (s *Server) handlePing(ctx context.Context, req jsonrpc.Request) error {
	return s.writeResponse(jsonrpc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{},
	})
}

// handleToolsList processes the tools/list request and responds with the list of registered tools.
func (s *Server) handleToolsList(ctx context.Context, req jsonrpc.Request) error {
	if err := s.requireInitialized(req); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	defs := make([]ToolDefinition, 0, len(s.definitions))
	for _, def := range s.definitions {
		defs = append(defs, def)
	}

	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Name < defs[j].Name
	})

	s.logger.Info("tools listed",
		slog.Any("id", req.ID),
		slog.Int("count", len(defs)),
	)

	return s.writeResponse(jsonrpc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"tools": defs,
		},
	})
}

// handleToolsCall processes the tools/call request, executes the corresponding tool handler,
// and responds with the tool result or an error.
func (s *Server) handleToolsCall(ctx context.Context, req jsonrpc.Request) error {
	if err := s.requireInitialized(req); err != nil {
		return err
	}

	type callParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	var params callParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.writeResponse(jsonrpc.Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &jsonrpc.Error{
				Code:    jsonrpc.InvalidParams,
				Message: fmt.Sprintf("invalid tools/call params: %v", err),
			},
		})
	}

	s.logger.Info("tool call started",
		slog.Any("id", req.ID),
		slog.String("tool", params.Name),
	)

	s.mu.RLock()
	handler, ok := s.tools[params.Name]
	s.mu.RUnlock()

	if !ok {
		s.logger.Warn("unknown tool",
			slog.Any("id", req.ID),
			slog.String("tool", params.Name),
		)

		return s.writeResponse(jsonrpc.Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &jsonrpc.Error{
				Code:    jsonrpc.InvalidParams,
				Message: fmt.Sprintf("unknown tool: %s", params.Name),
			},
		})
	}

	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := handler(callCtx, params.Arguments)
	if err != nil {
		s.logger.Error("tool call failed",
			slog.Any("id", req.ID),
			slog.String("tool", params.Name),
			slog.String("error", err.Error()),
		)

		return s.writeResponse(jsonrpc.Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": err.Error(),
					},
				},
				"isError": true,
			},
		})
	}

	pretty, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.WithMessage(err, "failed to indent json response")
	}

	s.logger.Info("tool call succeeded",
		slog.Any("id", req.ID),
		slog.String("tool", params.Name),
	)

	return s.writeResponse(jsonrpc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": string(pretty),
				},
			},
			"structuredContent": result,
			"isError":           false,
		},
	})
}

// writeResponse marshals the response to JSON and writes it to the output.
func (s *Server) writeResponse(resp jsonrpc.Response) error {
	b, err := jsonMarshal(resp)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal json response")
	}

	if _, err := fmt.Fprintf(s.out, "%s\n", b); err != nil {
		return errors.WithMessage(err, "failed to write response")
	}

	return nil
}

// errNotInitialized is returned by requireInitialized when the server
// has not yet completed the initialization handshake.
// It is not a fatal error — the error response has already been written to the client.
var errNotInitialized = errors.New("server not initialized")

// requireInitialized checks if the server is initialized and
// if not, writes an error response and returns errNotInitialized.
func (s *Server) requireInitialized(req jsonrpc.Request) error {
	s.mu.RLock()
	initialized := s.initialized
	s.mu.RUnlock()

	if initialized {
		return nil
	}

	s.logger.Warn("request rejected before initialization",
		slog.String("method", req.Method),
		slog.Any("id", req.ID),
	)

	if err := s.writeResponse(jsonrpc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Error: &jsonrpc.Error{
			Code:    jsonrpc.InvalidRequest,
			Message: "server has not received notifications/initialized yet",
		},
	}); err != nil {
		return err
	}

	return errNotInitialized
}
