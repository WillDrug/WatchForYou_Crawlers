package crss

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"errors"
	"strings"
)

type YRSSGetter struct{}

func (YRSSGetter) CheckRSSLink(url string) (bool, error) {
	return true, nil
}

func (yrs YRSSGetter) GetRSSLinkUniversal(req string) (string, error) {
	if req[0:4] == "http" {
		return yrs.GetRSSLinkByLink(req)
	} else {
		return yrs.GetRSSLinkByName(req)
	}
}

func (yrs YRSSGetter) GetRSSLinkByName(query string) (string, error) {
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
	return yrs.GetRSSLinkByLink(fmt.Sprintf("https://www.youtube.com%s", matched))
}

func (YRSSGetter) GetRSSLinkByLink(url string) (string, error) {
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
