package gdata

import (
	"errors"
	"fmt"
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

func ParseEntry(entry string) (ye YEntry, err error) {
	urlRe := regexp.MustCompile(`<media:player url='(.*?)&amp;feature=youtube_gdata_player'/>`)
	thumbRe := regexp.MustCompile(`<media:thumbnail url='(.*?)'`)
	titleRe := regexp.MustCompile(`<media:title type='plain'>(.*?)</media:title>`)
	descriptionRe := regexp.MustCompile(`(?s)<media:description type='plain'>(.*?)</media:description>`)
	categoryRe := regexp.MustCompile(`<media:category.*?>(.*?)</media:category>`)
	keywordsRe := regexp.MustCompile(`<media:keywords>(.*?)</media:keywords>`)
	publishedRe := regexp.MustCompile(`<published>(.*?)</published>`)
	ratingRe := regexp.MustCompile(`<gd:rating average='(.*?)' .*? numRaters='(.*?)'`)
	if url := urlRe.FindStringSubmatch(entry); len(url) == 2 {
		ye.Url = url[1]
	} else {
		err = errors.New("Malformed entry")
	}
	if thumb := thumbRe.FindStringSubmatch(entry); len(thumb) == 2 {
		ye.Thumb = thumb[1]
	} else {
		err = errors.New("Malformed entry")
	}
	if title := titleRe.FindStringSubmatch(entry); len(title) == 2 {
		ye.Title = title[1]
	} else {
		err = errors.New("Malformed entry")
	}
	if description := descriptionRe.FindStringSubmatch(entry); len(description) == 2 {
		ye.Description = description[1]
	} else {
		err = errors.New("Malformed entry")
	}
	if category := categoryRe.FindStringSubmatch(entry); len(category) == 2 {
		ye.Category = category[1]
	} else {
		err = errors.New("Malformed entry")
	}
	if keywords := keywordsRe.FindStringSubmatch(entry); len(keywords) == 2 {
		ye.Keywords = keywords[1]
	} else {
		err = errors.New("Malformed entry")
	}
	if published := publishedRe.FindStringSubmatch(entry); len(published) == 2 {
		ye.Published, err = time.Parse(time.RFC3339, published[1])
	} else {
		err = errors.New("Malformed entry")
	}
	if rating := ratingRe.FindStringSubmatch(entry); len(rating) == 3 {
		ye.Rating = fmt.Sprintf("Rating: %s of 5 stars<br/>%s Votes", rating[1], rating[2])
	} else {
		err = errors.New("Malformed entry")
	}
	return
}

func ParseFeed(feed string, client *http.Client) (yentries []YEntry, err error) {
	var resp *http.Response
	if client != nil {
		resp, err = client.Get(feed)
	} else {
		resp, err = http.Get(feed)
	}
	if err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		entryRe := regexp.MustCompile(`(?s)<entry>.*?</entry>`)
		entries := entryRe.FindAllString(string(body), 50)
		for _, e := range entries {
			if ye, err := ParseEntry(e); err == nil {
				yentries = append(yentries, ye)
			}
		}
	}
	return
}
