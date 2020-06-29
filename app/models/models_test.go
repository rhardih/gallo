package models

import (
	"context"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
	"gallo/app/constants"

	"github.com/adlio/trello"
	"github.com/jarcoal/httpmock"
	"gotest.tools/assert"
)

var (
	testData map[string][]byte

	defaultContext context.Context

	trelloClient *trello.Client
)

func TestMain(m *testing.M) {
	testData = make(map[string][]byte)

	dataFiles, _ := filepath.Glob("testdata/*.json")
	for _, dataFile := range dataFiles {
		data, err := ioutil.ReadFile(dataFile)
		if err != nil {
			log.Fatal(err)
		}

		testData[dataFile] = data
	}

	// Http responses will be mocked, so no need for seting key & token
	trelloClient = trello.NewClient("", "")

	defaultContext = context.WithValue(
		context.Background(),
		constants.TrelloClientContextKey,
		trelloClient,
	)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Initialize RNG
	rand.Seed(time.Now().UnixNano())

	os.Exit(m.Run())
}

func Test_clientFromContext(t *testing.T) {
	bgCtx := context.Background()

	_, err := clientFromContext(bgCtx)
	assert.ErrorContains(t, err, "context value is nil")

	key := constants.TrelloClientContextKey

	in := trelloClient
	ctx := context.WithValue(bgCtx, key, trelloClient)

	out, err := clientFromContext(ctx)
	assert.NilError(t, err)
	assert.Equal(t, in, out)
}
