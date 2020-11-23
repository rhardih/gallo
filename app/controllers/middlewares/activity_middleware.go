package middlewares

import (
	"fmt"
	"gallo/app/constants"
	"gallo/lib"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// TODO; Extract to env
const TOKEN_TIMEOUT = 60 // minutes

// The ActivityMiddleware is responsible for keeping track of what users are
// actively using the application, by writing user tokens to a shared redis
// store with a somewhat short expiration.
// In order to provide an accurate picture of currently active tokens, this
// middleware *must* be ordered before any other possibly intercepting
// middlewares.
type ActivityMiddleware struct {
	cache lib.RedisProvider
	clock lib.Clock
	store *sessions.CookieStore
}

func NewActivityMiddleware(cache lib.RedisProvider, clock lib.Clock, store *sessions.CookieStore) *ActivityMiddleware {
	return &ActivityMiddleware{cache, clock, store}
}

func (a ActivityMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := a.store.Get(r, constants.SessionName)

		if token, ok := session.Values[constants.TrelloTokenSessionKey]; ok {
			statusCmd := a.cache.Set(
        r.Context(),
				fmt.Sprintf("tokens:%s", token.(string)),
				fmt.Sprintf("%d", a.clock.Now().Unix()),
				time.Duration(10)*time.Second,
			)

			if statusCmd.Err() != nil {
				log.Println(statusCmd.Err())
			}
		}

		next.ServeHTTP(w, r)
	})
}
