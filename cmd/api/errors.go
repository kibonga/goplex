package main

import (
	"fmt"
	"net/http"
)

func (app *app) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

func (app *app) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	payload := &payload{
		"error": message,
	}

	err := app.writeJson(w, status, payload, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *app) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem processing your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *app) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *app) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}