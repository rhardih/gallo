package models

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/adlio/trello"
	"github.com/jarcoal/httpmock"
	"gotest.tools/assert"
)

func Test_boardsCardsBatchURLs(t *testing.T) {
	boards := []*Board{
		&Board{TrelloBoard: &trello.Board{ID: "0"}},
		&Board{TrelloBoard: &trello.Board{ID: "1"}},
	}

	// Empty case
	assert.Equal(t, "",
		boardsCardsBatchURLs([]*Board{}))

	// One board
	assert.Equal(t, "/boards/0/cards",
		boardsCardsBatchURLs(boards[0:1]),
	)

	// Multiple boards
	assert.Equal(t, "/boards/0/cards,/boards/1/cards",
		boardsCardsBatchURLs(boards),
	)
}

func Test_getBoardCardsBatchLimited(t *testing.T) {
	t.Run("Too many boards", func(t *testing.T) {
		_, err := getBoardCardsBatchLimited(defaultContext, make([]*Board, 11))

		assert.Error(t, err, "Max 10 boards should be supplied")
	})

	t.Run("Two boards", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/batch?urls=%2Fboards%2F1234%2Fcards%2C%2Fboards%2F4567%2Fcards",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/batch-000.json"]),
		)
		defer httpmock.Reset()

		// IDs match the ones in the httpmock data file
		boards := []*Board{
			&Board{TrelloBoard: &trello.Board{ID: "1234"}},
			&Board{TrelloBoard: &trello.Board{ID: "4567"}},
		}

		cards, err := getBoardCardsBatchLimited(defaultContext, boards)

		assert.NilError(t, err)
		assert.Equal(t, httpmock.GetTotalCallCount(), 1)

		// 404s should be ignored, so only the three cards from the two boards are
		// expected
		assert.Equal(t, len(cards), 3)
		assert.Equal(t, cards[0].Name, "Lorem")
		assert.Equal(t, cards[1].Name, "Ipsum")
		assert.Equal(t, cards[2].Name, "Dolor")
	})
}

func Test_getBoardCardsBatch(t *testing.T) {
	t.Run("More than ten boards, should result in two http requests", func(t *testing.T) {
		makeUrl := func(query string) string {
			baseUrl, _ := url.Parse("https://api.trello.com/1/batch")
			params := url.Values{}
			params.Add("urls", query)
			baseUrl.RawQuery = params.Encode()
			return baseUrl.String()
		}

		endpoints := []string{
			"/boards/1/cards",
			"/boards/2/cards",
			"/boards/3/cards",
			"/boards/4/cards",
			"/boards/5/cards",
			"/boards/6/cards",
			"/boards/7/cards",
			"/boards/8/cards",
			"/boards/9/cards",
			"/boards/10/cards",
		}

		// Specifically what response, as long as it is valid, comes back from
		// this request, doesn't really matter. We only want to test, that requesting
		// more than ten endpoints, gets split up into batches of ten.
		httpmock.RegisterResponder(
			"GET",
			makeUrl(strings.Join(endpoints, ",")),
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/batch-000.json"]),
		)

		httpmock.RegisterResponder(
			"GET",
			makeUrl("/boards/11/cards"),
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/batch-000.json"]),
		)

		boards := []*Board{
			&Board{TrelloBoard: &trello.Board{ID: "1"}},
			&Board{TrelloBoard: &trello.Board{ID: "2"}},
			&Board{TrelloBoard: &trello.Board{ID: "3"}},
			&Board{TrelloBoard: &trello.Board{ID: "4"}},
			&Board{TrelloBoard: &trello.Board{ID: "5"}},
			&Board{TrelloBoard: &trello.Board{ID: "6"}},
			&Board{TrelloBoard: &trello.Board{ID: "7"}},
			&Board{TrelloBoard: &trello.Board{ID: "8"}},
			&Board{TrelloBoard: &trello.Board{ID: "9"}},
			&Board{TrelloBoard: &trello.Board{ID: "10"}},
			&Board{TrelloBoard: &trello.Board{ID: "11"}},
		}

		_, err := getBoardCardsBatch(defaultContext, boards)

		assert.NilError(t, err)

		assert.Equal(t, httpmock.GetTotalCallCount(), 2)
	})
}

func Test_cardOnSubscribedList(t *testing.T) {
	boards := []*Board{
		&Board{
			Lists: []*List{
				&List{
					TrelloList: &trello.List{
						ID:         "0",
						Name:       "lorem",
						Subscribed: true,
					},
				},
				&List{
					TrelloList: &trello.List{
						ID:   "1",
						Name: "ipsum",
					},
				},
			},
		},
	}

	goodCard := &trello.Card{IDList: "0"}
	badCard := &trello.Card{IDList: "1"}

	assert.Assert(t,
		cardOnSubscribedList(goodCard, boards),
		"Card is on a subscribed list",
	)
	assert.Assert(t,
		!cardOnSubscribedList(badCard, boards),
		"Card is not on a subscribed list",
	)
}

func TestGetRandomCard(t *testing.T) {
	t.Run("No cards", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/batch?urls=%2Fboards%2F1234%2Fcards%2C%2Fboards%2F4567%2Fcards",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/batch-000.json"]),
		)
		defer httpmock.Reset()

		// IDs match the ones in the httpmock data file
		boards := []*Board{
			&Board{TrelloBoard: &trello.Board{ID: "1234"}},
			&Board{TrelloBoard: &trello.Board{ID: "4567"}},
		}

		_, err := GetRandomCard(defaultContext, boards)

		assert.Error(t, err, "No cards found for GetRandomCard")
	})

	t.Run("Only one list subscribed, single card", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/batch?urls=%2Fboards%2F1234%2Fcards%2C%2Fboards%2F1235%2Fcards",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/batch-001.json"]),
		)
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/cards/12?attachments=true&list=true",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-004.json"]),
		)
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/cards/14?attachments=true&list=true",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-005.json"]),
		)
		defer httpmock.Reset()

		// Matchup ids with mocked json
		boards := []*Board{
			&Board{
				Lists: []*List{
					&List{
						TrelloList: &trello.List{
							ID:         "123",
							Subscribed: true,
						},
					},
					&List{
						TrelloList: &trello.List{
							ID:         "124",
							Subscribed: false,
						},
					},
				},
				TrelloBoard: &trello.Board{ID: "1234"},
			},
			&Board{
				Lists: []*List{
					&List{
						TrelloList: &trello.List{
							ID:         "125",
							Subscribed: true,
						},
					},
				},
				TrelloBoard: &trello.Board{ID: "1235"},
			},
		}

		card, err := GetRandomCard(defaultContext, boards)

		assert.NilError(t, err)

		// No point in asserting GetCallCountInfo() here, since the card request can
		// have two different ids
		assert.Equal(t, httpmock.GetTotalCallCount(), 2)
		assert.Assert(t, card.Name == "Lorem" || card.Name == "Dolor")
	})
}
