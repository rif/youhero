package hello

import (
	"appengine"
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
	SEARCH_FEED            = "https://gdata.youtube.com/feeds/api/videos?q=surfing"
)

func mainPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	front_page, err := memcache.Get(c, "front_page")
	if err != nil {
		yentries, err := gdata.ParseFeed(RECENTLY_FEATURED_FEED, urlfetch.Client(c))
		if err != nil {
			c.Infof("error getting entries: %v", err)
		}
		templateValues := map[string]interface{}{"entries": yentries,
			"title":    "Recently Featured Videos",
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
			c.Infof("error setting item: %v - %v", err)
		}
	}
	fmt.Fprint(w, string(front_page.Value))
}

func searchPage(w http.ResponseWriter, r *http.Request) {
	/*def get(self):    
	  search_term = cgi.escape(self.request.get("v")).encode('UTF-8')            
	  if not search_term:
	      self.redirect('/')
	      return

	  client = gdata.youtube.service.YouTubeService()
	  gdata.alt.appengine.run_on_appengine(client)
	  query = gdata.youtube.service.YouTubeVideoQuery()

	  query.vq = search_term
	  query.max_results = self.request.cookies.get('items_per_page', '25')
	  template_values = {
	      'feed': client.YouTubeQuery(query),
	      'title': "Searching for '%s'" % search_term.decode('UTF-8'),
	      'autoplay': 'true',
	  }
	  template = jinja_environment.get_template('templates/index.html')
	  self.response.out.write(template.render(template_values))*/
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	/*def get(self):
	      template = jinja_environment.get_template('templates/contact.html')
	      self.response.out.write(template.render({}))        
	  def post(self):
	      name = cgi.escape(self.request.get('from')).encode('UTF-8')
	      email = cgi.escape(self.request.get('email')).encode('UTF-8')
	      message = mail.EmailMessage(sender="Radu Fericean (YouHero) <fericean@gmail.com>",
	                          subject="YouHero message from %s (%s)" % (name, email))
	      message.to = "Radu Fericean <radu@fericean.ro>"
	      message.body = cgi.escape(self.request.get('content'))
	      message.send()
	      self.redirect('/')*/
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/base.html", "templates/about.html")
	t.Execute(w, nil)
}

func itemsPerPageQuery(w http.ResponseWriter, r *http.Request) {
	/*def get(self):
	  expiration = datetime.datetime.utcnow() + datetime.timedelta(days=30)
	  self.response.headers.add_header('Set-Cookie','items_per_page=%s; expires=%s; path=/search;' %
	      (str(self.request.get("nb", '25')), expiration.strftime("%a, %d-%b-%Y %H:%M:%S UTC")))*/
}

func init() {
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/search", searchPage)
	http.HandleFunc("/about", aboutPage)
	http.HandleFunc("/contact", contactPage)
	http.HandleFunc("/items", itemsPerPageQuery)
}
