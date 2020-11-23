package lib

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/elliotchance/redismock/v8"
	"github.com/go-redis/redis/v8"
	"gotest.tools/assert"
)

func TestGetTokens(t *testing.T) {
	testTokens := []string{
		"tokens:lorem",
		"tokens:ipsum",
		"tokens:dolor",
		"tokens:sit",
		"tokens:amet",
		"tokens:consectetur",
		"tokens:adipiscing",
		"tokens:elit",
		"tokens:phasellus",
		"tokens:eget",
		"tokens:erat",
		"tokens:eget",
		"tokens:libero",
		"tokens:facilisis",
		"tokens:blandit",
		"tokens:quisque",
	}

	// This test is made to resemble the nature of redis not giving any guarantees
	// about the specific number of results it will return from a scan. For that
	// reason we do mocks that return randomly sized subslices of the test tokens.

	cursor := uint64(0)
	count := int64(10)
	match := "tokens:*"
	calls := 0

	redisMock := redismock.NewMock()

	for {
		high := rand.Intn(len(testTokens)-int(cursor)) + int(cursor) + 1
		nextCursor := uint64(high)

		if nextCursor >= uint64(len(testTokens)) {
			nextCursor = 0
		}

		redisMock.On("Scan", cursor, match, count).Return(
			redis.NewScanCmdResult(
				testTokens[cursor:high],
				nextCursor,
				nil,
			),
		)

		cursor = nextCursor
		calls++

		if nextCursor == 0 {
			break
		}
	}

	wm := NewWebooksManager(
		redisMock,
		context.Background(),
		10,
	)

	resultTokens := wm.getTokens()

	redisMock.AssertNumberOfCalls(t, "Scan", calls)

	assert.Equal(t, len(resultTokens), len(testTokens), "Expected number of tokens to be equal")

	for i := range testTokens {
		expected := testTokens[i]
		actual := resultTokens[i]

		if expected != actual {
			fmt.Errorf("Expected token '%s', got '%s'", expected, actual)
		}
	}
}
