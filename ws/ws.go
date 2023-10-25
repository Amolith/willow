// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"text/template"

	"git.sr.ht/~amolith/willow/project"
	"github.com/microcosm-cc/bluemonday"
)

type Handler struct {
	DbConn        *sql.DB
	Mutex         *sync.Mutex
	Req           *chan struct{}
	ManualRefresh *chan struct{}
	Res           *chan []project.Project
}

//go:embed static
var fs embed.FS

// bmUGC    = bluemonday.UGCPolicy()
var bmStrict = bluemonday.StrictPolicy()

func (h Handler) RootHandler(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorised(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	*h.Req <- struct{}{}
	data := <-*h.Res
	tmpl := template.Must(template.ParseFS(fs, "static/home.html"))
	if err := tmpl.Execute(w, data); err != nil {
		fmt.Println(err)
	}
}

func (h Handler) NewHandler(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorised(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	params := r.URL.Query()
	action := bmStrict.Sanitize(params.Get("action"))
	if r.Method == http.MethodGet {
		if action == "" {
			tmpl := template.Must(template.ParseFS(fs, "static/new.html"))
			if err := tmpl.Execute(w, nil); err != nil {
				fmt.Println(err)
			}
		} else if action != "delete" {
			submittedURL := bmStrict.Sanitize(params.Get("url"))
			if submittedURL == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("No URL provided"))
				if err != nil {
					fmt.Println(err)
				}
				return
			}

			forge := bmStrict.Sanitize(params.Get("forge"))
			if forge == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("No forge provided"))
				if err != nil {
					fmt.Println(err)
				}
				return
			}

			name := bmStrict.Sanitize(params.Get("name"))
			if name == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("No name provided"))
				if err != nil {
					fmt.Println(err)
				}
			}

			proj := project.Project{
				URL:   submittedURL,
				Name:  name,
				Forge: forge,
			}
			proj, err := project.GetReleases(h.DbConn, proj)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte(fmt.Sprintf("Error getting releases: %s", err)))
				if err != nil {
					fmt.Println(err)
				}
			}
			tmpl := template.Must(template.ParseFS(fs, "static/select-release.html"))
			if err := tmpl.Execute(w, proj); err != nil {
				fmt.Println(err)
			}
		} else if action == "delete" {
			submittedURL := params.Get("url")
			if submittedURL == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("No URL provided"))
				if err != nil {
					fmt.Println(err)
				}
			}

			project.Untrack(h.DbConn, h.ManualRefresh, submittedURL)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
		nameValue := bmStrict.Sanitize(r.FormValue("name"))
		urlValue := bmStrict.Sanitize(r.FormValue("url"))
		forgeValue := bmStrict.Sanitize(r.FormValue("forge"))
		releaseValue := bmStrict.Sanitize(r.FormValue("release"))

		if nameValue != "" && urlValue != "" && forgeValue != "" && releaseValue != "" {
			project.Track(h.DbConn, h.ManualRefresh, nameValue, urlValue, forgeValue, releaseValue)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if nameValue != "" && urlValue != "" && forgeValue != "" && releaseValue == "" {
			http.Redirect(w, r, "/new?action=yoink&name="+url.QueryEscape(nameValue)+"&url="+url.QueryEscape(urlValue)+"&forge="+url.QueryEscape(forgeValue), http.StatusSeeOther)
			return
		}

		if nameValue == "" && urlValue == "" && forgeValue == "" && releaseValue == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("No data provided"))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (h Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: do this
}

func (h Handler) isAuthorised(r *http.Request) bool {
	// TODO: do this
	return false
}

func StaticHandler(writer http.ResponseWriter, request *http.Request) {
	resource := strings.TrimPrefix(request.URL.Path, "/")
	// if path ends in .css, set content type to text/css
	if strings.HasSuffix(resource, ".css") {
		writer.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(resource, ".js") {
		writer.Header().Set("Content-Type", "text/javascript")
	}
	home, err := fs.ReadFile(resource)
	if err != nil {
		fmt.Println(err)
	}
	if _, err = io.WriteString(writer, string(home)); err != nil {
		fmt.Println(err)
	}
}
