package controllers

import (
	"gallo/app/controllers/middlewares"
	"gallo/app/views"
	"gallo/lib"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

const IMAGE_SHOW_DURATION = 15 // TODO: This should be a setting

func init() {
	encKey := []byte(lib.MustGetEnv("SESSION_ENC_KEY"))
	authKey := []byte(lib.MustGetEnv("SESSION_AUTH_KEY"))

	store = sessions.NewCookieStore(encKey, authKey)
	views.Store = store
}

func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(middlewares.LoggingMiddleware)

	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": lib.MustGetEnv("REDIS_ADDR"),
		},
	})

	requestCache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	trelloClientMiddleware := middlewares.NewTrelloClientMiddleware(
		requestCache,
		lib.MustGetEnv("TRELLO_KEY"),
		store,
	)

	responseCache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	blacklist := []string{"shuffle$"}
	cachingMiddleware := middlewares.NewCachingMiddleware(
		responseCache,
		store,
		blacklist,
	)

	authorizedRouter := router.NewRoute().Subrouter()
	authorizedRouter.Use(cachingMiddleware.Handler)
	authorizedRouter.Use(trelloClientMiddleware.Handler)

	applicationController := ApplicationController{}
	authController := AuthController{store}
	listsController := ListsController{}
	boardsController := BoardsController{}
	cardsController := CardsController{}

	authorizedRouter.HandleFunc("/boards", boardsController.Index)
	authorizedRouter.HandleFunc("/shuffle", boardsController.Shuffle)
	authorizedRouter.HandleFunc("/boards/{id}/shuffle", boardsController.Shuffle)

	authorizedRouter.HandleFunc("/lists/{id}/shuffle", listsController.Shuffle)
	authorizedRouter.PathPrefix("/lists/{id}").HandlerFunc(listsController.Show)
	authorizedRouter.PathPrefix("/cards/{id}").HandlerFunc(cardsController.Show)

	anonymousRouter := router.NewRoute().Subrouter()
	anonymousRouter.HandleFunc("/auth", authController.Authenticate).
		Methods("GET").
		Queries("token", "{token:[0-9a-f]{64}}")
	anonymousRouter.HandleFunc("/auth", authController.Authorize).
		Methods("GET").
		Queries("trello", "{trello:return|stay}")
	anonymousRouter.HandleFunc("/auth", authController.Show).
		Methods("GET")
	anonymousRouter.HandleFunc("/auth", authController.Deauthenticate).
		Methods("POST")

	// Static assets etc.
	router.PathPrefix("/").HandlerFunc(applicationController.RootHandler)

	return router
}
