package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/go-redis/cache/v8"
)

// CachingTransport is an implementation of http.RoundTripper which provides a
// caching wrapper around http.DefaultTransport.RoundTrip.
type CachingTransport struct {
	Cache      RedisCacheProvider
	expiration time.Duration
}

func NewCachingTransport(
	rcp RedisCacheProvider,
	expiration time.Duration,
) *CachingTransport {
	return &CachingTransport{rcp, expiration}
}

// RoundTrip adds caching behaviour to the default http transport, such that if
// a cached http response exists for a given request, that is returned,
// pre-empting a full http request. If a cached response doesn't exist, a
// regular request is sent to the target server and then the response is cached,
// before being retured to the caller.
func (c *CachingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var cachedDump []byte

	err := c.Cache.Get(r.Context(), cacheKey(r), &cachedDump)
	if err == nil {
		log.Println(fmt.Sprintf("Cache hit for %s", r.URL.Path))

		reader := bufio.NewReader(bytes.NewBuffer(cachedDump))

		return http.ReadResponse(reader, r)
	}

	log.Println(fmt.Sprintf("Cache miss for %s", r.URL.Path))

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	err = c.Cache.Set(&cache.Item{
		Ctx:   r.Context(),
		Key:   cacheKey(r),
		Value: dump,
		TTL:   c.expiration,
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// cacheKey is the full url for the request including query params with key and
// token
func cacheKey(r *http.Request) string {
	return r.URL.String()
}
