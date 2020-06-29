package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"gallo/app/constants"
	"gallo/app/views"
	"gallo/lib"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type AuthController struct {
	Store *sessions.CookieStore
}

// Show renders the login page
func (_ AuthController) Show(w http.ResponseWriter, r *http.Request) {
	views.Execute(w, r, "auth/show.html.tmpl", nil)
}

// Authenticate creates a new session for the user
func (a AuthController) Authenticate(w http.ResponseWriter, r *http.Request) {
	if token, ok := mux.Vars(r)["token"]; ok {
		session, _ := a.Store.Get(r, constants.SessionName)

		session.Values[constants.TrelloTokenSessionKey] = token

		err := session.Save(r, w)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
	return
}

// Authorize takes care of creating, and redirecting to, the trello
// authorization url with correct parameters
func (_ AuthController) Authorize(w http.ResponseWriter, r *http.Request) {
	trelloAuthUrl, err := url.Parse("https://trello.com/1/authorize")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	q := trelloAuthUrl.Query()
	q.Set("expiration", "never")
	q.Set("name", "Gallo")
	q.Set("scope", "read")
	q.Set("response_type", "token")
	q.Set("key", lib.MustGetEnv("TRELLO_KEY"))

	if trello, ok := mux.Vars(r)["trello"]; ok && trello == "return" {
		host := lib.MustGetEnv("HOST")
		q.Set("return_url", fmt.Sprintf("%s/auth", host))
	}

	trelloAuthUrl.RawQuery = q.Encode()

	http.Redirect(w, r, trelloAuthUrl.String(), http.StatusFound)
}

// Deauthenticate destroys a user session
func (a AuthController) Deauthenticate(w http.ResponseWriter, r *http.Request) {
	session, err := a.Store.Get(r, constants.SessionName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// This is the recommended way of deleting a session
	// http://www.gorillatoolkit.org/pkg/sessions#FilesystemStore.MaxAge
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
