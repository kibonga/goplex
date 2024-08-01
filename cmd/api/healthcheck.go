package main

import (
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

	err := app.writeJsonToStream(w, http.StatusOK, data, nil)
	if err != nil {
		http.Error(w, "failed to process the request", http.StatusInternalServerError)
	}
}
