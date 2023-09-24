// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// listRemoteTags lists all tags in a remote repository, whether HTTP(S) or SSH.
func listRemoteTags(url string) (tags []string, err error) {
	// TODO: Implement listRemoteTags
	// https://pkg.go.dev/github.com/go-git/go-git/v5@v5.8.0#NewRemote
	return nil, nil
}

// fetchReleases fetches all releases in a remote repository, whether HTTP(S) or SSH.
func getGitReleases(p project) (project, error) {
	r, err := minimalClone(p.URL)
	if err != nil {
		return p, err
	}
	tagRefs, err := r.Tags()
	if err != nil {
		return p, err
	}
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		obj, err := r.TagObject(tagRef.Hash())
		switch err {
		case plumbing.ErrObjectNotFound:
			// This is a lightweight tag, not an annotated tag, skip it
			return nil
		case nil:
			url := ""
			tagName := bmStrict.Sanitize(tagRef.Name().Short())
			switch p.Forge {
			case "sourcehut":
				url = p.URL + "/refs/" + tagName
			case "gitlab":
				url = p.URL + "/-/releases/" + tagName
			default:
				url = ""
			}
			p.Releases = append(p.Releases, release{
				Tag:     tagName,
				Content: bmUGC.Sanitize(obj.Message),
				URL:     url,
				Date:    obj.Tagger.When,
			})
		default:
			return err
		}
		return nil
	})
	if err != nil {
		return p, err
	}

	sort.Slice(p.Releases, func(i, j int) bool { return p.Releases[i].Date.After(p.Releases[j].Date) })

	return p, nil
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
		if err == git.NoErrAlreadyUpToDate {
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

// removeRepo removes a repository from the local filesystem.
func removeRepo(url string) (err error) {
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
