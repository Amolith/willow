// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"database/sql"
	"embed"
	"fmt"
	"git.sr.ht/~amolith/willow/users"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"text/template"
	"time"

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
				return
			}

			proj := project.Project{
				URL:   submittedURL,
				Name:  name,
				Forge: forge,
			}

			if strings.HasSuffix(proj.URL, ".git") {
				proj.URL = proj.URL[:len(proj.URL)-4]
			}

			proj, err := project.GetReleases(h.DbConn, proj)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte(fmt.Sprintf("Error getting releases: %s", err)))
				if err != nil {
					fmt.Println(err)
				}
				return
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
				return
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
	if r.Method == http.MethodGet {
		if h.isAuthorised(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		login, err := fs.ReadFile("static/login.html")
		if err != nil {
			fmt.Println("Error reading login.html:", err)
		}

		if _, err := io.WriteString(w, string(login)); err != nil {
			fmt.Println(err)
		}
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
		username := bmStrict.Sanitize(r.FormValue("username"))
		password := bmStrict.Sanitize(r.FormValue("password"))

		if username == "" || password == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("No data provided"))
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		authorised, err := users.UserAuthorised(h.DbConn, username, password)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(fmt.Sprintf("Error logging in: %s", err)))
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		if !authorised {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("Incorrect username or password"))
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		session, expiry, err := users.CreateSession(h.DbConn, username)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(fmt.Sprintf("Error creating session: %s", err)))
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		maxAge := int(expiry.Sub(time.Now()).Seconds())

		cookie := http.Cookie{
			Name:     "id",
			Value:    session,
			MaxAge:   maxAge,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   true,
		}

		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (h Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("id")
	if err != nil {
		fmt.Println(err)
	}

	err = users.InvalidateSession(h.DbConn, cookie.Value)
	if err != nil {
		fmt.Println(err)
		_, err = w.Write([]byte(fmt.Sprintf("Error logging out: %s", err)))
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// isAuthorised makes a database request to the sessions table to see if the
// user has a valid session cookie.
func (h Handler) isAuthorised(r *http.Request) bool {
	cookie, err := r.Cookie("id")
	if err != nil {
		return false
	}

	authorised, err := users.SessionAuthorised(h.DbConn, cookie.Value)
	if err != nil {
		fmt.Println("Error checking session:", err)
		return false
	}

	return authorised
}

func StaticHandler(writer http.ResponseWriter, request *http.Request) {
	resource := strings.TrimPrefix(request.URL.Path, "/")
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
