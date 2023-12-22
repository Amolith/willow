// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed sql/schema.sql
var schema string

// Open opens a connection to the SQLite database
func Open(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite", dbPath)
}

// VerifySchema checks whether the schema has been initalised and initialises it
// if not
func InitialiseDatabase(dbConn *sql.DB) error {
	var name string
	err := dbConn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&name)
	if err == nil {
		return nil
	}

	tables := []string{
		"users",
		"sessions",
		"projects",
		"releases",
	}

	for _, table := range tables {
		name := ""
		err := dbConn.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=@table",
			sql.Named("table", table),
		).Scan(&name)
		if err != nil {
			if err = loadSchema(dbConn); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadSchema loads the initial schema into the database
func loadSchema(dbConn *sql.DB) error {
	if _, err := dbConn.Exec(schema); err != nil {
		return err
	}
	return nil
}
