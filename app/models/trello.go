package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"gallo/lib"

	"github.com/adlio/trello"
)

// This file contains convenience methods that interact with the trello api in
// non-standard ways

// Creates a comma separated string of board cards URLs for use with the Trello
// batch api endpoint
func boardsCardsBatchURLs(boards []*Board) string {
	// Construct the comma separated list of urls, used as the argument to the
	// batch request.
	var urlBuilder strings.Builder

	for i := range boards {
		if i > 0 {
			fmt.Fprintf(&urlBuilder, ",")
		}
		fmt.Fprintf(&urlBuilder, "/boards/%s/cards", boards[i].TrelloBoard.ID)
	}

	return urlBuilder.String()
}

// Retrieve cards from up to ten Boards at once via a batch request
func getBoardCardsBatchLimited(ctx context.Context, boards []*Board) ([]*trello.Card, error) {
	if len(boards) > 10 {
		return nil, errors.New("Max 10 boards should be supplied")
	}

	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	args := trello.Defaults()
	args["urls"] = boardsCardsBatchURLs(boards)

	var responses []map[string]interface{}

	err = client.Get("batch", args, &responses)
	if err != nil {
		return nil, err
	}

	trelloCards := make([]*trello.Card, 0)

	for _, response := range responses {
		if value, ok := response["200"]; ok {
			// This approach is a bit backwards, but since response values can pretty
			// much be anything, this seems to be the easiest way to get back to a
			// slice of structs

			b, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}

			var cards []*trello.Card

			err = json.Unmarshal(b, &cards)

			if err != nil {
				return nil, err
			}

			trelloCards = append(trelloCards, cards...)
		}
	}

	return trelloCards, nil
}

// Retrieve cards from multiple Boards
func getBoardCardsBatch(ctx context.Context, boards []*Board) ([]*trello.Card, error) {
	defer lib.Track(lib.RunningTime("getBoardCardsBatch"))

	trelloCards := make([]*trello.Card, 0)

	// Process boards here in batches of 10
	for i := 0; i < len(boards); i += 10 {
		j := i + 10

		if j > len(boards) {
			j = len(boards)
		}

		cards, err := getBoardCardsBatchLimited(ctx, boards[i:j])
		if err != nil {
			return nil, err
		}

		trelloCards = append(trelloCards, cards...)
	}

	return trelloCards, nil
}

// Determine if a card belongs to a list which is subscribed to (watched)
func cardOnSubscribedList(card *trello.Card, boards []*Board) bool {
	for i := range boards {
		for j := range boards[i].Lists {
			if card.IDList == boards[i].Lists[j].TrelloList.ID &&
				boards[i].Lists[j].TrelloList.Subscribed {
				return true
			}
		}
	}
	return false
}

// Returns a random Card, belonging to a Subscribed List, from any Board. Any
// card fitting the above criteria can be returned with equal probability.
func GetRandomCard(ctx context.Context, boards []*Board) (*Card, error) {
	// Get all cards from provided boards
	trelloCards, err := getBoardCardsBatch(ctx, boards)
	if err != nil {
		return nil, err
	}

	// Only consider cards on lists which are subscribed to
	filteredCards := make([]*trello.Card, 0)

	for i := range trelloCards {
		if cardOnSubscribedList(trelloCards[i], boards) {
			filteredCards = append(filteredCards, trelloCards[i])
		}
	}

	if len(filteredCards) == 0 {
		return nil, errors.New("No cards found for GetRandomCard")
	}

	// Pick one at random
	trelloCard := filteredCards[rand.Intn(len(filteredCards))]

	// Now that we have the card ID, fetch it with attachments sideloaded
	return GetCard(ctx, trelloCard.ID)
}
