// Copyright (c) 2026 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHello(t *testing.T) {
	t.Run("should return a personalized greeting message", func(t *testing.T) {
		result, err := Hello(HelloArgs{Name: "Tiago"})
		require.NoError(t, err)

		expected := "Hello, Tiago"
		require.Equal(t, expected, result.Message)
	})

	t.Run("should return a default greeting message when name is empty", func(t *testing.T) {
		result, err := Hello(HelloArgs{Name: "   "})
		require.NoError(t, err)

		expected := "Hello, world"
		require.Equal(t, expected, result.Message)
	})
}
