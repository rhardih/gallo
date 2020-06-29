package models

import "sort"

// CardGroup represents a logical grouping of cards that belongs to the same
// calendar year.
type CardGroup struct {
	Year  int
	Cards []*Card
}

func NewCardGroups(cards []*Card) (cardGroups []CardGroup) {
	cardsByYear := make(map[int][]*Card)

	for _, card := range cards {
		cardsByYear[card.Date().Year()] = append(cardsByYear[card.Date().Year()], card)
	}

	for year, cards := range cardsByYear {
		group := CardGroup{
			Year:  year,
			Cards: cards,
		}

		cardGroups = append(cardGroups, group)
	}

	// Sort by year descending
	sort.Slice(cardGroups, func(i, j int) bool {
		return cardGroups[i].Year > cardGroups[j].Year
	})

	return
}
