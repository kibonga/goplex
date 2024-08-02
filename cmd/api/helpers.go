package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type payload map[string]interface{}

func (app *app) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 32)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid id provided")
	}

	return id, nil
}

func (app *app) writeJson(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	payload = append(payload, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(payload)

	return nil
}

func (app *app) writeJsonToStream(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	for k, v := range headers {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}
