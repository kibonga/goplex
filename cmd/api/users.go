package main

import (
	"errors"
	"net/http"
	"time"

	"goplex.kibonga/internal/data"
	"goplex.kibonga/internal/validator"
)

func (app *app) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var userReq struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.decodeJson(r, &userReq)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var user data.User = data.User{
		Name:      userReq.Name,
		Email:     userReq.Email,
		Activated: false,
	}

	// Set password from plaintext
	err = user.Password.Set(userReq.Password)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate
	v := validator.New()
	data.ValidateUser(v, &user)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Register user
	err = app.models.Users.Insert(&user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	app.background(func() {
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}
	})

	// Handle response
	err = app.writeJson(w, http.StatusCreated, payload{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) fooHandler(w http.ResponseWriter, r *http.Request) {

	go func() {
		time.Sleep(time.Second * 3)
		app.writeJson(w, http.StatusAccepted, payload{"bar": "foo"}, nil)
	}()

	app.writeJson(w, http.StatusOK, payload{"foo": "bar"}, nil)
}

func (app *app) tokenHandler(w http.ResponseWriter, r *http.Request) {
	userID := 6
	ttl := time.Duration(time.Second * 300)
	scope := data.ScopeActivation

	app.models.Tokens.New(int64(userID), ttl, scope)
}
