package middlewares

import (
	"fmt"
	"gallo/app/constants"
	"gallo/lib"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
)

const TOKEN_TIMEOUT = 60 // minutes

// The ActivityMiddleware is responsible for keeping track of what users are
// actively using the application, by writing user tokens to a shared redis
// store with a somewhat short expiration.
// In order to provide an accurate picture of currently active tokens, this
// middleware *must* be ordered before any other possibly intercepting
// middlewares.
type ActivityMiddleware struct {
	store *sessions.CookieStore
	cache lib.CacheProvider
}

func NewActivityMiddleware(store *sessions.CookieStore) *ActivityMiddleware {
	return &ActivityMiddleware{
		store,
		lib.RedisClientDecorator{
			redis.NewClient(&redis.Options{
				Addr: lib.MustGetEnv("REDIS_ADDR"),
			}),
		},
	}
}

func (a ActivityMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := a.store.Get(r, constants.SessionName)

		if token, ok := session.Values[constants.TrelloTokenSessionKey]; ok {
			a.cache.Set(
				fmt.Sprintf("tokens:%s", token.(string)),
				fmt.Sprintf("%d", time.Now().Unix()),
				time.Duration(10)*time.Second,
			)
		}

		next.ServeHTTP(w, r)
	})
}
