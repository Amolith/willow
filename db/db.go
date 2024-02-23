// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"sync"

	_ "modernc.org/sqlite"
)

//go:embed sql/schema.sql
var schema string

var mutex = &sync.Mutex{}

// Open opens a connection to the SQLite database
func Open(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite", "file:"+dbPath+"?_pragma=journal_mode%3DWAL")
}

// VerifySchema checks whether the schema has been initalised and initialises it
// if not
func InitialiseDatabase(dbConn *sql.DB) error {
	var name string
	err := dbConn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&name)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		mutex.Lock()
		defer mutex.Unlock()
		if _, err := dbConn.Exec(schema); err != nil {
			return err
		}
		return nil
	}
	return err
}
