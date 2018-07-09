package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"errors"
	"strings"
	"github.com/mmcdole/gofeed"
	"time"
	"../../winterfaces"
)

type YRSSGetter struct{}

func CheckRSSLink(url string) (bool, error) { // TODO
	return true, nil
}

func GetRSSLinkUniversal(req string) (string, error) {
	if req[0:4] == "http" {
		return GetRSSLinkByLink(req)
	} else {
		return GetRSSLinkByName(req)
	}
}

func GetRSSLinkByName(query string) (string, error) {
	var body []byte
	var err error

	body, err = getBody(strings.Replace(fmt.Sprintf("https://www.youtube.com/results?search_query=%s", query), " ", "+", -1))
	if err != nil {
		return "", err
	}
	// match user link
	var pattern *regexp.Regexp
	var matched string

	pattern = regexp.MustCompile("href=\"/user/([a-z,A-Z,-,_,0-9]*)\"")
	matched = pattern.FindString(string(body))
	if matched != "" {
		matched = fmt.Sprintf(matched[6 : len(matched)-1])
	}
	return GetRSSLinkByLink(fmt.Sprintf("https://www.youtube.com%s", matched))
}

func GetRSSLinkByLink(url string) (string, error) {
	// get body from URL
	var body []byte
	var err error
	body, err = getBody(url)

	if err != nil {
		log.Fatal(err)
		return "", err
	}

	// match external channel ID from the page
	var pattern *regexp.Regexp
	var matched string

	pattern = regexp.MustCompile("channel-external-id=\"([a-z,A-Z,-,_,0-9]*)\"")
	matched = pattern.FindString(string(body))
	if len(matched)<22 {
		return "", errors.New("Couldn't extract URL")
	}
		
	matched = matched[21 : len(matched)-1]

	// return link to RSS feed
	return fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", matched), nil
}

func getBody(url string) ([]byte, error) {
	var resp *http.Response
	var err error

	// GET request to URL (presumed youtube)
	resp, err = http.Get(url)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// close connection after function done
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func GetFeedUpdates(feedURI string, requestType string, since time.Time) ([]winterfaces.Entry, error) {
	var url string
	var err error
	if requestType == "query" { // TODO : change to config
		url, err = GetRSSLinkUniversal(feedURI)
		if err != nil {
			return nil, err
		}
	} else if requestType == "URL" {
		url = feedURI
	} else {
		return nil, errors.New("Unrecognized request type")
	}
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}
	//newEntries = make([]winterfaces.Entry, 0)
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
	updates := make([]winterfaces.Entry, last+1)
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
func GetFeedUpdatesWithPOSIX(feedURI string, requestType string, since int64) ([]winterfaces.Entry, error) {
	time := time.Unix(since, 0)
	return GetFeedUpdates(feedURI, requestType, time)
}

// main function
func (yr *YRSSGetter) CheckUpdates(request string, requestType string, since interface{}) ([]winterfaces.Entry, error) {
	switch since.(type) {
		case int64: return GetFeedUpdatesWithPOSIX(request, requestType, since.(int64))
		case time.Time: return GetFeedUpdates(request, requestType, since.(time.Time))
		default: return nil, errors.New("Wrong time type for since variable")
	}
}

var Parser YRSSGetter