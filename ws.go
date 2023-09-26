// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

//go:embed static
var fs embed.FS

func rootHandler(w http.ResponseWriter, r *http.Request) {
	req <- struct{}{}
	data := <-res
	tmpl := template.Must(template.ParseFS(fs, "static/home.html"))
	if err := tmpl.Execute(w, data); err != nil {
		fmt.Println(err)
	}
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	action := bmStrict.Sanitize(params.Get("action"))
	if r.Method == http.MethodGet {
		if action == "" {
			tmpl := template.Must(template.ParseFS(fs, "static/new.html"))
			if err := tmpl.Execute(w, nil); err != nil {
				fmt.Println(err)
			}
		} else if action != "delete" {
			submittedURL := bmStrict.Sanitize(params.Get("submittedURL"))
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

			proj := project{
				URL:   submittedURL,
				Name:  name,
				Forge: forge,
			}
			proj, err := getReleases(proj)
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
			submittedURL := params.Get("submittedURL")
			if submittedURL == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("No URL provided"))
				if err != nil {
					fmt.Println(err)

				}
			}

			untrack(submittedURL)
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
			track(nameValue, urlValue, forgeValue, releaseValue)
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

func staticHandler(writer http.ResponseWriter, request *http.Request) {
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
