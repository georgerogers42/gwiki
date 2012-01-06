package main

import (
	"github.com/russross/blackfriday"
	"html"
	"html/template"
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

var dbname = "gwiki"
var server = "localhost"
var viewtpl = template.New("view.html")

func index(w http.ResponseWriter, r *http.Request) {
	rx := regexp.MustCompile("/(\\w+)$")
	page := "index"
	if x := rx.FindStringSubmatch(r.URL.Path); x != nil {
		page = x[1]
	}
	if page == "" {
		page = "index"
	}
	http.Redirect(w, r, "/view/"+page, 302)
}

type Page struct {
	Title string
	Body  string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func view(w http.ResponseWriter, c *http.Request) {
	r := regexp.MustCompile("/(\\w+)$")
	title := "index"
	if x := r.FindStringSubmatch(c.URL.Path); x != nil {
		title = x[1]
	}
	session, err := mgo.Mongo(server)
	check(err)
	defer session.Close()
	result, err := getPage(session, title)
	check(err)
	result.Body = html.EscapeString(result.Body)
	result.Body = string(blackfriday.MarkdownCommon([]byte(result.Body)))
	viewtpl.Execute(w, result)
}

var createtpl = template.New("create.html")

func getPage(session *mgo.Session, title string) (result *Page, err error) {
	result = new(Page)
	c := session.DB(dbname).C("pages")
	err = c.Find(bson.M{"title": title}).One(result)
	return
}
func matchUrl(pat, against string) (title string) {
	r := regexp.MustCompile(pat + "$")
	title = "index"
	if x := r.FindStringSubmatch(against); x != nil {
		title = x[1]
	}
	return
}
func edit(w http.ResponseWriter, c *http.Request) {
	title := matchUrl(c.URL.Path, "/(\\w+)")
	session, err := mgo.Mongo(server)
	check(err)
	defer session.Close()
	result, err := getPage(session, title)
	if err == mgo.NotFound {
		result.Title = title
	} else {
		check(err)
	}
	createtpl.Execute(w, result)
}
func main() {
	s := os.Args[2]
	x, err := url.Parse(s)
	check(err)
	dbname = x.Path[1:]
	server = s
	http.HandleFunc("/view/", view)
	http.HandleFunc("/edit/", edit)
	http.HandleFunc("/", index)
	http.ListenAndServe(":"+os.Args[1], nil)
}
