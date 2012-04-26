package gdata

import (
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type YEntry struct {
	Url         string
	Thumb       string
	Title       string
	Description string
	Category    string
	Keywords    string
	Published   time.Time
	Rating      string
}

func (ye *YEntry) GetPublishedTime() string {
	return ye.Published.Format(time.RFC1123)
}

func ParseEntry(entry string) (ye *YEntry, err error) {
	urlRe := regexp.MustCompile(`<media:player url='(.*?)&amp;feature=youtube_gdata_player'/>`)
	thumbRe := regexp.MustCompile(`<media:thumbnail url='(.*?)'`)
	titleRe := regexp.MustCompile(`<media:title type='plain'>(.*?)</media:title>`)
	descriptionRe := regexp.MustCompile(`(?s)<media:description type='plain'>(.*?)</media:description>`)
	categoryRe := regexp.MustCompile(`<media:category.*?>(.*?)</media:category>`)
	keywordsRe := regexp.MustCompile(`<media:keywords>(.*?)</media:keywords>`)
	publishedRe := regexp.MustCompile(`<published>(.*?)</published>`)
	ratingRe := regexp.MustCompile(`<gd:rating average='(.*?)' .*? numRaters='(.*?)'`)
	ye = &YEntry{}
	if url := urlRe.FindStringSubmatch(entry); len(url) == 2 {
		ye.Url = url[1]
	} else {
		err = errors.New("Malformed entry: url")
	}
	if thumb := thumbRe.FindStringSubmatch(entry); len(thumb) == 2 {
		ye.Thumb = thumb[1]
	} else {
		err = errors.New("Malformed entry: thumb")
	}
	if title := titleRe.FindStringSubmatch(entry); len(title) == 2 {
		ye.Title = html.UnescapeString(title[1])
	} else {
		err = errors.New("Malformed entry: title")
	}
	if description := descriptionRe.FindStringSubmatch(entry); len(description) == 2 {
		ye.Description = html.UnescapeString(description[1])
	} else {
		err = errors.New("Malformed entry: description")
	}
	if category := categoryRe.FindStringSubmatch(entry); len(category) == 2 {
		ye.Category = html.UnescapeString(category[1])
	} else {
		err = errors.New("Malformed entry: category")
	}
	if keywords := keywordsRe.FindStringSubmatch(entry); len(keywords) == 2 {
		ye.Keywords = html.UnescapeString(keywords[1])
	} else {
		err = errors.New("Malformed entry: keywords")
	}
	if published := publishedRe.FindStringSubmatch(entry); len(published) == 2 {
		ye.Published, err = time.Parse(time.RFC3339, published[1])
	} else {
		err = errors.New("Malformed entry: published")
	}
	if rating := ratingRe.FindStringSubmatch(entry); len(rating) == 3 {
		ye.Rating = fmt.Sprintf("Rating: %s of 5 stars - %s Votes", rating[1], rating[2])
	}
	return
}

func ParseFeed(feed string, client *http.Client) (yentries []*YEntry, err error) {
	var resp *http.Response
	if client != nil {
		resp, err = client.Get(feed)
	} else {
		resp, err = http.Get(feed)
	}
	if err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		entryRe := regexp.MustCompile(`(?s)<entry.*?>(.*?)</entry>`)
		entries := entryRe.FindAllString(string(body), 50)
		for _, e := range entries {
			if ye, err := ParseEntry(e); err == nil {
				yentries = append(yentries, ye)
			}
		}
	}
	return
}
