// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"git.sr.ht/~amolith/willow/db"
	"git.sr.ht/~amolith/willow/project"
	"git.sr.ht/~amolith/willow/ws"

	"github.com/BurntSushi/toml"
	flag "github.com/spf13/pflag"
)

type (
	Config struct {
		Server      server
		CSVLocation string
		DBConn      string
		// TODO: Make cache location configurable
		// CacheLocation string
		FetchInterval int
	}

	server struct {
		Listen string
	}
)

var (
	flagConfig          = flag.StringP("config", "c", "config.toml", "Path to config file")
	flagAddUser         = flag.StringP("add", "a", "", "Username of account to add")
	flagDeleteUser      = flag.StringP("deleteuser", "d", "", "Username of account to delete")
	flagCheckAuthorised = flag.StringP("validatecredentials", "v", "", "Username of account to check")
	flagListUsers       = flag.BoolP("listusers", "l", false, "List all users")
	config              Config
	req                 = make(chan struct{})
	res                 = make(chan []project.Project)
	manualRefresh       = make(chan struct{})
)

func main() {
	flag.Parse()

	err := checkConfig()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Opening database at", config.DBConn)

	dbConn, err := db.Open(config.DBConn)
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	fmt.Println("Checking whether database needs initialising")
	err = db.InitialiseDatabase(dbConn)
	if err != nil {
		fmt.Println("Error initialising database:", err)
		os.Exit(1)
	}
	fmt.Println("Checking whether there are pending migrations")
	err = db.Migrate(dbConn)
	if err != nil {
		fmt.Println("Error migrating database schema:", err)
		os.Exit(1)
	}

	if len(*flagAddUser) > 0 && len(*flagDeleteUser) == 0 && !*flagListUsers && len(*flagCheckAuthorised) == 0 {
		createUser(dbConn, *flagAddUser)
		os.Exit(0)
	} else if len(*flagAddUser) == 0 && len(*flagDeleteUser) > 0 && !*flagListUsers && len(*flagCheckAuthorised) == 0 {
		deleteUser(dbConn, *flagDeleteUser)
		os.Exit(0)
	} else if len(*flagAddUser) == 0 && len(*flagDeleteUser) == 0 && *flagListUsers && len(*flagCheckAuthorised) == 0 {
		listUsers(dbConn)
		os.Exit(0)
	} else if len(*flagAddUser) == 0 && len(*flagDeleteUser) == 0 && !*flagListUsers && len(*flagCheckAuthorised) > 0 {
		checkAuthorised(dbConn, *flagCheckAuthorised)
		os.Exit(0)
	}

	fmt.Println("Starting refresh loop")
	go project.RefreshLoop(dbConn, config.FetchInterval, &manualRefresh, &req, &res)

	var mutex sync.Mutex

	wsHandler := ws.Handler{
		DbConn:        dbConn,
		Mutex:         &mutex,
		Req:           &req,
		Res:           &res,
		ManualRefresh: &manualRefresh,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/static/", ws.StaticHandler)
	mux.HandleFunc("/new", wsHandler.NewHandler)
	mux.HandleFunc("/login", wsHandler.LoginHandler)
	mux.HandleFunc("/logout", wsHandler.LogoutHandler)
	mux.HandleFunc("/", wsHandler.RootHandler)

	httpServer := &http.Server{
		Addr:    config.Server.Listen,
		Handler: mux,
	}

	fmt.Println("Starting web server on", config.Server.Listen)
	if err := httpServer.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Web server closed")
		os.Exit(0)
	} else {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkConfig() error {
	defaultDBConn := "willow.sqlite"
	defaultFetchInterval := 3600
	defaultListen := "127.0.0.1:1313"

	defaultConfig := fmt.Sprintf(`# Path to SQLite database
DBConn = "%s"
# How often to fetch new releases in seconds
## Minimum is %ds to avoid rate limits and unintentional abuse
FetchInterval = %d

[Server]
# Address to listen on
Listen = "%s"`, defaultDBConn, defaultFetchInterval, defaultFetchInterval, defaultListen)

	file, err := os.Open(*flagConfig)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(*flagConfig)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = file.WriteString(defaultConfig)
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

	if config.FetchInterval < defaultFetchInterval {
		fmt.Println("Fetch interval is set to", strconv.Itoa(config.FetchInterval), "seconds, but the minimum is", defaultFetchInterval, "seconds, using", strconv.Itoa(defaultFetchInterval)+"s")
		config.FetchInterval = defaultFetchInterval
	}

	if config.Server.Listen == "" {
		fmt.Println("No listen address specified, using", defaultListen)
		config.Server.Listen = defaultListen
	}

	if config.DBConn == "" {
		fmt.Println("No SQLite path specified, using \"" + defaultDBConn + "\"")
		config.DBConn = defaultDBConn
	}

	return nil
}
