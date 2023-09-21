package main

import (
	"encoding/csv"
	"log"
	"os"
)

func getReleases(p project) (project, error) {
	var err error
	switch p.Forge {
	case "github", "gitea", "forgejo":
		p, err = getRSSReleases(p)
	// case "gitlab":
	// 	// TODO: maybe use GitLab's API?
	default:
		p, err = getGitReleases(p)
	}
	return p, err
}

func track(name, url, forge, release string) {
	projectExists := false
	for i := range m.Projects {
		if m.Projects[i].URL == url {
			projectExists = true
			m.Projects[i].Running = release
		}
	}

	if !projectExists {
		m.Projects = append(m.Projects, project{
			URL:     url,
			Name:    name,
			Forge:   forge,
			Running: release,
		})
	}

	manualRefresh <- struct{}{}

	writeCSV()
}

func untrack(url string) {
	for i := range m.Projects {
		if m.Projects[i].URL == url {
			m.Projects = append(m.Projects[:i], m.Projects[i+1:]...)
			break
		}
	}

	manualRefresh <- struct{}{}

	writeCSV()
	err := removeRepo(url)
	if err != nil {
		log.Println(err)
	}
}

func writeCSV() {
	file, err := os.OpenFile("projects.csv", os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	if err := writer.Write([]string{"url", "name", "forge", "running"}); err != nil {
		log.Fatalln(err)
	}
	for _, project := range m.Projects {
		if err := writer.Write([]string{project.URL, project.Name, project.Forge, project.Running}); err != nil {
			log.Fatalln(err)
		}
	}

	writer.Flush()
}
