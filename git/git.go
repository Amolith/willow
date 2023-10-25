// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type Release struct {
	Tag     string
	Content string
	URL     string
	Date    time.Time
}

var (
	bmUGC    = bluemonday.UGCPolicy()
	bmStrict = bluemonday.StrictPolicy()
)

// listRemoteTags lists all tags in a remote repository, whether HTTP(S) or SSH.
// func listRemoteTags(url string) (tags []string, err error) {
// 	// TODO: Implement listRemoteTags
// 	// https://pkg.go.dev/github.com/go-git/go-git/v5@v5.8.0#NewRemote
// 	return nil, nil
// }

// GetReleases fetches all releases in a remote repository, whether HTTP(S) or
// SSH.
func GetReleases(gitURI, forge string) ([]Release, error) {
	r, err := minimalClone(gitURI)
	if err != nil {
		return nil, err
	}
	tagRefs, err := r.Tags()
	if err != nil {
		return nil, err
	}

	parsedURI, err := url.Parse(gitURI)
	if err != nil {
		fmt.Println("Error parsing URI: " + err.Error())
	}

	var httpURI string
	if parsedURI.Scheme != "" {
		httpURI = parsedURI.Host + parsedURI.Path
	}

	releases := make([]Release, 0)

	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		obj, err := r.TagObject(tagRef.Hash())
		switch {
		case errors.Is(err, plumbing.ErrObjectNotFound):
			// This is a lightweight tag, not an annotated tag, skip it
			return nil
		case err == nil:
			tagURL := ""
			tagName := bmStrict.Sanitize(tagRef.Name().Short())
			switch forge {
			case "sourcehut":
				tagURL = "https://" + httpURI + "/refs/" + tagName
			case "gitlab":
				tagURL = "https://" + httpURI + "/-/releases/" + tagName
			default:
				tagURL = ""
			}
			releases = append(releases, Release{
				Tag:     tagName,
				Content: bmUGC.Sanitize(obj.Message),
				URL:     tagURL,
				Date:    obj.Tagger.When,
			})
		default:
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(releases, func(i, j int) bool { return releases[i].Date.After(releases[j].Date) })

	return releases, nil
}

// minimalClone clones a repository with a depth of 1 and no checkout.
func minimalClone(url string) (r *git.Repository, err error) {
	path, err := stringifyRepo(url)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); err == nil {
		r, err := git.PlainOpen(path)
		if err != nil {
			return nil, err
		}
		err = r.Fetch(&git.FetchOptions{
			RemoteName: "origin",
			Depth:      1,
			Tags:       git.AllTags,
		})
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return r, nil
		}
		return r, err
	}

	r, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:          url,
		SingleBranch: true,
		NoCheckout:   true,
		Depth:        1,
	})
	return r, err
}

// RemoveRepo removes a repository from the local filesystem.
func RemoveRepo(url string) (err error) {
	path, err := stringifyRepo(url)
	if err != nil {
		return err
	}
	err = os.RemoveAll(path)
	return err
}

// stringifyRepo accepts a repository URI string and the corresponding local
// filesystem path, whether the URI is HTTP, HTTPS, or SSH.
func stringifyRepo(url string) (path string, err error) {
	ep, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	if ep.Protocol == "http" || ep.Protocol == "https" {
		return "data/" + strings.Split(url, "://")[1], nil
	} else if ep.Protocol == "ssh" {
		return "data/" + ep.Host + ep.Path, nil
	} else {
		return "", errors.New("unsupported protocol")
	}
}
