// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import "database/sql"

// CreateProject adds a project to the database
func CreateProject(db *sql.DB, url, name, forge, running string) error {
	_, err := db.Exec("INSERT INTO projects (url, name, forge, version) VALUES (?, ?, ?, ?)", url, name, forge, running)
	return err
}

// DeleteProject deletes a project from the database
func DeleteProject(db *sql.DB, url string) error {
	_, err := db.Exec("DELETE FROM projects WHERE url = ?", url)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM releases WHERE project_url = ?", url)
	return err
}

// GetProject returns a project from the database
func GetProject(db *sql.DB, url string) (map[string]string, error) {
	var name, forge, version string
	err := db.QueryRow("SELECT name, forge, version FROM projects WHERE url = ?", url).Scan(&name, &forge, &version)
	if err != nil {
		return nil, err
	}
	project := map[string]string{
		"name":    name,
		"url":     url,
		"forge":   forge,
		"version": version,
	}
	return project, nil
}

// UpdateProject updates an existing project in the database
func UpdateProject(db *sql.DB, url, name, forge, running string) error {
	_, err := db.Exec("UPDATE projects SET name=?, forge=?, version=? WHERE url=?", name, forge, running, url)
	return err
}

// UpsertProject adds or updates a project in the database
func UpsertProject(db *sql.DB, url, name, forge, running string) error {
	_, err := db.Exec(`INSERT INTO projects (url, name, forge, version)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(url) DO 
			UPDATE SET
				name = excluded.name,
				forge = excluded.forge,
				version = excluded.version;`, url, name, forge, running)
	return err
}

// GetProjects returns a list of all projects in the database
func GetProjects(db *sql.DB) ([]map[string]string, error) {
	rows, err := db.Query("SELECT name, url, forge, version FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []map[string]string
	for rows.Next() {
		var name, url, forge, version string
		err = rows.Scan(&name, &url, &forge, &version)
		if err != nil {
			return nil, err
		}
		project := map[string]string{
			"name":    name,
			"url":     url,
			"forge":   forge,
			"version": version,
		}
		projects = append(projects, project)
	}
	return projects, nil
}
