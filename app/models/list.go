package models

import (
	"context"
	"errors"
	"fmt"
	"gallo/lib"
	"math/rand"

	"github.com/adlio/trello"
)

type List struct {
	Name string

	Cards []*Card

	TrelloList *trello.List
}

func NewList(trelloList *trello.List) (*List, error) {
	if trelloList == nil {
		return nil, errors.New("List is nil")
	}

	list := &List{}
	list.Name = trelloList.Name
	list.TrelloList = trelloList

	return list, nil
}

func GetList(ctx context.Context, id string) (list *List, err error) {
	defer lib.Track(lib.RunningTime("GetList"))

	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	trelloList, err := client.GetList(id, trello.Defaults())
	if err != nil {
		return nil, err
	}

	return NewList(trelloList)
}

// Memoized slice of cards on a list, with attachments sideloaded.
func (l *List) GetCards() ([]*Card, error) {
	defer lib.Track(lib.RunningTime(fmt.Sprintf("list.GetCards - %s", l.Name)))

	if l.Cards == nil {
		args := trello.Defaults()
		args["attachments"] = "true"

		if l.TrelloList == nil {
			return nil, errors.New("TrelloList is nil")
		}

		trelloCards, err := l.TrelloList.GetCards(args)
		if err != nil {
			return nil, err
		}

		l.Cards = make([]*Card, 0)

		for i := range trelloCards {
			// Re-attach parent list, since that isn't sideloaded for this endpoint
			trelloCards[i].List = l.TrelloList

			// Skip if there's no cover, since then there's no image attachments
			// neither
			if trelloCards[i].IDAttachmentCover == "" {
				continue
			}

			card, err := NewCard(trelloCards[i])
			if err != nil {
				return nil, err
			}

			l.Cards = append(l.Cards, card)
		}
	}

	return l.Cards, nil
}

func (l *List) GetRandomCard() (*Card, error) {
	defer lib.Track(lib.RunningTime("GetCards"))

	cards, err := l.GetCards()
	if err != nil {
		return nil, err
	}

	return cards[rand.Intn(len(cards))], nil
}

func (l List) ID() string {
	return l.TrelloList.ID
}

func (List) PluralName() string {
	return "lists"
}
