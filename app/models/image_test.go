package models

import (
	"testing"

	"github.com/adlio/trello"
	"gotest.tools/assert"
)

func createTestImage() Image {
	return NewImage(
		&trello.Attachment{
			Previews: []trello.AttachmentPreview{
				trello.AttachmentPreview{
					Width:  100,
					Height: 200,
					URL:    "https://example.com/100.jpg",
				},
				trello.AttachmentPreview{
					Width:  200,
					Height: 300,
					URL:    "https://example.com/200.jpg",
				},
				trello.AttachmentPreview{},
			},
		},
	)
}

// -----------------------------------------------------------------------------

func TestNewImage(t *testing.T) {
	image := createTestImage()

	assert.Equal(t, len(image.GetPreviews()), 2)
}

func TestGetters(t *testing.T) {
	image := createTestImage()

	assert.Equal(t, image.GetWidth(), 200)
	assert.Equal(t, image.GetHeight(), 300)
	assert.Equal(t, image.GetURL(), "https://example.com/200.jpg")
}

func TestGetPreviews(t *testing.T) {
	t.Run("No attachments, no panic", func(t *testing.T) {
		image := Image{
			&trello.Attachment{
				Previews: []trello.AttachmentPreview{},
			},
		}

		previews := image.GetPreviews()

		assert.Equal(t, len(previews), 0)
	})
}
