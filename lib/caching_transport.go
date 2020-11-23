package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

// CachingTransport is an implementation of http.RoundTripper which provides a
// caching wrapper around http.DefaultTransport.RoundTrip.
type CachingTransport struct {
	cache      RedisProvider
	expiration time.Duration
}

func NewCachingTransport(cache RedisProvider, expiration time.Duration) *CachingTransport {
	return &CachingTransport{cache, expiration}
}

// RoundTrip adds caching behaviour to the default http transport, such that if
// a cached http response exists for a given request, that is returned,
// pre-empting a full http request. If a cached response doesn't exist, a
// regular request is sent to the target server and then the response is cached,
// before being retured to the caller.
func (c *CachingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	stringCmd := c.cache.Get(r.Context(), cacheKey(r))
	if val, err := stringCmd.Result(); err == nil {
		log.Println(fmt.Sprintf("Cache hit for %s", r.URL.Path))

		reader := bufio.NewReader(bytes.NewBuffer([]byte(val)))

		return http.ReadResponse(reader, r)
	}

	log.Println(fmt.Sprintf("Cache miss for %s", r.URL.Path))

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	buf, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	statusCmd := c.cache.Set(r.Context(), cacheKey(r), string(buf), c.expiration)
	if statusCmd.Err() != nil {
		return nil, statusCmd.Err()
	}

	return resp, nil
}

func cacheKey(r *http.Request) string {
	return r.URL.String()
}
