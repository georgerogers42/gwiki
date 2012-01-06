package main

import (
	"github.com/hoisie/web.go"
	"github.com/hoisie/mustache.go"
	"github.com/russross/blackfriday"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
	"html"
	"os"
	"url"
)

var dbname = "gwiki"
var server = "localhost"
var viewtpl, viewtplerr = mustache.ParseFile("view.html")

func index(c *web.Context, page string) {
	if page == "" {
		page = "index"
	}
	c.Redirect(302, "/view/"+page)
}

type Page struct {
	Title string
	Body  string
}

func check(e os.Error) {
	if e != nil {
		panic(e)
	}
}
func view(title string) string {
	session, err := mgo.Mongo(server)
	check(err)
	defer session.Close()
	result, err := getPage(session, title)
	check(err)
	result.Body = html.EscapeString(result.Body)
	result.Body = string(blackfriday.MarkdownCommon([]byte(result.Body)))
	return viewtpl.Render(result)
}

var createtpl, createtplerr = mustache.ParseFile("create.html")

func getPage(session *mgo.Session, title string) (result *Page, err os.Error) {
	result = new(Page)
	c := session.DB(dbname).C("pages")
	err = c.Find(bson.M{"title": title}).One(result)
	return
}

func edit(title string) string {
	session, err := mgo.Mongo(server)
	check(err)
	defer session.Close()
	result, err := getPage(session, title)
	if err == mgo.NotFound {
		result.Title = title
	} else {
		check(err)
	}
	return createtpl.Render(result)
}
func create(c *web.Context, title string) {
	if title != c.Params["title"] {
		panic("Invalid params")
	}
	session, err := mgo.Mongo(server)
	defer session.Close()
	check(err)
	result, err := getPage(session, title)
	if err == mgo.NotFound {
		result.Title = title
	} else {
		check(err)
	}
	res := &Page{Title: title, Body: c.Params["body"]}
	db := session.DB(dbname).C("pages")
	db.Upsert(result, res)
	c.Redirect(302, "/view/"+title)
}
func main() {
	check(viewtplerr)
	check(createtplerr)
	s := os.Args[2]
	x, err := url.Parse(s)
	check(err)
	dbname = x.Path[1:]
	server = s
	web.Get("/view/([a-zA-Z]*)", view)
	web.Get("/edit/([a-zA-Z]*)", edit)
	web.Get("/([a-zA-Z]*)", index)
	web.Post("/edit/(.*)", create)
	web.Run(":" + os.Args[1])
}
