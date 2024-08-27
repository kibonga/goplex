package main

import (
	"errors"
	"net/http"
	"time"

	"goplex.kibonga/internal/data"
	"goplex.kibonga/internal/validator"
)

func (app *app) createAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Extract credentials from request
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.decodeJson(r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate request (email and password) using validator
	v := validator.New()
	data.ValidateEmail(v, req.Email)
	data.ValidatePasswordPlaintext(v, req.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Check if user exist and fetch user from DB
	user, err := app.models.Users.GetByEmail(req.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Validate user creds (password and email)
	ok, err := user.Password.Matches(req.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !ok {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Create new token
	token, err := app.models.Tokens.New(user.Id, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send back auth token as json response
	err = app.writeJson(w, http.StatusOK, payload{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
