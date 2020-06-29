package models

import (
	"net/http"
	"strings"
	"testing"

	"github.com/adlio/trello"
	"github.com/jarcoal/httpmock"
	"gotest.tools/assert"
)

func TestNewBoard(t *testing.T) {
	t.Run("Fails if no *trello.Board is provided", func(t *testing.T) {
		_, err := NewBoard(nil)
		assert.ErrorContains(t, err, "Board is nil")
	})

	t.Run("Sets background brightness, name & *trello.Board reference", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/boards/1234?",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-000.json"]),
		)
		defer httpmock.Reset()

		trelloBoard, err := trelloClient.GetBoard("1234", trello.Defaults())
		assert.NilError(t, err)

		board, err := NewBoard(trelloBoard)
		assert.NilError(t, err)

		assert.Equal(t, board.Name, "Foo")
		assert.Equal(t, board.BackgroundBrightness, "dark")
		assert.Equal(t, board.TrelloBoard, trelloBoard)
	})

	t.Run("Sets lists on board", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/boards/1234?",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-001.json"]),
		)
		defer httpmock.Reset()

		trelloBoard, err := trelloClient.GetBoard("1234", trello.Defaults())
		assert.NilError(t, err)

		board, err := NewBoard(trelloBoard)
		assert.NilError(t, err)

		assert.Equal(t, len(board.Lists), 2)
	})
}

func TestBoardGetLists(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/1235?",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-001.json"]),
	)
	defer httpmock.Reset()

	trelloBoard, err := trelloClient.GetBoard("1235", trello.Defaults())
	assert.NilError(t, err)

	board, err := NewBoard(trelloBoard)
	assert.NilError(t, err)

	assert.Equal(t, len(board.Lists), 2)
	assert.Equal(t, board.Lists[0].Name, "Lorem")
	assert.Equal(t, board.Lists[1].Name, "Ipsum")
}

func TestBoardGetValidLists(t *testing.T) {
	// These mocks are here, due to data being fetched as follows:
	//
	// 1. Board is fetched, lists are sideloaded.
	// 2. Each list is checked and only included if it has any cards. This means
	// list cards are fetched for each list.

	// TODO Convert to using the batch endpoint to avoid these mulitple requests

	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/1235?",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-001.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/234/cards?attachments=true",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-000.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/235/cards?attachments=true",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-001.json"]),
	)
	defer httpmock.Reset()

	trelloBoard, err := trelloClient.GetBoard("1235", trello.Defaults())
	assert.NilError(t, err)

	board, err := NewBoard(trelloBoard)
	assert.NilError(t, err)

	t.Run("All lists should be subscribed", func(t *testing.T) {
		lists, err := board.GetValidLists()
		assert.NilError(t, err)

		assert.Assert(t, len(lists) > 0)

		for _, list := range lists {
			assert.Assert(t, list.TrelloList.Subscribed)
		}
	})

	t.Run("All lists should have cards", func(t *testing.T) {
		lists, err := board.GetValidLists()
		assert.NilError(t, err)

		assert.Assert(t, len(lists) > 0)

		for _, list := range lists {
			assert.Assert(t, len(list.Cards) > 0)
		}
	})
}

func TestBoardGetList(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/1234?",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-001.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/234/cards?attachments=true",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-000.json"]),
	)
	defer httpmock.Reset()

	trelloBoard, err := trelloClient.GetBoard("1234", trello.Defaults())
	assert.NilError(t, err)

	board, err := NewBoard(trelloBoard)
	assert.NilError(t, err)

	t.Run("Valid id", func(t *testing.T) {
		list, err := board.GetList("234")
		assert.NilError(t, err)
		assert.Equal(t, list.TrelloList.Name, "Lorem")
	})

	t.Run("Invalid id", func(t *testing.T) {
		_, err := board.GetList("0")
		assert.ErrorContains(t, err, "Failed to find")
	})
}

func TestBoardGetCards(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/1235?",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-001.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/1235/cards?card_attachments=true",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-002.json"]),
	)
	// This extra mock here is due to adlio/trello, pre-emptively looking to see
	// if there's more cards, even if the number of cards are below the limit.
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/1235/cards?before=37&card_attachments=true",
		httpmock.NewStringResponder(http.StatusOK, "[]"),
	)

	trelloBoard, err := trelloClient.GetBoard("1235", trello.Defaults())
	assert.NilError(t, err)

	board, err := NewBoard(trelloBoard)
	assert.NilError(t, err)

	t.Run("Only cards from a subscribed list", func(t *testing.T) {
		cards, err := board.GetCards()
		assert.NilError(t, err)

		assert.Equal(t, len(cards), 1)

		onSubscribeTested := false

		for _, card := range cards {
			for _, list := range board.Lists {
				if card.TrelloCard.IDList == list.TrelloList.ID {
					assert.Assert(t, list.TrelloList.Subscribed)
					onSubscribeTested = true
					break
				}
			}
		}

		assert.Assert(t, onSubscribeTested)
	})
}

func TestBoardGetRandomCard(t *testing.T) {
	t.Run("Fails if board has no lists", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/boards/1234?",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-000.json"]),
		)
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/boards/1234/lists?",
			httpmock.NewStringResponder(http.StatusOK, "[]"),
		)
		defer httpmock.Reset()

		// trelloClient set by TestMain
		trelloBoard, err := trelloClient.GetBoard("1234", trello.Defaults())
		assert.NilError(t, err)

		board, err := NewBoard(trelloBoard)
		assert.NilError(t, err)

		_, err = board.GetRandomCard()
		assert.ErrorContains(t, err, "No lists")
	})

	t.Run("Gets a random card from one of the board lists", func(t *testing.T) {
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/boards/1235?",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-002.json"]),
		)
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/lists/236/cards?attachments=true",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-006.json"]),
		)
		httpmock.RegisterResponder(
			"GET",
			"https://api.trello.com/1/lists/238/cards?attachments=true",
			httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-003.json"]),
		)
		defer httpmock.Reset()

		// trelloClient set by TestMain
		trelloBoard, err := trelloClient.GetBoard("1235", trello.Defaults())
		assert.NilError(t, err)

		board, err := NewBoard(trelloBoard)
		assert.NilError(t, err)

		card, err := board.GetRandomCard()
		assert.NilError(t, err)
		assert.Assert(t, card.TrelloCard.ID == "37" || card.TrelloCard.ID == "38")
	})
}

func TestGetBoard(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/1234?lists=all",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-000.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/1235?lists=all",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-001.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/boards/42?lists=all",
		httpmock.NewStringResponder(http.StatusNotFound, "board not found"),
	)
	defer httpmock.Reset()

	t.Run("Description mismatch", func(t *testing.T) {
		_, err := GetBoard(defaultContext, "1234")
		assert.ErrorContains(t, err, "doesn't have \"gallo\"")
	})

	t.Run("Valid id", func(t *testing.T) {
		board, err := GetBoard(defaultContext, "1235")
		assert.NilError(t, err)
		assert.Equal(t, board.TrelloBoard.Name, "Foo")
	})

	t.Run("Invalid id", func(t *testing.T) {
		_, err := GetBoard(defaultContext, "42")
		assert.ErrorContains(t, err, "board not found")
	})
}

func TestGetBoards(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/members/me/boards",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-003.json"]),
	)

	t.Run("All boards", func(t *testing.T) {
		boards, err := GetBoards(defaultContext)
		assert.NilError(t, err)
		assert.Equal(t, len(boards), 3)
	})

	t.Run("Lists are sideloaded", func(t *testing.T) {
		boards, err := GetBoards(defaultContext)
		assert.NilError(t, err)

		for _, board := range boards {
			assert.Assert(t, board.Lists != nil)
		}
	})
}

func TestGetValidBoards(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/members/me/boards",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/boards-003.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/236/cards?attachments=true",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-006.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/lists/237/cards?attachments=true",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-006.json"]),
	)
	defer httpmock.Reset()

	t.Run("Get boards with 'gallo' in description' and at least one list", func(t *testing.T) {
		boards, err := GetValidBoards(defaultContext)
		assert.NilError(t, err)
		assert.Equal(t, len(boards), 2)

		for _, board := range boards {
			assert.Assert(t, strings.Contains(board.TrelloBoard.Desc, "gallo"))
			assert.Assert(t, len(board.Lists) > 0)
		}
	})
}
