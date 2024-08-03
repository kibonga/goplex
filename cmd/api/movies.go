package main

import (
	"net/http"
	"time"

	"goplex.kibonga/internal/data"
)

func (app *app) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	data := &data.Movie{
		Id:        id,
		CreatedAt: time.Now(),
		Title:     "Romul",
		Year:      int32(time.Now().Year()),
		Runtime:   150,
		Genres:    []string{"Sci-Fi", "Horror", "Thriller"},
	}

	if err := app.writeJsonToStream(w, http.StatusOK, payload{"movie": data}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

type MovieCreateRequest struct {
	Title   string   `json:"title"`
	Year    int32    `json:"year"`
	Runtime int32    `json:"runtime"`
	Genres  []string `json:"genres"`
}

func (app *app) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var movie MovieCreateRequest

	err := app.decodeJson(r, &movie)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	defer r.Body.Close()
}

func (app *app) createMovieHandlerMarshal(w http.ResponseWriter, r *http.Request) {
	var movie MovieCreateRequest
	err := app.unmarshalJson(r, &movie)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	defer r.Body.Close()
}
