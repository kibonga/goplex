package main

import (
	"errors"
	"fmt"
	"net/http"

	"goplex.kibonga/internal/data"
	"goplex.kibonga/internal/validator"
)

func validSortVals() *[]string {
	return &[]string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
}

type MovieCreateRequest struct {
	Title   string       `json:"title"`
	Year    int32        `json:"year"`
	Runtime data.Runtime `json:"runtime"`
	Genres  []string     `json:"genres"`
}

type MovieUpdateRequest struct {
	Title   *string       `json:"title"`
	Year    *int32        `json:"year"`
	Runtime *data.Runtime `json:"runtime"`
	Genres  []string      `json:"genres"`
}

type ListMoviesRequest struct {
	Title   string
	Genres  []string
	Filters *data.Filters
}

func listMoviesReq() *ListMoviesRequest {
	return &ListMoviesRequest{Filters: &data.Filters{}}
}

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

	if req.Title != nil {
		movie.Title = *req.Title
	}

	if req.Year != nil {
		movie.Year = *req.Year
	}

	if req.Runtime != nil {
		movie.Runtime = *req.Runtime
	}

	if req.Genres != nil {
		movie.Genres = req.Genres
	}

	validator := validator.New()
	data.ValidateMovie(validator, movie)
	if !validator.Valid() {
		app.failedValidationResponse(w, r, validator.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
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
	id, err := app.readIdParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

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

	err = app.writeJson(w, http.StatusNoContent, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List movies handler")
	req := listMoviesReq()
	urlVals := r.URL.Query()

	v := validator.New()

	req.Title = app.readStr(urlVals, "title", "")
	req.Genres = app.readCSV(urlVals, "genres", []string{})
	req.Filters.PageSize = app.readInt(urlVals, "page_size", v, 20)
	req.Filters.Page = app.readInt(urlVals, "page", v, 1)
	req.Filters.Sort = app.readStr(urlVals, "sort", "id")
	req.Filters.ValidSortValues = *validSortVals()

	if data.ValidateFilters(v, req.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	movies, metadata, err := app.models.Movies.GetAll(req.Title, req.Genres, req.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJson(w, http.StatusOK, payload{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
