package main

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

func getRSSReleases(p project) (project, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(p.URL + "/releases.atom")
	if err != nil {
		fmt.Println(err)
		return p, err
	}

	for _, item := range feed.Items {
		p.Releases = append(p.Releases, release{
			Tag:     bmStrict.Sanitize(item.Title),
			Content: bmUGC.Sanitize(item.Content),
			URL:     bmStrict.Sanitize(item.Link),
			Date:    *item.PublishedParsed,
		})
	}

	// TODO: Doesn't seem to work?
	// sort.Slice(p.Releases, func(i, j int) bool { return p.Releases[i].Date.After(p.Releases[j].Date) })

	return p, nil
}
