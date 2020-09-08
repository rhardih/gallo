package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/assert"
)

func TestWebhooksMiddleware(t *testing.T) {
	wm, _ := NewWebhooksMiddleware()
	handlerToTest := wm.Handler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))

	t.Run("Initialises without error", func(t *testing.T) {
		_, err := NewWebhooksMiddleware()
		assert.NilError(t, err)
	})

	t.Run("Fails if IP is not allowed", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "http://testing", nil)

		handlerToTest.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("handler returned wrong status code: got %v want %v", recorder.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("Succeeds if IP is allowed", func(t *testing.T) {
		ips := []string{
			"107.23.104.115",
			"107.23.149.70",
			"54.152.166.250",
			"54.164.77.56",
			"54.209.149.230",
			"18.234.32.224",
			"18.234.32.225",
			"18.234.32.226",
			"18.234.32.227",
			"18.234.32.228",
			"18.234.32.229",
			"18.234.32.230",
			"18.234.32.231",
			"18.234.32.232",
			"18.234.32.233",
			"18.234.32.234",
			"18.234.32.235",
			"18.234.32.236",
			"18.234.32.237",
			"18.234.32.238",
			"18.234.32.239",
		}

		for _, ip := range ips {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "http://testing", nil)

			req.RemoteAddr = ip

			handlerToTest.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", recorder.Code, http.StatusOK)
			}
		}
	})
}
