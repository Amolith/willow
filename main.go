// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	flag "github.com/spf13/pflag"
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

	Config struct {
		Server      server
		CSVLocation string
		// TODO: Make cache location configurable
		// CacheLocation string
		FetchInterval int
	}

	server struct {
		Listen string
	}
)

var (
	flagConfig    *string = flag.StringP("config", "c", "config.toml", "Path to config file")
	config        Config
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

	flag.Parse()

	err := checkConfig()
	if err != nil {
		log.Fatalln(err)
	}

	err = checkCSV()
	if err != nil {
		log.Fatalln(err)
	}

	reader := csv.NewReader(strings.NewReader(config.CSVLocation))

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
		Addr:    config.Server.Listen,
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

func checkConfig() error {
	file, err := os.Open(*flagConfig)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(*flagConfig)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = file.WriteString("# Location of the CSV file containing the projects\nCSVLocation = \"projects.csv\"\n# How often to fetch new releases in seconds\nFetchInterval = 3600\n\n[Server]\n# Address to listen on\nListen = \"127.0.0.1:1313\"\n")
			if err != nil {
				return err
			}

			fmt.Println("Config file created at", *flagConfig)
			fmt.Println("Please edit it and restart the server")
			os.Exit(0)
		} else {
			return err
		}
	}
	defer file.Close()

	_, err = toml.DecodeFile(*flagConfig, &config)
	if err != nil {
		return err
	}

	if config.CSVLocation == "" {
		fmt.Println("No CSV location specified, using projects.csv")
		config.CSVLocation = "projects.csv"
	}

	if config.FetchInterval < 10 {
		fmt.Println("Fetch interval is set to", config.FetchInterval, "seconds, but the minimum is 10, using 10")
		config.FetchInterval = 10
	}

	if config.Server.Listen == "" {
		fmt.Println("No listen address specified, using 127.0.0.1:1313")
		config.Server.Listen = "127.0.0.1:1313"
	}

	return nil
}

func checkCSV() error {
	file, err := os.Open(config.CSVLocation)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(config.CSVLocation)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = file.WriteString("url,name,forge,running\nhttps://git.sr.ht/~amolith/earl,earl,sourcehut,v0.0.1-rc0\n")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer file.Close()
	return nil
}
