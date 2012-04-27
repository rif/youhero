package gdata

import (
	"appengine"
	"appengine/urlfetch"
	"strings"
)

func RemoveDuplicates(query, ipp string, c appengine.Context) (nodupes []*YEntry) {
	nodupes, err := ParseFeed(query, urlfetch.Client(c))
	if err != nil {
		c.Errorf("error getting entries: %v", err)
	}
	for x, ye := range nodupes {
		for y, other := range nodupes[x+1:] { // compare with the following
			similatity := checkSimilarity(ye.Title, other.Title)
			if similatity > 70 {
				nodupes = sliceRemove(nodupes, y)
				c.Infof("%s <==> %s : &v", ye.Title, other.Title, similatity)
			}
		}
		if x > len(nodupes)-2 {
			break
		}
	}
	return
}

func checkSimilarity(w1, w2 string) (procentage int) {
	w1 = strings.ToLower(w1)
	w2 = strings.ToLower(w2)
	words := strings.Split(w1, " ")
	similarities := 0
	for _, word := range words { // take word by words
		if len(word) > 1 && strings.Contains(w2, word) {
			similarities += 1
		}
	}
	s1 := (similarities * 100) / len(words)
	words = strings.Split(w2, " ")
	similarities = 0
	for _, word := range words { // take word by words
		if len(word) > 1 && strings.Contains(w1, word) {
			similarities += 1
		}
	}
	s2 := (similarities * 100) / len(words)
	return (s1 + s2) / 2
}

func sliceRemove(slice []*YEntry, index int) (out []*YEntry) {
	if index < len(slice)-1 {
		out = append(slice[:index], slice[index+1:]...)
	} else {
		out = slice[:index]
	}
	return
}
