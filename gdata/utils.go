package gdata

import (
	"appengine"
	"appengine/urlfetch"
	"strconv"
	"strings"
)

const (
	SIMILARITY_FACTOR = 75
)

func RemoveDuplicates(query, ipp string, c appengine.Context) (nodupes []*YEntry) {
	itemPerPage, err := strconv.ParseInt(ipp, 10, 64)
	if err != nil {
		itemPerPage = 25
	}
	nodupes, err = ParseFeed(query, urlfetch.Client(c))
	if err != nil {
		c.Errorf("error getting entries: %v", err)
	}
	for x := 0; x < len(nodupes); x++ {
		ye := nodupes[x]
		for y := x + 1; y < len(nodupes); y++ { // compare with the following
			other := nodupes[y]
			similatity := checkSimilarity(ye.Title, other.Title)
			//c.Infof("%s <==> %s : &v", ye.Title, other.Title, similatity)
			if similatity > SIMILARITY_FACTOR {
				nodupes = sliceRemove(nodupes, y)
			}
		}
		if x > len(nodupes)-2 {
			break
		}
	}
	return nodupes[:itemPerPage]
}

func checkSimilarity(w1, w2 string) (procentage int) {
	w1 = strings.ToLower(w1)
	w2 = strings.ToLower(w2)
	words := strings.Split(w1, " ")
	similarities := 0
	for _, word := range words { // take word by words
		if len(word) > 1 && (word[0] != '(' || word[len(word)-1] != ')') && strings.Contains(w2, word) {
			similarities++
		}
	}
	s1 := (similarities * 100) / len(words)
	words = strings.Split(w2, " ")
	similarities = 0
	for _, word := range words { // take word by words
		if len(word) > 1 && (word[0] != '(' || word[len(word)-1] != ')') && strings.Contains(w1, word) {
			similarities++
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
