package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *app) readIdParam(r *http.Request) (int, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 32)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid id provided")
	}

	return int(id), nil
}
