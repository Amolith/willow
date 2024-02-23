// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"sync"
)

// DeleteProject deletes a project from the database
func DeleteProject(db *sql.DB, mu *sync.Mutex, id string) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := db.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM releases WHERE project_id = ?", id)
	return err
}

// GetProject returns a project from the database
func GetProject(db *sql.DB, url string) (map[string]string, error) {
	var id, name, forge, version string
	err := db.QueryRow("SELECT id, name, forge, version FROM projects WHERE url = ?", url).Scan(&id, &name, &forge, &version)
	if err != nil {
		return nil, err
	}
	project := map[string]string{
		"id":      id,
		"name":    name,
		"url":     url,
		"forge":   forge,
		"version": version,
	}
	return project, nil
}

// UpsertProject adds or updates a project in the database
func UpsertProject(db *sql.DB, mu *sync.Mutex, id, url, name, forge, running string) error {
	mu.Lock()
	defer mu.Unlock()
	_, err := db.Exec(`INSERT INTO projects (id, url, name, forge, version)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO 
			UPDATE SET
				name = excluded.name,
				forge = excluded.forge,
				version = excluded.version;`, id, url, name, forge, running)
	return err
}

// GetProjects returns a list of all projects in the database
func GetProjects(db *sql.DB) ([]map[string]string, error) {
	rows, err := db.Query("SELECT id, name, url, forge, version FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []map[string]string
	for rows.Next() {
		var id, name, url, forge, version string
		err = rows.Scan(&id, &name, &url, &forge, &version)
		if err != nil {
			return nil, err
		}
		project := map[string]string{
			"id":      id,
			"name":    name,
			"url":     url,
			"forge":   forge,
			"version": version,
		}
		projects = append(projects, project)
	}
	return projects, nil
}
