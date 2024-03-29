// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"errors"
	"fmt"
	"net/url"
	"os"
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
		tagObj, err := r.TagObject(tagRef.Hash())

		var message string
		var date time.Time
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			commitTag, err := r.CommitObject(tagRef.Hash())
			if err != nil {
				return err
			}
			message = commitTag.Message
			date = commitTag.Committer.When
		} else {
			message = tagObj.Message
			date = tagObj.Tagger.When
		}

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
			Content: bmUGC.Sanitize(message),
			URL:     tagURL,
			Date:    date,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

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
	if err != nil {
		return err
	}

	path = path[:strings.LastIndex(path, "/")]
	dirs := strings.Split(path, "/")

	for range dirs {
		if path == "data" {
			break
		}
		err = os.Remove(path)
		if err != nil {
			// This folder likely has data, so might as well save some time by
			// not checking the parents we can't delete anyway.
			break
		}
		path = path[:strings.LastIndex(path, "/")]
	}

	return nil
}

// stringifyRepo accepts a repository URI string and the corresponding local
// filesystem path, whether the URI is HTTP, HTTPS, or SSH.
func stringifyRepo(url string) (path string, err error) {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimSuffix(url, "/")

	ep, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	if ep.Protocol == "http" || ep.Protocol == "https" {
		return "data/" + strings.Split(url, "://")[1], nil
	} else if ep.Protocol == "ssh" {
		return "data/" + ep.Host + "/" + ep.Path, nil
	} else {
		return "", errors.New("unsupported protocol")
	}
}
