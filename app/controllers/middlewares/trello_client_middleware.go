package middlewares

import (
	"context"
	"gallo/app/constants"
	"gallo/lib"
	"net/http"
	"time"

	"github.com/adlio/trello"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

var CACHING_TRANSPORT_TIMEOUT = 3

type TrelloClientMiddleware struct {
	store            *sessions.CookieStore
	sessionKey       string
	cachingTransport *lib.CachingTransport
	clientTimeout    time.Duration
}

func NewTrelloClientMiddleware(
	cache lib.RedisCacheProvider,
	key string,
	store *sessions.CookieStore,
) *TrelloClientMiddleware {
	return &TrelloClientMiddleware{
		store, key,
		lib.NewCachingTransport(cache, time.Duration(CACHING_TRANSPORT_TIMEOUT)*time.Hour),
		time.Second * 10,
	}
}

func (c TrelloClientMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := c.store.Get(r, constants.SessionName)

		if token, ok := session.Values[constants.TrelloTokenSessionKey]; ok {
			client := trello.NewClient(c.sessionKey, token.(string))

			logger := logrus.New()
			logger.SetLevel(logrus.DebugLevel)
			client.Logger = logger

			// Replace the default http client used by trello.Client, with a version
			// that caches, as well as times out after ten seconds
			client.Client = &http.Client{
				Transport: c.cachingTransport,
				Timeout:   c.clientTimeout,
			}

			ctx := context.WithValue(r.Context(), constants.TrelloClientContextKey, client)

			w.Header().Set("Logged-In", "True")

			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Redirect(w, r, "/auth", http.StatusFound)
		}
	})
}
