// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"sync"
)

// UpsertRelease adds or updates a release for a project with a given ID in the
// database
func UpsertRelease(db *sql.DB, mu *sync.Mutex, id, projectID, url, tag, content, date string) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := db.Exec(`INSERT INTO releases (id, project_id, url, tag, content, date)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO 
			UPDATE SET
				url = excluded.url,
				content = excluded.content,
				tag = excluded.tag,
				content = excluded.content,
				date = excluded.date;`, id, projectID, url, tag, content, date)
	return err
}

// GetReleases returns all releases for a project with a given id from the database
func GetReleases(db *sql.DB, projectID string) ([]map[string]string, error) {
	rows, err := db.Query(`SELECT id, url, tag, content, date FROM releases WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	releases := make([]map[string]string, 0)
	for rows.Next() {
		var (
			id      string
			url     string
			tag     string
			content string
			date    string
		)
		err := rows.Scan(&id, &url, &tag, &content, &date)
		if err != nil {
			return nil, err
		}
		releases = append(releases, map[string]string{
			"id":         id,
			"project_id": projectID,
			"url":        url,
			"tag":        tag,
			"content":    content,
			"date":       date,
		})
	}
	return releases, nil
}
