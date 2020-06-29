package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adlio/trello"
)

type Card struct {
	Name       string
	CoverImage Image
	List       *List

	TrelloCard *trello.Card
}

// Create a new card and attach parent list, coverImage and attachments if
// present
func NewCard(trelloCard *trello.Card) (*Card, error) {
	var coverAttachment *trello.Attachment

	for _, attachment := range trelloCard.Attachments {
		if attachment.ID == trelloCard.IDAttachmentCover {
			coverAttachment = attachment
			break
		}
	}

	if coverAttachment == nil {
		return nil, errors.New(fmt.Sprintf(
			"No cover attachment for card (%s, %s)",
			trelloCard.ID,
			trelloCard.Name,
		))
	}

	coverImage := NewImage(coverAttachment)

	if trelloCard.List != nil {
		list, err := NewList(trelloCard.List)
		if err != nil {
			return nil, err
		}

		return &Card{trelloCard.Name, coverImage, list, trelloCard}, nil
	}

	return &Card{trelloCard.Name, coverImage, nil, trelloCard}, nil
}

func GetCard(ctx context.Context, id string) (*Card, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	args := trello.Defaults()
	args["attachments"] = "true"
	args["list"] = "true"

	var trelloCard *trello.Card

	trelloCard, err = client.GetCard(id, args)
	if err != nil {
		return nil, err
	}

	card, err := NewCard(trelloCard)
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (c Card) GetImages() []Image {
	images := make([]Image, len(c.TrelloCard.Attachments))

	for i := range c.TrelloCard.Attachments {
		images[i] = NewImage(c.TrelloCard.Attachments[i])
	}

	return images
}

func (c Card) Date() *time.Time {
	if c.TrelloCard.Due != nil {
		return c.TrelloCard.Due
	} else {
		return c.TrelloCard.DateLastActivity
	}
}

func (c Card) DueDate() *time.Time {
	return c.TrelloCard.Due
}

func (c Card) ID() string {
	return c.TrelloCard.ID
}

func (c Card) PluralName() string {
	return "cards"
}
