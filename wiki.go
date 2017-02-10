// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

type page struct {
	Title string
	Body  []byte
}

func (p *page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var edit = template.Must(template.ParseFiles("index.html", "edit.html"))
var view = template.Must(template.ParseFiles("index.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *page) {
	switch tmpl {
	case "view":
		err := view.ExecuteTemplate(w, "layout", p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		break
	case "edit":
		err := edit.ExecuteTemplate(w, "layout", p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		break
	}

	// t, err := templates.ParseFiles(tmpl + ".html")
	// // t, err := template.ParseFiles("index.html", tmpl+".html")
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// err = t.ExecuteTemplate(w, "layout", p)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	fs := http.FileServer(http.Dir("node_modules"))
	http.Handle("/node_modules/", http.StripPrefix("/node_modules/", fs))

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.ListenAndServe(":8080", nil)
}
