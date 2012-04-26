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
	"pat"
	"strings"
	"time"
)

const (
	RECENTLY_FEATURED_FEED = "https://gdata.youtube.com/feeds/api/standardfeeds/recently_featured"
	SEARCH_FEED            = "https://gdata.youtube.com/feeds/api/videos?q=%s&v=2&max-results=%s"
	COOKIE_NAME            = "items_per_page"
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
		if err := memcache.Set(c, front_page); err != nil {
			c.Errorf("error setting item: %v - %v", err)
		}
	}
	fmt.Fprint(w, string(front_page.Value))
}

func searchPage(w http.ResponseWriter, r *http.Request) {	
	c := appengine.NewContext(r)	
	searchTerm, err := url.QueryUnescape(r.URL.Query().Get(":query"))	
	if searchTerm == "" || searchTerm == "search" || err != nil { // try using get parameters (maybe js disabled)
		searchTerm = r.FormValue("q")		
		if searchTerm == "" { // now redirect
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} 		
	}
	searchElements := strings.SplitN(searchTerm, "&", 2)	
	searchTerm = searchElements[0]
	advanced := ""
	if len(searchElements) > 1 {
		advanced = searchElements[1]
	}	
	ippCookie, err := r.Cookie(COOKIE_NAME)
	ipp := "25"
	if err == nil {
		ipp = ippCookie.Value
	}
	query := fmt.Sprintf(SEARCH_FEED, url.QueryEscape(searchTerm), ipp)
	if advanced != "" {
		query += "&" + advanced
	}
	//c.Infof("query: %s", query)
	yentries, err := gdata.ParseFeed(query, urlfetch.Client(c))	
	if err != nil {
		c.Errorf("error getting entries: %v", err)
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
	m := pat.New()
	m.Get("/search/:query", http.HandlerFunc(searchPage))	
	m.Get("/:query", http.HandlerFunc(searchPage)) // in case of js disabled
	m.Get("/", http.HandlerFunc(mainPage))
	http.Handle("/", m)

	//http.HandleFunc("/", mainPage)
	//http.HandleFunc("/search", searchPage)
	http.HandleFunc("/about", aboutPage)
	http.HandleFunc("/contact", contactPage)
	http.HandleFunc("/items", itemsPerPageQuery)
	http.HandleFunc("/advanced", advancedSearchQuery)
}
