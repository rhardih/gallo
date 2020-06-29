package controllers

import (
	"gallo/app/models"
	"gallo/app/views"
	"log"
	"net/http"

	"github.com/adlio/trello"
	"github.com/gorilla/mux"
)

type ImagePreviews struct {
	Previews []trello.AttachmentPreview `json:"previews"`
}

type CardsController struct{}

func (e CardsController) Show(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	card, err := models.GetCard(r.Context(), id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	images := card.GetImages()

	data := struct {
		Card            *models.Card
		BackgroundColor string
		Images          []ImagePreviews
		ShowDuration    int
	}{
		Card:            card,
		BackgroundColor: images[0].EdgeColor,
		Images:          make([]ImagePreviews, len(images)),
		ShowDuration:    IMAGE_SHOW_DURATION,
	}

	for i := range images {
		data.Images[i].Previews = images[i].GetPreviews()
	}

	views.Execute(w, r, "cards/show.html.tmpl", data)
}
