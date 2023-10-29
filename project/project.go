// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"git.sr.ht/~amolith/willow/db"
	"git.sr.ht/~amolith/willow/git"
	"git.sr.ht/~amolith/willow/rss"
)

type Project struct {
	URL      string
	Name     string
	Forge    string
	Running  string
	Releases []Release
}

type Release struct {
	URL     string
	Tag     string
	Content string
	Date    time.Time
}

// GetReleases returns a list of all releases for a project from the database
func GetReleases(dbConn *sql.DB, proj Project) (Project, error) {
	ret, err := db.GetReleases(dbConn, proj.URL)
	if err != nil {
		return proj, err
	}

	if len(ret) == 0 {
		return fetchReleases(dbConn, proj)
	}

	for _, row := range ret {
		proj.Releases = append(proj.Releases, Release{
			Tag:     row["tag"],
			Content: row["content"],
			URL:     row["release_url"],
			Date:    time.Time{},
		})
	}
	sort.Slice(proj.Releases, func(i, j int) bool {
		return proj.Releases[i].Date.After(proj.Releases[j].Date)
	})
	return proj, nil
}

// fetchReleases fetches releases from a project's forge given its URI
func fetchReleases(dbConn *sql.DB, p Project) (Project, error) {
	var err error
	switch p.Forge {
	case "github", "gitea", "forgejo":
		rssReleases, err := rss.GetReleases(p.URL)
		if err != nil {
			fmt.Println("Error getting RSS releases:", err)
			return p, err
		}
		for _, release := range rssReleases {
			p.Releases = append(p.Releases, Release{
				Tag:     release.Tag,
				Content: release.Content,
				URL:     release.URL,
				Date:    release.Date,
			})
			err = upsert(dbConn, p.URL, p.Releases)
			if err != nil {
				log.Printf("Error upserting release: %v", err)
				return p, err
			}
		}
	default:
		gitReleases, err := git.GetReleases(p.URL, p.Forge)
		if err != nil {
			return p, err
		}
		for _, release := range gitReleases {
			p.Releases = append(p.Releases, Release{
				Tag:     release.Tag,
				Content: release.Content,
				URL:     release.URL,
				Date:    release.Date,
			})
			err = upsert(dbConn, p.URL, p.Releases)
			if err != nil {
				log.Printf("Error upserting release: %v", err)
				return p, err
			}
		}
	}
	sort.Slice(p.Releases, func(i, j int) bool {
		return p.Releases[i].Date.After(p.Releases[j].Date)
	})
	return p, err
}

// upsert updates or inserts a project release into the database
func upsert(dbConn *sql.DB, url string, releases []Release) error {
	for _, release := range releases {
		date := release.Date.Format("2006-01-02 15:04:05")
		idByte := sha256.Sum256([]byte(url + release.URL + release.Tag + date))
		id := fmt.Sprintf("%x", idByte)
		err := db.UpsertRelease(dbConn, id, url, release.URL, release.Tag, release.Content, date)
		if err != nil {
			log.Printf("Error upserting release: %v", err)
			return err
		}
	}
	return nil
}

func Track(dbConn *sql.DB, manualRefresh *chan struct{}, name, url, forge, release string) {
	err := db.UpsertProject(dbConn, url, name, forge, release)
	if err != nil {
		fmt.Println("Error upserting project:", err)
	}
	*manualRefresh <- struct{}{}
}

func Untrack(dbConn *sql.DB, manualRefresh *chan struct{}, url string) {
	err := db.DeleteProject(dbConn, url)
	if err != nil {
		fmt.Println("Error deleting project:", err)
	}

	*manualRefresh <- struct{}{}

	err = git.RemoveRepo(url)
	if err != nil {
		log.Println(err)
	}
}

func RefreshLoop(dbConn *sql.DB, interval int, manualRefresh, req *chan struct{}, res *chan []Project) {
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	fetch := func() []Project {
		projectsList, err := GetProjects(dbConn)
		if err != nil {
			fmt.Println("Error getting projects:", err)
		}
		for i, p := range projectsList {
			p, err := fetchReleases(dbConn, p)
			if err != nil {
				fmt.Println(err)
				continue
			}
			projectsList[i] = p
		}
		sort.Slice(projectsList, func(i, j int) bool {
			return strings.ToLower(projectsList[i].Name) < strings.ToLower(projectsList[j].Name)
		})
		for i := range projectsList {
			err = upsert(dbConn, projectsList[i].URL, projectsList[i].Releases)
			if err != nil {
				fmt.Println("Error upserting release:", err)
				continue
			}
		}
		return projectsList
	}

	projects := fetch()

	for {
		select {
		case <-ticker.C:
			projects = fetch()
		case <-*manualRefresh:
			ticker.Reset(time.Second * 3600)
			projects = fetch()
		case <-*req:
			projectsCopy := make([]Project, len(projects))
			copy(projectsCopy, projects)
			*res <- projectsCopy
		}
	}
}

// GetProject returns a project from the database
func GetProject(dbConn *sql.DB, url string) (Project, error) {
	var p Project
	projectDB, err := db.GetProject(dbConn, url)
	if err != nil {
		return p, err
	}
	p = Project{
		URL:     projectDB["url"],
		Name:    projectDB["name"],
		Forge:   projectDB["forge"],
		Running: projectDB["version"],
	}
	return p, err
}

// GetProjects returns a list of all projects from the database
func GetProjects(dbConn *sql.DB) ([]Project, error) {
	projectsDB, err := db.GetProjects(dbConn)
	if err != nil {
		return nil, err
	}

	projects := make([]Project, len(projectsDB))
	for i, p := range projectsDB {
		projects[i] = Project{
			URL:     p["url"],
			Name:    p["name"],
			Forge:   p["forge"],
			Running: p["version"],
		}
	}

	return projects, nil
}
