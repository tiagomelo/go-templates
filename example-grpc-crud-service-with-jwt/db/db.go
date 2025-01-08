// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// For ease of unit testing.
var sqlOpen = sql.Open

// ConnectToSqlite establishes a connection to a SQLite database.
func ConnectToSqlite(sqliteFilePath string) (*sql.DB, error) {
	db, err := sqlOpen("sqlite3", sqliteFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "opening sqlite file %s", sqliteFilePath)
	}
	return db, nil
}
