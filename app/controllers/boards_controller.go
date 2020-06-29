package controllers

import (
	"gallo/app/models"
	"gallo/app/views"
	"gallo/lib"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type BoardsController struct{}

func (c BoardsController) Index(w http.ResponseWriter, r *http.Request) {
	defer lib.Track(lib.RunningTime("BoardsController.Index"))

	boards, err := models.GetValidBoards(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	views.Execute(w, r, "boards/index.html.tmpl", boards)
}

func (c BoardsController) Shuffle(w http.ResponseWriter, r *http.Request) {
	card, err := func() (card *models.Card, err error) {
		id, ok := mux.Vars(r)["id"]

		// If an id is present, narrow selection to cards from that specific board
		if ok {
			board, err := models.GetBoard(r.Context(), id)
			if err != nil {
				return nil, err
			}

			return board.GetRandomCard()
		} else {
			// Get all boards on the account
			boards, err := models.GetValidBoards(r.Context())
			if err != nil {
				return nil, err
			}

			return models.GetRandomCard(r.Context(), boards)
		}
	}()

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
