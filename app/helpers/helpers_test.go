package helpers

import (
	"testing"
	"gallo/app/models"

	"github.com/adlio/trello"
	"gotest.tools/assert"
)

func createTestImage() models.Image {
	return models.NewImage(
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

func TestSrcSetSizes(t *testing.T) {
	image := createTestImage()

	actual, err := SrcSetSizes(image)
	expected := `srcset="https://example.com/100.jpg 100w, https://example.com/200.jpg 200w" sizes="(max-width: 320px) 100vw, (max-width: 630px) 50vw, 33vw"`
	assert.NilError(t, err)

	assert.Equal(t, actual, expected)
}

func TestShrink(t *testing.T) {
	image := createTestImage()

	width, height := shrink(image)

	assert.Equal(t, width, 2)
	assert.Equal(t, height, 3)
}
