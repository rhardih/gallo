package models

import (
	"net/http"
	"testing"

	"github.com/adlio/trello"
	"github.com/jarcoal/httpmock"
	"gotest.tools/assert"
)

func TestNewList(t *testing.T) {
	t.Run("valid list", func(t *testing.T) {
		trelloList := &trello.List{}
		list, err := NewList(trelloList)
		assert.NilError(t, err)
		assert.Equal(t, list.TrelloList, trelloList)
	})

	t.Run("nil list", func(t *testing.T) {
		_, err := NewList(nil)
		assert.ErrorContains(t, err, "List is nil")
	})
}

func TestGetList(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/236?",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/lists-000.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/237?",
		httpmock.NewStringResponder(http.StatusNotFound, "list not found"),
	)
	defer httpmock.Reset()

	_, err := GetList(defaultContext, "236")
	assert.NilError(t, err)

	_, err = GetList(defaultContext, "237")
	assert.ErrorContains(t, err, "list not found")
}

func TestListGetCards(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/234?",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/lists-000.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/234/cards?attachments=true",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-000.json"]),
	)
	defer httpmock.Reset()

	// There's no NewList in adlio/trello, so this is the only way to get a list
	// where the correct client is set
	trelloList, err := trelloClient.GetList("234", trello.Defaults())
	assert.NilError(t, err)

	t.Run("Returns error if there's no trello list", func(t *testing.T) {
		list := &List{}

		_, err := list.GetCards()
		assert.ErrorContains(t, err, "TrelloList is nil")
	})

	t.Run("Gets list cards", func(t *testing.T) {
		list, err := NewList(trelloList)
		assert.NilError(t, err)

		cards, err := list.GetCards()
		assert.NilError(t, err)
		assert.Equal(t, len(cards), 1)
		assert.Equal(t, cards[0].TrelloCard.Name, "Foo")
	})

	t.Run("Only fetches cards once", func(t *testing.T) {
		httpmock.ZeroCallCounters()

		list, err := NewList(trelloList)
		assert.NilError(t, err)

		list.GetCards()
		assert.Equal(t, httpmock.GetTotalCallCount(), 1)

		list.GetCards()
		assert.Equal(t, httpmock.GetTotalCallCount(), 1)
	})
}
