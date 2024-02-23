// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
)

// generateAndInsertProjectIDs runs during migration 1, fetches all rows from
// projects_tmp, loops through the rows generating a repeatable ID for each
// project, and inserting it into the new table along with the data from the old
// table.
func generateAndInsertProjectIDs(tx *sql.Tx) error {
	// Loop through projects_tmp, generate a project_id for each, and insert
	// into projects
	rows, err := tx.Query("SELECT url, name, forge, version, created_at FROM projects_tmp")
	if err != nil {
		return fmt.Errorf("failed to list projects in projects_tmp: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			url        string
			name       string
			forge      string
			version    string
			created_at string
		)
		if err := rows.Scan(&url, &name, &forge, &version, &created_at); err != nil {
			return fmt.Errorf("failed to scan row from projects_tmp: %w", err)
		}
		id := fmt.Sprintf("%x", sha256.Sum256([]byte(url+name+forge+created_at)))
		_, err = tx.Exec(
			"INSERT INTO projects (id, url, name, forge, version, created_at) VALUES (@id, @url, @name, @forge, @version, @created_at)",
			sql.Named("id", id),
			sql.Named("url", url),
			sql.Named("name", name),
			sql.Named("forge", forge),
			sql.Named("version", version),
			sql.Named("created_at", created_at),
		)
		if err != nil {
			return fmt.Errorf("failed to insert project into projects: %w", err)
		}
	}

	if _, err := tx.Exec("DROP TABLE projects_tmp"); err != nil {
		return fmt.Errorf("failed to drop projects_tmp: %w", err)
	}

	return nil
}

// Basing the project's ID on when it was created (L37) was a bad idea.
func correctProjectIDs(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT id, url, name, forge FROM projects")
	if err != nil {
		return fmt.Errorf("failed to list projects in projects_tmp: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			old_id string
			url    string
			name   string
			forge  string
		)
		if err := rows.Scan(&old_id, &url, &name, &forge); err != nil {
			return fmt.Errorf("failed to scan row from projects_tmp: %w", err)
		}
		id := fmt.Sprintf("%x", sha256.Sum256([]byte(url+name+forge)))
		_, err = tx.Exec(
			"UPDATE projects SET id = @id WHERE id = @old_id",
			sql.Named("id", id),
			sql.Named("old_id", old_id),
		)
		if err != nil {
			return fmt.Errorf("failed to insert project into projects: %w", err)
		}
	}

	return nil
}
