package main

import (
	"encoding/json"
	"net/http"
)

type Healthcheck struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

func (app *app) healtcheckHandler(w http.ResponseWriter, r *http.Request) {

	data := &Healthcheck{
		Status:      "available",
		Environment: app.config.env,
		Version:     app.version,
	}

	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "failed to process the request", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}
