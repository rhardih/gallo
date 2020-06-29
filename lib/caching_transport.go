package lib

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/go-redis/redis"
)

type CacheProvider interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
}

type cache struct {
	Client *redis.Client
}

func (c cache) Get(key string) (string, error) {
	//log.Println(fmt.Sprintf("CachingTransport Get(%s)", key))
	val, err := c.Client.Get(key).Result()
	if err == nil {
		return val, nil
	}

	return "", errors.New("key not found in cache")
}

func (c cache) Set(key, value string, expiration time.Duration) error {
	//log.Println(fmt.Sprintf("CachingTransport Set(%s)", cacheKey(r)))
	err := c.Client.Set(key, value, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

// CachingTransport is an implementation of http.RoundTripper which provides a
// caching wrapper around http.DefaultTransport.RoundTrip.
type CachingTransport struct {
	expiration time.Duration
	Cache      CacheProvider
}

func NewCachingTransport(expiration time.Duration) *CachingTransport {
	client := redis.NewClient(&redis.Options{
		Addr: MustGetEnv("REDIS_ADDR"),
	})

	return &CachingTransport{expiration, &cache{client}}
}

// RoundTrip adds caching behaviour to the default http transport, such that if
// a cached http response exists for a given request, that is returned,
// pre-empting a full http request. If a cached response doesn't exist, a
// regular request is sent to the target server and then the response is cached,
// before being retured to the caller.
func (c *CachingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if val, err := c.Cache.Get(cacheKey(r)); err == nil {
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

	err = c.Cache.Set(cacheKey(r), string(buf), c.expiration)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func cacheKey(r *http.Request) string {
	return r.URL.String()
}
