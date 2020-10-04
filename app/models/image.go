package models

import (
	"sort"

	"github.com/adlio/trello"
)

type Image struct {
	*trello.Attachment
}

func NewImage(attachment *trello.Attachment) Image {
	image := Image{attachment}

	return image
}

func (i Image) GetWidth() int {
	previews := i.GetPreviews()
	return previews[len(previews)-1].Width
}

func (i Image) GetHeight() int {
	previews := i.GetPreviews()
	return previews[len(previews)-1].Height
}

func (i Image) GetURL() string {
	previews := i.GetPreviews()
	return previews[len(previews)-1].URL
}

func (i Image) GetPreviews() []trello.AttachmentPreview {
	previewsCount := len(i.Previews)

	if previewsCount <= 1 {
		return i.Previews
	}

	// Sometimes the last preview image is rotated 90 degrees onto its side, so
	// that one is excluded
	slicedPreviews := i.Previews[:previewsCount-1]

	// Sort by width ascending
	sort.Slice(slicedPreviews, func(j, k int) bool {
		return slicedPreviews[j].Width < slicedPreviews[k].Width
	})

	return slicedPreviews
}
