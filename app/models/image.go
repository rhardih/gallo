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

func (i Image) GetPreviews() (previews []trello.AttachmentPreview) {
	// Sometimes the last preview image is rotated 90 degrees onto its side, so
	// that one is excluded
	previews = i.Previews[:len(i.Previews)-1]

	// Sort by width ascending
	sort.Slice(previews, func(j, k int) bool {
		return previews[j].Width < previews[k].Width
	})

	return
}
