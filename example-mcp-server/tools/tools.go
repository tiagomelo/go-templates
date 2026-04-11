// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

// Package tools provides a set of example tools that can be registered with the MCP server.
// These tools demonstrate how to define tool metadata, input schemas, and handlers.
// You can use these as a starting point for creating your own custom tools.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tiagomelo/go-templates/example-mcp-server/server"
)

// RegisterDefaultTools registers the default tools
// with the provided server.
func RegisterDefaultTools(s *server.Server) {
	s.RegisterTool(
		server.ToolDefinition{
			Name:        "hello_world",
			Description: "Generate a greeting message for a given name. Use this when the user asks to greet someone, say hello, or produce a simple greeting.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Optional name to greet.",
					},
				},
			},
		},
		func(ctx context.Context, raw json.RawMessage) (any, error) {
			var args HelloArgs
			if len(raw) > 0 {
				if err := json.Unmarshal(raw, &args); err != nil {
					return nil, fmt.Errorf("decoding arguments: %w", err)
				}
			}
			return Hello(args)
		},
	)
}
