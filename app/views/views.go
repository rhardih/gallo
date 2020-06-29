package views

import (
	"gallo/app/constants"
	"gallo/app/helpers"
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/sessions"
	"github.com/oxtoacart/bpool"
)

var (
	bufferPool *bpool.BufferPool
	Store      *sessions.CookieStore
)

func init() {
	bufferPool = bpool.NewBufferPool(16)
}

func Execute(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	fileName := path.Join("app", "views", name)

	files := []string{
		"app/views/layouts/application.html.tmpl",
		fileName,
	}

	buf := bufferPool.Get()
	defer bufferPool.Put(buf)

	requestDependantFuncs := template.FuncMap{
		"isLoggedIn": func() bool {
			session, _ := Store.Get(r, constants.SessionName)

			if _, ok := session.Values[constants.TrelloTokenSessionKey]; ok {
				return true
			}

			return false
		},
	}

	tmpl := template.New(fileName).Funcs(helpers.Funcs).Funcs(requestDependantFuncs)
	tmpl = template.Must(tmpl.ParseFiles(files...))

	err := tmpl.ExecuteTemplate(buf, "application.html.tmpl", data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println("Error in views.Execute:", err)
	}
}
