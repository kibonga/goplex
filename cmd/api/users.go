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

	err = user.Password.Set(userReq.Password)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateUser(v, &user)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

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

	token, err := app.models.Tokens.New(user.Id, time.Duration(time.Hour*24*3), data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.PlainText,
			"userID":          user.Id,
		}
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}
	})

	err = app.writeJson(w, http.StatusCreated, payload{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	err := app.decodeJson(r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var token *data.Token = &data.Token{
		PlainText: req.Token,
	}

	v := validator.New()
	data.ValidateToken(v, token)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByToken(data.ScopeActivation, token.PlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			app.editConflictResponse(w, r)
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r) // Logically this case can never happen, but I included it anyways
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteTokensForUser(data.ScopeActivation, user.Id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJson(w, http.StatusOK, payload{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
