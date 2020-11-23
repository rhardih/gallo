package middlewares

import (
	"context"
	"fmt"
	"gallo/app/constants"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elliotchance/redismock/v8"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
)

type ClockMock struct {
	Time time.Time
}

func (m ClockMock) Now() time.Time {
	return m.Time
}

func TestHandler(t *testing.T) {
	// Setup
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "http://testing", nil)
	store := sessions.NewCookieStore(
		securecookie.GenerateRandomKey(16),
		securecookie.GenerateRandomKey(16),
	)
	session, _ := store.Get(request, constants.SessionName)
	token := "somerandomstring"
	session.Values[constants.TrelloTokenSessionKey] = token
	clockMock := ClockMock{time.Now()}
	nextHandlerCalled := false

	t.Run("Set succeeds", func(t *testing.T) {
		redisMock := redismock.NewMock()
		redisMock.On(
			"Set",
			request.Context(),
			fmt.Sprintf("tokens:%s", token),
			mock.Anything,
			mock.Anything,
		).Return(redis.NewStatusCmd(context.TODO(), "", nil))

		// Invocation
		activityMiddleware := NewActivityMiddleware(redisMock, clockMock, store)
		handlerFunc := activityMiddleware.Handler(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) { nextHandlerCalled = true },
		))
		handlerFunc.ServeHTTP(recorder, request)

		// Assertions
		redisMock.AssertNumberOfCalls(t, "Set", 1)
		redisMock.AssertCalled(
			t,
			"Set",
			request.Context(),
			fmt.Sprintf("tokens:%s", token),
			fmt.Sprintf("%d", clockMock.Now().Unix()),
			10*time.Second,
		)
		assert.Assert(t, nextHandlerCalled, "Next handler wasn't called")
	})

	t.Run("Set fails", func(t *testing.T) {
		// TODO: There's currently no way to set err on a StatusCmd
		// https://github.com/go-redis/redis/blob/v8/command.go#L411
	})
}
