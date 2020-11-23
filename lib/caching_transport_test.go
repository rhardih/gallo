package lib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/elliotchance/redismock/v8"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
)

func bodyToString(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	return buf.String()
}

func TestRoundTrip(t *testing.T) {
	expiration := time.Minute // Doesn't matter for this test
	timestamp := "2006-01-02 15:04:05"

	cachedContent := "foo"

	cachedResponse := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(cachedContent))),
	}

	cachedResponseBuf, err := httputil.DumpResponse(cachedResponse, true)
	if err != nil {
		t.Error(err)
	}

	freshContent := "bar"

	serverResponse := &http.Response{
		StatusCode: 200,
		ProtoMinor: 1,
		ProtoMajor: 1,
		Header: map[string][]string{
			"Content-Type": []string{"text/plain; charset=utf-8"},
			"Date":         []string{timestamp},
		},
		Body:          ioutil.NopCloser(bytes.NewReader([]byte(freshContent))),
		ContentLength: 3,
	}

	serverResponseBuf, err := httputil.DumpResponse(serverResponse, true)
	if err != nil {
		t.Error(err)
	}

	// http test server that constructs a simple but properly formatted  http
	// response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Avoid automatically generated Date header
		w.Header().Set("Date", timestamp)
		fmt.Fprint(w, freshContent)
	}))
	defer server.Close()

	// An http request against the test server created above
	request := httptest.NewRequest("GET", server.URL, nil)

	t.Run("cache hit", func(t *testing.T) {
		redisMock := redismock.NewMock()
		transport := NewCachingTransport(redisMock, expiration)
		redisMock.On(
			"Get",
			request.Context(),
			server.URL,
		).Return(redis.NewStringResult(string(cachedResponseBuf), nil))

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
			redisMock := redismock.NewMock()
			transport := NewCachingTransport(redisMock, expiration)
			redisMock.On(
				"Get",
				request.Context(),
				server.URL,
			).Return(redis.NewStringResult("", errors.New("")))
			redisMock.On(
				"Set",
				request.Context(),
				server.URL,
				mock.Anything,
				mock.Anything,
			).Return(
				redis.NewStatusCmd(request.Context(), "", nil),
			)

			response, err := transport.RoundTrip(request)
			if err != nil {
				t.Error(err)
			}

			expected := freshContent
			actual := bodyToString(response.Body)

			if expected != actual {
				t.Errorf("Expected response '%s', actually got '%s'", expected, actual)
			}
		})

		t.Run("sets cache content", func(t *testing.T) {
			redisMock := redismock.NewMock()
			transport := NewCachingTransport(redisMock, expiration)
			redisMock.On(
				"Get",
				request.Context(),
				server.URL,
			).Return(redis.NewStringResult("", errors.New("")))
			redisMock.On(
				"Set",
				request.Context(),
				server.URL,
				mock.Anything,
				mock.Anything,
			).Return(
				redis.NewStatusCmd(request.Context(), "", nil),
			)

			_, err := transport.RoundTrip(request)
			if err != nil {
				t.Error(err)
			}

			redisMock.AssertNumberOfCalls(t, "Set", 1)
      redisMock.AssertCalled(
        t,
        "Set",
        request.Context(),
        server.URL,
        string(serverResponseBuf),
        expiration,
      )
		})
	})
}
