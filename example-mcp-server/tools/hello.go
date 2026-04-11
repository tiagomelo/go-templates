// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package tools

import (
	"fmt"
	"strings"
)

// HelloArgs defines the arguments for the Hello tool.
type HelloArgs struct {
	Name string `json:"name"`
}

// HelloResult defines the result of the Hello tool.
type HelloResult struct {
	Message string `json:"message"`
}

// Hello greets the user with a personalized message.
func Hello(args HelloArgs) (HelloResult, error) {
	name := strings.TrimSpace(args.Name)
	if name == "" {
		name = "world"
	}

	return HelloResult{
		Message: fmt.Sprintf("Hello, %s", name),
	}, nil
}
