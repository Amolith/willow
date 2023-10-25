// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"embed"

	_ "modernc.org/sqlite"
)

// Embed the schema into the binary
//
//go:embed sql
var embeddedSQL embed.FS

// Open opens a connection to the SQLite database
func Open(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite", dbPath)
}

func VerifySchema(dbConn *sql.DB) error {
	tables := []string{
		"users",
		"sessions",
		"projects",
	}

	for _, table := range tables {
		name := ""
		err := dbConn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadSchema loads the schema into the database
func LoadSchema(dbConn *sql.DB) error {
	schema, err := embeddedSQL.ReadFile("sql/schema.sql")
	if err != nil {
		return err
	}

	_, err = dbConn.Exec(string(schema))

	return err
}
