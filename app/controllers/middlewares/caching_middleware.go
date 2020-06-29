package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"gallo/app/constants"
	"gallo/lib"

	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/vmihailenco/msgpack"
)

// CachingMiddleware is a simple response cache. Responses are recorded by a
// httptest.ResponseRecorder, marshalled with msgpack and stored in Redis. The
// cache key for each response, is simply a concatenation of the url and a
// unique session token.
type CachingMiddleware struct {
	store      *sessions.CookieStore
	codec      *cache.Codec
	sessionKey string
	blacklist  []string // urls matching these patterns will not be cached
}

// NewCachingMiddleware creates a new middleware with a cookie session store.
// The blacklist should contain a set of regular expressions that matches URLs
// which should not be cached.
func NewCachingMiddleware(store *sessions.CookieStore, blacklist []string) *CachingMiddleware {
	return &CachingMiddleware{
		store,
		&cache.Codec{
			Redis: redis.NewRing(&redis.RingOptions{
				Addrs: map[string]string{
					"server1": lib.MustGetEnv("REDIS_ADDR"),
				},
			}),

			Marshal:   msgpack.Marshal,
			Unmarshal: msgpack.Unmarshal,
		},
		constants.TrelloTokenSessionKey,
		blacklist,
	}
}

func (c CachingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("cache-control") == "no-cache" {
			next.ServeHTTP(w, r)
			return
		}

		for i := range c.blacklist {
			pattern := c.blacklist[i]
			matched, err := regexp.Match(pattern, []byte(r.URL.String()))
			if err != nil {
				log.Println(matched, err)
			}
			if matched {
				next.ServeHTTP(w, r)
				return
			}
		}

		session, _ := c.store.Get(r, constants.SessionName)

		if token, ok := session.Values[c.sessionKey]; ok {
			recorder := new(lib.SlicedResponseRecorder)
			hit := "True"

			err := c.codec.Once(&cache.Item{
				Key:    fmt.Sprintf("%s-%s", token.(string), r.URL.String()),
				Object: recorder,
				Func: func() (interface{}, error) {
					rec := httptest.NewRecorder()
					next.ServeHTTP(rec, r)

					hit = "False"

					return lib.NewSlicedResponseRecorder(rec), nil
				},
			})

			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			result := recorder.Result()

			for k, v := range result.Header {
				w.Header()[k] = v
			}

			w.Header().Set("Cache-Hit", hit)

			w.WriteHeader(result.StatusCode)
			w.Write(recorder.Body)
		} else {
			log.Println("No token found")
			next.ServeHTTP(w, r)
		}
	})
}
