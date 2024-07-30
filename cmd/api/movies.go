package main

import (
	"fmt"
	"net/http"
)

func (app *app) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(fmt.Sprintf("Hello from show movie handler for movie with id=%d", id)))
}

func (app *app) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from create movie handler"))
}
