package controllers

import (
	"gallo/app/views"
	"net/http"
	"path"
	"regexp"
)

type ApplicationController struct{}

func (c ApplicationController) AssetsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("app", r.URL.Path))
}

func (c ApplicationController) RootHandler(w http.ResponseWriter, r *http.Request) {
	if regexp.MustCompile("^/$").MatchString(r.URL.Path) {
		views.Execute(w, r, "application/home.html.tmpl", nil)
	} else {
		http.ServeFile(w, r, path.Join("public", r.URL.Path))
	}
}
