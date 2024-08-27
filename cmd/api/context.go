package main

import (
	"context"
	"net/http"

	"goplex.kibonga/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *app) contextSetUser(r *http.Request, u *data.User) *http.Request {
	ctx := context.WithValue(context.Background(), userContextKey, u)
	return r.WithContext(ctx)
}

func (app *app) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
