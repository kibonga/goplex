package main

import "net/http"

func (app *app) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var userReq struct {
		name     string `json:"name"`
		email    string `json:"email"`
		password string `json:"password"`
	}

	err := app.decodeJson(r, &userReq)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
}
