// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/unascribed/FlexVer/go/flexver"

	"git.sr.ht/~amolith/willow/db"
	"git.sr.ht/~amolith/willow/git"
	"git.sr.ht/~amolith/willow/rss"
)

type Project struct {
	ID       string
	URL      string
	Name     string
	Forge    string
	Running  string
	Releases []Release
}

type Release struct {
	ID        string
	ProjectID string
	URL       string
	Tag       string
	Content   string
	Date      time.Time
}

// GetReleases returns a list of all releases for a project from the database
func GetReleases(dbConn *sql.DB, mu *sync.Mutex, proj Project) (Project, error) {
	proj.ID = GenProjectID(proj.URL, proj.Name, proj.Forge)

	ret, err := db.GetReleases(dbConn, proj.ID)
	if err != nil {
		return proj, err
	}

	if len(ret) == 0 {
		proj, err = fetchReleases(dbConn, mu, proj)
		if err != nil {
			return proj, err
		}
		err = upsertReleases(dbConn, mu, proj.ID, proj.Releases)
		if err != nil {
			return proj, err
		}
		return proj, nil
	}

	for _, row := range ret {
		proj.Releases = append(proj.Releases, Release{
			ID:        row["id"],
			ProjectID: proj.ID,
			Tag:       row["tag"],
			Content:   row["content"],
			URL:       row["release_url"],
			Date:      time.Time{},
		})
	}
	proj.Releases = SortReleases(proj.Releases)
	return proj, nil
}

// fetchReleases fetches releases from a project's forge given its URI
func fetchReleases(dbConn *sql.DB, mu *sync.Mutex, p Project) (Project, error) {
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
				ID:      GenReleaseID(p.URL, release.URL, release.Tag),
				Tag:     release.Tag,
				Content: release.Content,
				URL:     release.URL,
				Date:    release.Date,
			})
			err = upsertReleases(dbConn, mu, p.ID, p.Releases)
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
				ID:      GenReleaseID(p.URL, release.URL, release.Tag),
				Tag:     release.Tag,
				Content: release.Content,
				URL:     release.URL,
				Date:    release.Date,
			})
			err = upsertReleases(dbConn, mu, p.ID, p.Releases)
			if err != nil {
				log.Printf("Error upserting release: %v", err)
				return p, err
			}
		}
	}
	p.Releases = SortReleases(p.Releases)
	return p, err
}

func SortReleases(releases []Release) []Release {
	sort.Slice(releases, func(i, j int) bool {
		return !flexver.Less(releases[i].Tag, releases[j].Tag)
	})
	return releases
}

func SortProjects(projects []Project) []Project {
	sort.Slice(projects, func(i, j int) bool {
		return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
	})
	return projects
}

// upsertReleases updates or inserts a release in the database
func upsertReleases(dbConn *sql.DB, mu *sync.Mutex, projID string, releases []Release) error {
	for _, release := range releases {
		date := release.Date.Format("2006-01-02 15:04:05")
		err := db.UpsertRelease(dbConn, mu, release.ID, projID, release.URL, release.Tag, release.Content, date)
		if err != nil {
			log.Printf("Error upserting release: %v", err)
			return err
		}
	}
	return nil
}

// GenReleaseID generates a likely-unique ID from its project's URL, its release's URL, and its tag
func GenReleaseID(projectURL, releaseURL, tag string) string {
	idByte := sha256.Sum256([]byte(projectURL + releaseURL + tag))
	return fmt.Sprintf("%x", idByte)
}

// GenProjectID generates a likely-unique ID from a project's URI, name, and forge
func GenProjectID(url, name, forge string) string {
	idByte := sha256.Sum256([]byte(url + name + forge))
	return fmt.Sprintf("%x", idByte)
}

func Track(dbConn *sql.DB, mu *sync.Mutex, manualRefresh *chan struct{}, name, url, forge, release string) {
	id := GenProjectID(url, name, forge)
	err := db.UpsertProject(dbConn, mu, id, url, name, forge, release)
	if err != nil {
		fmt.Println("Error upserting project:", err)
	}
	*manualRefresh <- struct{}{}
}

func Untrack(dbConn *sql.DB, mu *sync.Mutex, id string) {
	err := db.DeleteProject(dbConn, mu, id)
	if err != nil {
		fmt.Println("Error deleting project:", err)
	}

	err = git.RemoveRepo(id)
	if err != nil {
		log.Println(err)
	}
}

func RefreshLoop(dbConn *sql.DB, mu *sync.Mutex, interval int, manualRefresh, req *chan struct{}, res *chan []Project) {
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	fetch := func() []Project {
		projectsList, err := GetProjects(dbConn)
		if err != nil {
			fmt.Println("Error getting projects:", err)
		}
		for i, p := range projectsList {
			p, err := fetchReleases(dbConn, mu, p)
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
			err = upsertReleases(dbConn, mu, projectsList[i].ID, projectsList[i].Releases)
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
func GetProject(dbConn *sql.DB, proj Project) (Project, error) {
	projectDB, err := db.GetProject(dbConn, proj.ID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return proj, nil
	} else if err != nil {
		return proj, err
	}
	p := Project{
		ID:      proj.ID,
		URL:     proj.URL,
		Name:    proj.Name,
		Forge:   proj.Forge,
		Running: projectDB["version"],
	}
	return p, err
}

// GetProjectWithReleases returns a single project from the database along with its releases
func GetProjectWithReleases(dbConn *sql.DB, mu *sync.Mutex, proj Project) (Project, error) {
	project, err := GetProject(dbConn, proj)
	if err != nil {
		return Project{}, err
	}

	return GetReleases(dbConn, mu, project)
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
			ID:      p["id"],
			URL:     p["url"],
			Name:    p["name"],
			Forge:   p["forge"],
			Running: p["version"],
		}
	}

	return SortProjects(projects), nil
}

// GetProjectsWithReleases returns a list of all projects and all their releases
// from the database
func GetProjectsWithReleases(dbConn *sql.DB, mu *sync.Mutex) ([]Project, error) {
	projects, err := GetProjects(dbConn)
	if err != nil {
		return nil, err
	}

	for i := range projects {
		projects[i], err = GetReleases(dbConn, mu, projects[i])
		if err != nil {
			return nil, err
		}
		projects[i].Releases = SortReleases(projects[i].Releases)
	}

	return SortProjects(projects), nil
}
