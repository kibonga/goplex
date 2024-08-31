package main

import (
	"net/http"
	"time"

	"goplex.kibonga/internal/data"
)

func (app *app) fooHandler(w http.ResponseWriter, r *http.Request) {

	go func() {
		time.Sleep(time.Second * 3)

		app.writeJson(w, http.StatusAccepted, payload{"bar": "foo"}, nil)
	}()

	app.writeJson(w, http.StatusOK, payload{"foo": "bar"}, nil)
}

func (app *app) fooPermissionsHandlerGetAllForUser(w http.ResponseWriter, r *http.Request) {
	perms, err := app.models.Permissions.GetAllForUser(6)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJson(w, http.StatusOK, payload{"permissions": perms}, nil)
}

func (app *app) tokenHandler(w http.ResponseWriter, r *http.Request) {
	userID := 6
	ttl := time.Duration(time.Second * 300)
	scope := data.ScopeActivation

	app.models.Tokens.New(int64(userID), ttl, scope)
}
