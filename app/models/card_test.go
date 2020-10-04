package models

import (
	"net/http"
	"testing"

	"github.com/adlio/trello"
	"github.com/jarcoal/httpmock"
	"gotest.tools/assert"
)

func TestNewCard(t *testing.T) {
	t.Run("Without attachment cover", func(t *testing.T) {
		_, err := NewCard(&trello.Card{})
		assert.ErrorContains(t, err, "No cover attachment")
	})

	t.Run("With attachment cover", func(t *testing.T) {
		_, err := NewCard(&trello.Card{
			Attachments: []*trello.Attachment{
				&trello.Attachment{ID: "42"},
			},
			IDAttachmentCover: "42",
		})
		assert.NilError(t, err)
	})

	t.Run("Attaches list if present", func(t *testing.T) {
		card, err := NewCard(&trello.Card{
			Attachments: []*trello.Attachment{
				&trello.Attachment{ID: "42"},
			},
			IDAttachmentCover: "42",
			List: &trello.List{
				Name: "Foo",
			},
		})
		assert.NilError(t, err)
		assert.Assert(t, card.List != nil)
		assert.Equal(t, card.List.Name, "Foo")
	})
}

func TestGetCard(t *testing.T) {
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/cards/34?attachments=true&list=true",
		httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-004.json"]),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.trello.com/1/cards/35?attachments=true&list=true",
		httpmock.NewStringResponder(http.StatusNotFound, "card not found"),
	)
	defer httpmock.Reset()

	t.Run("Sideloads list", func(t *testing.T) {
		card, err := GetCard(defaultContext, "34")
		assert.NilError(t, err)
		assert.Assert(t, card.List != nil)
		assert.Equal(t, card.List.Name, "Foo")
	})

	t.Run("Error if not found", func(t *testing.T) {
		_, err := GetCard(defaultContext, "35")
		assert.ErrorContains(t, err, "card not found")
	})
}

func TestGetImages(t *testing.T) {
  httpmock.RegisterResponder(
    "GET",
    "https://api.trello.com/1/cards/34?attachments=true&list=true",
    httpmock.NewBytesResponder(http.StatusOK, testData["testdata/cards-008.json"]),
  )
  defer httpmock.Reset()

  t.Run("Gets attachments with image/ mime types", func(t *testing.T) {
    card, err := GetCard(defaultContext, "34")
    assert.NilError(t, err)

    assert.Equal(t, len(card.TrelloCard.Attachments), 2)

    images := card.GetImages()

    assert.Equal(t, len(images), 1)
    assert.Equal(t, images[0].MimeType, "image/jpeg")
    assert.Equal(t, images[0].Name, "image.jpg")
  })
}
