package crss

import (
    "github.com/mmcdole/gofeed"
    //"github.com/mmcdole/gofeed/rss"
	//"fmt"
	"time"
)

type Entry struct {
	Title string `json:"title"`
	Description string `json:"desc"`
	URL string `json:"url"`
	Updated int64 `json:"updated"`	
}
func GetFeedUpdatesWithPOSIX(feedURI string, since int64) ([]Entry, error) {
	time := time.Unix(since, 0)
	return GetFeedUpdates(feedURI, time)
}
func GetFeedUpdates(feedURI string, since time.Time) ([]Entry, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURI)
	if err != nil {
		return nil, err
	}
	//newEntries = make([]Entry, 0)
	var last int
	last = 0
	for en:=range feed.Items {
		pub, err := time.Parse(time.RFC3339, feed.Items[en].Published)
		if err != nil {
			//panic(fmt.Sprintf("Error converting publish datetime (%V)!", err))
			return nil, err
		}
		if pub.After(since) {
			last=en
		}
	}
	updates := make([]Entry, last+1)
	for en:=range updates {
		updates[en].Title = feed.Items[en].Title
		updates[en].Description = feed.Items[en].Description
		updates[en].URL = feed.Items[en].Link
		pub, err := time.Parse(time.RFC3339, feed.Items[en].Published)
		if err != nil {
			//panic(fmt.Sprintf("Error converting publish datetime (%V)!", err))
			//return nil, err
			continue // TODO: fix this
		}
		updates[en].Updated = pub.Unix()
	}
	return updates, nil
}