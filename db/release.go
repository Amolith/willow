// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
)

// AddRelease adds a release for a project with a given URL to the database

// DeleteRelease deletes a release for a project with a given URL from the database

// UpdateRelease updates a release for a project with a given URL in the database

// UpsertRelease adds or updates a release for a project with a given URL in the
// database
func UpsertRelease(db *sql.DB, id, projectURL, releaseURL, tag, content, date string) error {
	_, err := db.Exec(`INSERT INTO releases (id, project_url, release_url, tag, content, date)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO 
			UPDATE SET
				release_url = excluded.release_url,
				content = excluded.content,
				tag = excluded.tag,
				content = excluded.content,
				date = excluded.date;`, id, projectURL, releaseURL, tag, content, date)
	return err
}

// GetRelease returns a release for a project with a given URL from the database

// GetReleases returns all releases for a project with a given URL from the database
func GetReleases(db *sql.DB, projectURL string) ([]map[string]string, error) {
	rows, err := db.Query(`SELECT project_url, release_url, tag, content, date FROM releases WHERE project_url = ?`, projectURL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	releases := make([]map[string]string, 0)
	for rows.Next() {
		var (
			projectURL string
			releaseURL string
			tag        string
			content    string
			date       string
		)
		err := rows.Scan(&projectURL, &releaseURL, &tag, &content, &date)
		if err != nil {
			return nil, err
		}
		releases = append(releases, map[string]string{
			"projectURL": projectURL,
			"releaseURL": releaseURL,
			"tag":        tag,
			"content":    content,
			"date":       date,
		})
	}
	return releases, nil
}
