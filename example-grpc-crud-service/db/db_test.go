// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package db

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectToSqlite(t *testing.T) {
	testCases := []struct {
		name          string
		mockSqlOpen   func(driverName string, dataSourceName string) (*sql.DB, error)
		expectedError error
	}{
		{
			name: "happy path",
			mockSqlOpen: func(driverName, dataSourceName string) (*sql.DB, error) {
				return new(sql.DB), nil
			},
		},
		{
			name: "error",
			mockSqlOpen: func(driverName, dataSourceName string) (*sql.DB, error) {
				return nil, errors.New("open error")
			},
			expectedError: errors.New("opening sqlite file path/to/file.db: open error"),
		},
	}
	for _, tc := range testCases {
		sqlOpen = tc.mockSqlOpen
		t.Run(tc.name, func(t *testing.T) {
			db, err := ConnectToSqlite("path/to/file.db")
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.NotNil(t, db)
			}
		})
	}
}
