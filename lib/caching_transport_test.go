package lib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockCache struct {
	Hit      bool
	Content  string
	SetValue string
}

func (m MockCache) Get(key string) (string, error) {
	if m.Hit {
		buf := new(bytes.Buffer)

		body := ioutil.NopCloser(bytes.NewBufferString(m.Content))
		t := http.Response{Body: body}
		t.Write(buf)

		return buf.String(), nil
	} else {
		return "", errors.New("")
	}
}

func (m *MockCache) Set(key string, value string, expiration time.Duration) error {
	m.SetValue = value
	return nil
}

func bodyToString(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	return buf.String()
}

func TestRoundTrip(t *testing.T) {
	cachedContent := "foo"
	serverContent := "bar"
	expiration := time.Minute // Doesn't matter for this test
	timestamp := "2006-01-02 15:04:05"

	transport := NewCachingTransport(expiration)

	// http test server that constructs a simple but properly formatted  http
	// response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Date", timestamp) // Avoid automatically generated Date header
		fmt.Fprint(w, serverContent)
	}))
	defer server.Close()

	// An http request against the test server created above
	request := httptest.NewRequest("GET", server.URL, nil)

	t.Run("cache hit", func(t *testing.T) {
		transport.Cache = &MockCache{true, cachedContent, ""}

		response, err := transport.RoundTrip(request)
		if err != nil {
			t.Error(err)
		}

		expected := cachedContent
		actual := bodyToString(response.Body)

		if expected != actual {
			t.Errorf("Expected response '%s', actually got '%s'", expected, actual)
		}
	})

	t.Run("cache miss", func(t *testing.T) {
		t.Run("gets server response", func(t *testing.T) {
			transport.Cache = &MockCache{false, cachedContent, ""}

			response, err := transport.RoundTrip(request)
			if err != nil {
				t.Error(err)
			}

			expected := serverContent
			actual := bodyToString(response.Body)

			if expected != actual {
				t.Errorf("Expected response '%s', actually got '%s'", expected, actual)
			}
		})

		t.Run("sets cache content", func(t *testing.T) {
			cache := &MockCache{false, cachedContent, ""}
			transport.Cache = cache

			_, err := transport.RoundTrip(request)
			if err != nil {
				t.Error(err)
			}

			expected := "HTTP/1.1 200 OK\r\n" +
				"Content-Length: 3\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"Date: 2006-01-02 15:04:05\r\n" +
				"\r\n" +
				"bar"

			actual := cache.SetValue

			if expected != actual {
				t.Errorf("Expected content '%s', actual '%s'", expected, actual)
			}
		})
	})
}
