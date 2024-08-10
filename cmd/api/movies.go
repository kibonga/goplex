package main

import (
	"errors"
	"fmt"
	"net/http"

	"goplex.kibonga/internal/data"
	"goplex.kibonga/internal/validator"
)

func (app *app) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJsonToStream(w, http.StatusOK, payload{"movie": movie}, nil); err != nil {
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

type MovieUpdateRequest struct {
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

	err = app.models.Movies.Insert(&movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.Id))

	err = app.writeJson(w, http.StatusCreated, payload{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
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

func (app *app) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var req MovieUpdateRequest
	err = app.decodeJson(r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	movie.Title = req.Title
	movie.Year = req.Year
	movie.Runtime = req.Runtime
	movie.Genres = req.Genres

	validator := validator.New()
	data.ValidateMovie(validator, movie)
	if !validator.Valid() {
		app.failedValidationResponse(w, r, validator.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("v1/movies/%d", movie.Id))

	err = app.writeJson(w, http.StatusOK, payload{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the id
	id, err := app.readIdParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Delete movie from DB
	err = app.models.Movies.Delete(int(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write back response
	err = app.writeJson(w, http.StatusNoContent, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
