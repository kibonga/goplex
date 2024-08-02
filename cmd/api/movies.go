package main

import (
	"net/http"
	"time"

	"goplex.kibonga/internal/data"
)

func (app *app) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		w.Write([]byte(err.Error()))
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
		http.Error(w, "failed to process request", http.StatusInternalServerError)
		return
	}
}

func (app *app) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from create movie handler"))
}
