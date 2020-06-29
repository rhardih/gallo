package controllers

import (
	"gallo/app/models"
	"gallo/app/views"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type ListsController struct{}

func (c ListsController) Show(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	list, err := models.GetList(r.Context(), id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cards, err := list.GetCards()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cardGroups := models.NewCardGroups(cards)

	data := struct {
		List       *models.List
		CardGroups []models.CardGroup
	}{
		List:       list,
		CardGroups: cardGroups,
	}

	views.Execute(w, r, "lists/show.html.tmpl", data)
}

// Shuffle picks a random card from a list and renders it in the same
// manner as /cards/{id}
func (c ListsController) Shuffle(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	list, err := models.GetList(r.Context(), id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	card, err := list.GetRandomCard()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	images := card.GetImages()

	data := struct {
		Card            *models.Card
		BackgroundColor string
		Images          []ImagePreviews
		AutoRefresh     int
		ShowDuration    int
	}{
		Card:            card,
		BackgroundColor: images[0].EdgeColor,
		Images:          make([]ImagePreviews, len(images)),
		AutoRefresh:     IMAGE_SHOW_DURATION * len(images),
		ShowDuration:    IMAGE_SHOW_DURATION,
	}

	for i := range images {
		data.Images[i].Previews = images[i].GetPreviews()
	}

	views.Execute(w, r, "cards/show.html.tmpl", data)
}
