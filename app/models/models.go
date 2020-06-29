package models

import (
	"context"
	"errors"
	"gallo/app/constants"

	"github.com/adlio/trello"
)

type Model interface {
	ID() string
	PluralName() string
}

func clientFromContext(ctx context.Context) (*trello.Client, error) {
	value := ctx.Value(constants.TrelloClientContextKey)
	if value == nil {
		return nil, errors.New("context value is nil")
	}

	return value.(*trello.Client), nil
}
