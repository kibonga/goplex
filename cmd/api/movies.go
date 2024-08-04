package main

import (
	"fmt"
	"net/http"
	"time"

	"goplex.kibonga/internal/data"
	"goplex.kibonga/internal/validator"
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
	Title   string       `json:"title"`
	Year    int32        `json:"year"`
	Runtime data.Runtime `json:"runtime"`
	Genres  []string     `json:"genres"`
}

func (app *app) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var req MovieCreateRequest

	err := app.decodeJson(r, &req)
	defer r.Body.Close()
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var movie data.Movie = data.Movie{
		Title:   req.Title,
		Year:    req.Year,
		Runtime: req.Runtime,
		Genres:  req.Genres,
	}

	v := validator.New()
	data.ValidateMovie(v, &movie)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", req)
}

func (app *app) createMovieHandlerMarshal(w http.ResponseWriter, r *http.Request) {
	var req MovieCreateRequest
	err := app.unmarshalJson(r, &req)
	defer r.Body.Close()
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var movie data.Movie = data.Movie{
		Title:   req.Title,
		Year:    req.Year,
		Runtime: req.Runtime,
		Genres:  req.Genres,
	}

	v := validator.New()
	data.ValidateMovie(v, &movie)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", req)
}
