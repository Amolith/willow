// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

type (
	Model struct {
		Projects []project
	}

	project struct {
		URL      string
		Name     string
		Forge    string
		Running  string
		Releases []release
	}

	release struct {
		Tag     string
		Content string
		URL     string
		Date    time.Time
	}
)

var (
	req           = make(chan struct{})
	manualRefresh = make(chan struct{})
	res           = make(chan []project)
	m             = Model{
		Projects: []project{},
	}
	bmUGC    = bluemonday.UGCPolicy()
	bmStrict = bluemonday.StrictPolicy()
)

func main() {
	file, err := os.Open("projects.csv")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	m.Projects = []project{}
	if len(records) > 0 {
		for i, record := range records {
			if i == 0 {
				continue
			}
			m.Projects = append(m.Projects, project{
				URL:      record[0],
				Name:     record[1],
				Forge:    record[2],
				Running:  record[3],
				Releases: []release{},
			})
		}
	}

	go refreshLoop(manualRefresh, req, res)

	mux := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    "0.0.0.0:1337",
		Handler: mux,
	}

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/static", staticHandler)
	mux.HandleFunc("/new", newHandler)

	if err := httpServer.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		log.Println("Web server closed")
	} else {
		log.Fatalln(err)
	}
}

func refreshLoop(manualRefresh, req chan struct{}, res chan []project) {
	ticker := time.NewTicker(time.Second * 3600)

	fetch := func() []project {
		projects := make([]project, len(m.Projects))
		copy(projects, m.Projects)
		for i, project := range projects {
			project, err := getReleases(project)
			if err != nil {
				fmt.Println(err)
				continue
			}
			projects[i] = project
		}
		sort.Slice(projects, func(i, j int) bool { return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name) })
		return projects
	}

	projects := fetch()

	for {
		select {
		case <-ticker.C:
			projects = fetch()
		case <-manualRefresh:
			ticker.Reset(time.Second * 3600)
			projects = fetch()
		case <-req:
			projectsCopy := make([]project, len(projects))
			copy(projectsCopy, projects)
			res <- projectsCopy
		}
	}
}
