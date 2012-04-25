package hello

import (
	"appengine"
	"appengine/mail"
	"appengine/memcache"
	"appengine/urlfetch"
	"bytes"
	"fmt"
	"gdata"
	"html/template"
	"net/http"
	"time"
)

const (
	RECENTLY_FEATURED_FEED = "https://gdata.youtube.com/feeds/api/standardfeeds/recently_featured"
	SEARCH_FEED            = "https://gdata.youtube.com/feeds/api/videos?q=%s"
)

func mainPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	front_page, err := memcache.Get(c, "front_page")
	if err != nil {
		yentries, err := gdata.ParseFeed(RECENTLY_FEATURED_FEED, urlfetch.Client(c))
		if err != nil {
			c.Errorf("error getting entries: %v", err)
		}
		templateValues := map[string]interface{}{"entries": yentries,
			"title":    "YouHero",
			"header":   "Recently Featured Videos",
			"autoplay": "false",
		}
		buf := &bytes.Buffer{}
		t, _ := template.ParseFiles("templates/base.html", "templates/index.html")
		t.Execute(buf, templateValues)
		oneHour, _ := time.ParseDuration("1h")
		front_page = &memcache.Item{
			Key:        "front_page",
			Value:      buf.Bytes(),
			Expiration: oneHour,
		}
		if err := memcache.Set(c, front_page); err != nil {
			c.Errorf("error setting item: %v - %v", err)
		}
	}
	fmt.Fprint(w, string(front_page.Value))
}

func searchPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	searchTerm := r.FormValue("v")
	if searchTerm == "" {
		http.Redirect(w, r, "/", http.StatusBadRequest)
		return
	}
	yentries, err := gdata.ParseFeed(fmt.Sprintf(SEARCH_FEED, searchTerm), urlfetch.Client(c))
	if err != nil {
		c.Errorf("error getting entries: %v", err)
	}
	templateValues := map[string]interface{}{"entries": yentries,
		"title":    searchTerm,
		"header":   fmt.Sprintf("Searching for '%s'", searchTerm),
		"autoplay": "true",
	}
	t, _ := template.ParseFiles("templates/base.html", "templates/index.html")
	t.Execute(w, templateValues)
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method == "POST" {
		name := r.FormValue("from")
		email := r.FormValue("email")
		content := r.FormValue("content")
		msg := &mail.Message{
			Sender:  "Radu Fericean (YouHero) <fericean@gmail.com>",
			To:      []string{"Radu Fericean <radu@fericean.ro>"},
			Subject: fmt.Sprintf("YouHero message from %s (%s)", name, email),
			Body:    content,
		}
		if err := mail.Send(c, msg); err != nil {
			c.Errorf("Couldn't send email: %v", err)
		}
		http.Redirect(w, r, "/", http.StatusOK)
		return
	}
	t, _ := template.ParseFiles("templates/base.html", "templates/contact.html")
	t.Execute(w, nil)
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/base.html", "templates/about.html")
	t.Execute(w, nil)
}

func itemsPerPageQuery(w http.ResponseWriter, r *http.Request) {
	thirtyDays, _ := time.ParseDuration("720h")
	expiration := time.Now().Add(thirtyDays)
	nb := r.FormValue("nb")
	if nb == "" {
		nb = "25"
	}
	w.Header().Add("Set-Cookie", fmt.Sprintf("items_per_page=%s; expires=%s; path=/search;", nb, expiration.Format(time.RFC1123)))
}

func init() {
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/search", searchPage)
	http.HandleFunc("/about", aboutPage)
	http.HandleFunc("/contact", contactPage)
	http.HandleFunc("/items", itemsPerPageQuery)
}
