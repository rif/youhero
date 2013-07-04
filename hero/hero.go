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
	"net/url"
	"time"
)

const (
	//RECENTLY_FEATURED_FEED = "https://gdata.youtube.com/feeds/api/standardfeeds/recently_featured"
	RECENTLY_FEATURED_FEED = "https://gdata.youtube.com/feeds/api/standardfeeds/top_rated?time=today"
	SEARCH_FEED            = "https://gdata.youtube.com/feeds/api/videos?q=%s&v=2&key=AI39si6Qiy5xKw3x-ODfoN94rbfcjFaAVAxXLtFpKOtHg2iAM23H77IGdhbhxnNl9YvcjxvmSIVjdaoqw76glQChwWr97_k5Yg"
	COOKIE_NAME            = "items_per_page"
)

func mainPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	front_page, err := memcache.Get(c, "front_page")
	if err != nil {
		yentries, err := gdata.ParseFeed(RECENTLY_FEATURED_FEED, urlfetch.Client(c))
		cache := true
		if err != nil || len(yentries) == 0 {
			c.Errorf("error getting entries: %v", err)
			cache = false
		}
		templateValues := map[string]interface{}{"entries": yentries,
			"title":    "YouHero Featured Videos",
			"header":   "Recently Featured Videos",
			"autoplay": false,
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
		if cache {
			if err := memcache.Set(c, front_page); err != nil {
				c.Errorf("error setting item: %v - %v", err)
			}
		}
	}
	fmt.Fprint(w, string(front_page.Value))
}

func searchPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	searchTerm := r.FormValue("q")
	if searchTerm == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	query := fmt.Sprintf(SEARCH_FEED, url.QueryEscape(searchTerm))
	// advanced search options
	if cat := r.FormValue("category"); cat != "" && cat != "Any" {
		query += "&category=" + cat
	}
	if hd := r.FormValue("hd"); hd != "" {
		query += "&hd="
	}
	if order := r.FormValue("orderby"); order != "" {
		query += "&orderby=" + order
	}
	ippCookie, err := r.Cookie(COOKIE_NAME)
	ipp := "25"
	if err == nil {
		ipp = ippCookie.Value
	}
	//c.Infof("query: %s", query)
	var yentries []*gdata.YEntry
	if dedup := r.FormValue("dedup"); dedup != "" {
		query += "&max-results=50" // get max results and truncate later
		yentries = gdata.RemoveDuplicates(query, ipp, c)
	} else {
		query += "&max-results=" + ipp
		yentries, err = gdata.ParseFeed(query, urlfetch.Client(c))
		if err != nil {
			c.Errorf("error getting entries: %v", err)
		}
	}
	templateValues := map[string]interface{}{"entries": yentries,
		"title":    searchTerm,
		"header":   fmt.Sprintf("Searching for '%s'", searchTerm),
		"autoplay": true,
	}
	t, _ := template.ParseFiles("templates/base.html", "templates/index.html")
	t.Execute(w, templateValues)
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method == "POST" && r.FormValue("vfy") == "tuc" {
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
	if nb := r.FormValue("nb"); nb != "" {
		http.SetCookie(w, &http.Cookie{Name: COOKIE_NAME, Value: nb, Path: "/search", Expires: expiration})
	}

}

func advancedSearchQuery(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/advanced.html")
	t.Execute(w, nil)
}

func init() {
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/search", searchPage)
	http.HandleFunc("/about", aboutPage)
	http.HandleFunc("/contact", contactPage)
	http.HandleFunc("/items", itemsPerPageQuery)
	http.HandleFunc("/advanced", advancedSearchQuery)
}
