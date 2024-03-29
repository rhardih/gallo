package middlewares

import (
	"bytes"
	"errors"
	"fmt"
	"gallo/app/constants"
	"gallo/lib"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/go-redis/cache/v8"
	"github.com/gorilla/sessions"
)

// CachingMiddleware is a simple response cache. Responses are recorded by a
// httptest.ResponseRecorder, marshalled with msgpack and stored in Redis. The
// cache key for each response, is simply a concatenation of the url and a
// unique session token.
type CachingMiddleware struct {
	cache      lib.RedisCacheProvider
	store      *sessions.CookieStore
	sessionKey string
	blacklist  []string // urls matching these patterns will not be cached
}

// NewCachingMiddleware creates a new middleware with a cookie session store.
// The blacklist should contain a set of regular expressions that matches URLs
// which should not be cached.
func NewCachingMiddleware(
	cache lib.RedisCacheProvider,
	store *sessions.CookieStore,
	blacklist []string,
) *CachingMiddleware {
	return &CachingMiddleware{
		cache,
		store,
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

			err := c.cache.Once(&cache.Item{
				Key:   fmt.Sprintf("%s-%s", token.(string), r.URL.String()),
				Value: recorder,
				Do: func(*cache.Item) (interface{}, error) {
					rec := httptest.NewRecorder()
					next.ServeHTTP(rec, r)

					hit = "False"

					result := rec.Result()
					isSuccess := result.StatusCode >= 200 && result.StatusCode <= 299

					if isSuccess {
						return lib.NewSlicedResponseRecorder(rec), nil
					} else {
						var sb strings.Builder
						buf := new(bytes.Buffer)
						buf.ReadFrom(result.Body)

						sb.WriteString(fmt.Sprintf(
							"Request for '%s' failed with status code '%d'.\n",
							r.URL.String(),
							result.StatusCode,
						))

						sb.WriteString(fmt.Sprintf("Response body:\n\n%s", buf.String()))

						return nil, errors.New(sb.String())
					}
				},
			})

			if err != nil {
				log.Println(err.Error())

				next.ServeHTTP(w, r)
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
