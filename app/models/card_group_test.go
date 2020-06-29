package models

import (
	"testing"
	"time"

	"github.com/adlio/trello"
	"gotest.tools/assert"
)

func timeFromYear(year string) *time.Time {
	t, _ := time.Parse("2006", year)
	return &t
}

// -----------------------------------------------------------------------------

func TestNewCardGroupings(t *testing.T) {
	cards := []*Card{
		&Card{
			TrelloCard: &trello.Card{
				Name:             "foo",
				DateLastActivity: timeFromYear("2018"),
			},
		},
		&Card{
			TrelloCard: &trello.Card{
				Name:             "bar",
				DateLastActivity: timeFromYear("2019"),
			},
		},
		&Card{
			TrelloCard: &trello.Card{
				Name:             "baz",
				DateLastActivity: timeFromYear("2018"),
			},
		},
	}

	groupings := NewCardGroups(cards)

	assert.Equal(t, len(groupings), 2)
	assert.Equal(t, groupings[0].Year, 2019)
	assert.Equal(t, len(groupings[0].Cards), 1)
	assert.Equal(t, groupings[1].Year, 2018)
	assert.Equal(t, len(groupings[1].Cards), 2)
}
