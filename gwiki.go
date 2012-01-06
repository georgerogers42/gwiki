package main

import (
	"github.com/hoisie/web.go"
	"github.com/hoisie/mustache.go"
	"github.com/russross/blackfriday"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
	"html"
	"os"
)
var server string
var viewtpl, viewtplerr = mustache.ParseFile("view.html")

func index() string {
	return view("index")
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
	defer session.Close()
	check(err)
	result, err := getPage(session, title)
	check(err)
	result.Body = string(blackfriday.MarkdownCommon([]byte(result.Body)))
	return viewtpl.Render(result)
}

var createtpl, createtplerr = mustache.ParseFile("create.html")

func getPage(session *mgo.Session, title string) (result *Page, err os.Error) {
	result = new(Page)
	c := session.DB("gwiki").C("pages")
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
	res := &Page{Title: title, Body: html.EscapeString(c.Params["body"])}
	db := session.DB("gwiki").C("pages")
	db.Upsert(result, res)
	c.Redirect(302, "/view/"+title)
}
func main() {
	check(viewtplerr)
	check(createtplerr)
	server = os.Args[2]
	web.Get("/", index)
	web.Get("/view/(.*)", view)
	web.Get("/edit/(.*)", edit)
	web.Post("/edit/(.*)", create)
	web.Run(":" + os.Args[1])
}
