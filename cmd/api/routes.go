package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *app) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healtcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermissions("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermissions("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies/bytes", app.requirePermissions("movies:write", app.createMovieHandlerMarshal))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermissions("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermissions("movies:write", app.deleteMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermissions("movies:read", app.listMoviesHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodGet, "/v1/foo", app.fooHandler)
	router.HandlerFunc(http.MethodGet, "/v1/foo/permissions", app.fooPermissionsHandlerGetAllForUser)
	router.HandlerFunc(http.MethodPost, "/v1/tokens", app.tokenHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthTokenHandler)

	return app.recoverPanic(app.limitRate(app.authenticate(router)))
}
