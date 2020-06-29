package models

import (
	"context"
	"errors"
	"fmt"
	"gallo/lib"
	"math/rand"
	"regexp"

	"github.com/adlio/trello"
)

// Board is a decorator for *trello.Board, which only exposes needed members
// data members, as well as hoists some methods to be funtion members in order
// to make it easier to stub out expected behaviour from adlio/trello.
type Board struct {
	BackgroundBrightness string
	Name                 string

	Lists []*List

	TrelloBoard *trello.Board
}

func NewBoard(trelloBoard *trello.Board) (*Board, error) {
	if trelloBoard == nil {
		return nil, errors.New("Board is nil")
	}

	board := &Board{}
	board.BackgroundBrightness = trelloBoard.Prefs.BackgroundBrightness
	board.Name = trelloBoard.Name

	board.TrelloBoard = trelloBoard

	if trelloBoard.Lists != nil {
		board.Lists = make([]*List, 0)

		for i := range trelloBoard.Lists {
			list, err := NewList(trelloBoard.Lists[i])
			if err != nil {
				return nil, err
			}

			board.Lists = append(board.Lists, list)
		}
	}

	return board, nil
}

// Memoized slice of lists on a board.
func (b *Board) GetLists() ([]*List, error) {
	if b.Lists == nil {
		if b.TrelloBoard.Lists == nil {
			trelloLists, err := b.TrelloBoard.GetLists(trello.Defaults())
			if err != nil {
				return nil, err
			}

			b.TrelloBoard.Lists = trelloLists
		}

		for i := range b.TrelloBoard.Lists {
			list, err := NewList(b.TrelloBoard.Lists[i])
			if err != nil {
				return nil, err
			}

			b.Lists = append(b.Lists, list)
		}
	}

	return b.Lists, nil
}

// The subset of lists on a board which follows the criteria of both being
// subscribed to, and having at least a single card in them.
func (b Board) GetValidLists() ([]*List, error) {
	lists := make([]*List, 0)

	boardLists, err := b.GetLists()
	if err != nil {
		return nil, err
	}

	for i := range boardLists {
		if !boardLists[i].TrelloList.Subscribed {
			continue
		}

		cards, err := boardLists[i].GetCards()
		if err != nil {
			return nil, err
		}

		if len(cards) > 0 {
			lists = append(lists, boardLists[i])
		}
	}

	return lists, nil
}

func (b Board) GetList(id string) (*List, error) {
	boardLists, err := b.GetLists()
	if err != nil {
		return nil, err
	}

	for i := range boardLists {
		if boardLists[i].TrelloList.ID == id {
			return boardLists[i], nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Failed to find List with id: %s", id))
}

// Returns cards on a board, which belongs to a subscribed list. Board lists
// should be sideloaded beforehand.
func (b Board) GetCards() ([]*Card, error) {
	if b.Lists == nil {
		return nil, errors.New("Board lists not loaded")
	}

	args := trello.Defaults()
	args["card_attachments"] = "true"

	trelloCards, err := b.TrelloBoard.GetCards(args)
	if err != nil {
		return nil, err
	}

	isOnSubscribedList := func(card *trello.Card, lists []*List) bool {
		for i := range lists {
			if card.IDList == lists[i].TrelloList.ID && lists[i].TrelloList.Subscribed {
				return true
			}
		}
		return false
	}

	cards := []*Card{}

	for _, trelloCard := range trelloCards {
		if !isOnSubscribedList(trelloCard, b.Lists) {
			continue
		}

		card, err := NewCard(trelloCard)
		if err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	return cards, nil
}

func (b Board) GetRandomCard() (*Card, error) {
	lists, err := b.GetValidLists()
	if err != nil {
		return nil, err
	}

	if len(lists) == 0 {
		return nil, errors.New(fmt.Sprintf("No lists in board %s", b.TrelloBoard.Name))
	}

	randomList := lists[rand.Intn(len(lists))]

	return randomList.GetRandomCard()
}

func (b Board) ID() string {
	return b.TrelloBoard.ID
}

func (b Board) PluralName() string {
	return "boards"
}

func GetBoard(ctx context.Context, id string) (*Board, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	args := trello.Defaults()
	args["lists"] = "all"

	board, err := client.GetBoard(id, args)
	if err != nil {
		return nil, err
	}

	match, err := regexp.MatchString("gallo", board.Desc)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, errors.New("Board doesn't have \"gallo\" in description")
	}

	return NewBoard(board)
}

// All boards of a member, with lists sideloaded
func GetBoards(ctx context.Context) ([]*Board, error) {
	defer lib.Track(lib.RunningTime("GetBoards"))

	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Sideload lists as a nested resource by default
	args := trello.Defaults()
	args["lists"] = "all"

	trelloBoards, err := client.GetMyBoards(args)
	if err != nil {
		return nil, err
	}

	var boards []*Board

	for i := range trelloBoards {
		board, err := NewBoard(trelloBoards[i])
		if err != nil {
			return nil, err
		}

		boards = append(boards, board)
	}

	return boards, nil
}

// Get boards which has at least one valid list and contains the word "gallo"
// somewhere in the description
func GetValidBoards(ctx context.Context) ([]*Board, error) {
	boards, err := GetBoards(ctx)
	if err != nil {
		return nil, err
	}

	filteredBoards := make([]*Board, 0)

	for i := range boards {
		// Skip boards without "gallo" in the description
		match, err := regexp.MatchString("gallo", boards[i].TrelloBoard.Desc)
		if err != nil {
			return nil, err
		}

		if !match {
			continue
		}

		lists, err := boards[i].GetValidLists()
		if err != nil {
			return nil, err
		}

		if len(lists) > 0 {
			boards[i].Lists = lists

			filteredBoards = append(filteredBoards, boards[i])
		}
	}

	return filteredBoards, nil
}
