package main

import (
	"github.com/georgerogers42/gologin"
	"github.com/georgerogers42/goweb"
	"github.com/russross/blackfriday"
	"html"
	"launchpad.net/gobson/bson"
	"launchpad.net/mgo"
	"net/http"
	"net/url"
	"os"
	"text/template"
)

var dbname = "gwiki"
var server = "localhost"
var viewtpl = template.Must(template.ParseFiles("view.html"))

type Page struct {
	Title string
	Body  string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func view(w http.ResponseWriter, c *http.Request, s goweb.Result, path ...string) goweb.Result {
	title := path[0]
	session, err := mgo.Mongo(server)
	check(err)
	defer session.Close()
	result, err := getPage(session, title)
	if err == mgo.NotFound {
		http.Redirect(w, c, "/edit/"+title, 302)
	}
	result.Body = html.EscapeString(result.Body)
	result.Body = string(blackfriday.MarkdownCommon([]byte(result.Body)))
	viewtpl.Execute(w, result)
	return s
}

var createtpl = template.Must(template.ParseFiles("create.html"))
var x = "users"
var _, _ = gologin.FromConfig("userpass:anonimous:coward")
var route = goweb.Or(
	goweb.Route(`/edit/(\w+)`, gologin.RequireAuth(edit)),
	goweb.Route(`/view/(\w+)`, view),
	goweb.Route(`/(\w+)`, view),
	goweb.Route(`/`, index))

func getPage(session *mgo.Session, title string) (result *Page, err error) {
	result = new(Page)
	c := session.DB(dbname).C("pages")
	err = c.Find(bson.M{"title": title}).One(result)
	return
}
func index(w http.ResponseWriter, c *http.Request, s goweb.Result, path ...string) goweb.Result {
	http.Redirect(w, c, "/index", 302)
	return s
}
func edit(w http.ResponseWriter, c *http.Request, s goweb.Result, path ...string) goweb.Result {
	title := path[0]
	session, err := mgo.Mongo(server)
	check(err)
	defer session.Close()
	result, err := getPage(session, title)
	if c.Method == "GET" {
		if err == mgo.NotFound {
			result.Title = title
		} else {
			check(err)
		}
		createtpl.Execute(w, result)
	} else if c.Method == "POST" {
		page := new(Page)
		page.Title = title
		page.Body = c.FormValue("body")
		ctx := session.DB(dbname).C("pages")
		if err == mgo.NotFound {
			err := ctx.Insert(page)
			check(err)
		} else {
			check(err)
			_, err := ctx.Upsert(result, page)
			check(err)
		}
		http.Redirect(w, c, "/view/"+title, 302)
	}
	return s
}

var handler = goweb.Handler(route)

func main() {
	s := os.Args[2]
	x, err := url.Parse(s)
	check(err)
	dbname = x.Path[1:]
	server = s
	http.ListenAndServe(":"+os.Args[1], http.HandlerFunc(handler))
}
