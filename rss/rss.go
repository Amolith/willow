// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package rss

import (
	"fmt"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"

	"github.com/mmcdole/gofeed"
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

func GetReleases(feedURL string) ([]Release, error) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(strings.TrimSuffix(feedURL, "/") + "/releases.atom")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	releases := make([]Release, 0)

	for _, item := range feed.Items {
		releases = append(releases, Release{
			Tag:     bmStrict.Sanitize(item.Title),
			Content: bmUGC.Sanitize(item.Content),
			URL:     bmStrict.Sanitize(item.Link),
			Date:    *item.PublishedParsed,
		})
	}

	return releases, nil
}
