package gallo

import (
	"log"
	"net/http"
	"time"
	"gallo/app/controllers"
)

type Application struct {
	Addr string
}

func (app Application) Run() {
	srv := &http.Server{
		Handler:      controllers.NewRouter(),
		Addr:         app.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
