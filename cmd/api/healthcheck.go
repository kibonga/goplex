package main

import (
	"net/http"
)

type SystemInfo struct {
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

type Healthcheck struct {
	Status     string      `json:"status"`
	SystemInfo *SystemInfo `json:"system_info"`
}

func (app *app) healtcheckHandler(w http.ResponseWriter, r *http.Request) {

	data := &Healthcheck{
		Status: "available",
		SystemInfo: &SystemInfo{
			Environment: app.config.env,
			Version:     app.version,
		},
	}

	err := app.writeJsonToStream(w, http.StatusOK, data, nil)
	if err != nil {
		http.Error(w, "failed to process the request", http.StatusInternalServerError)
	}
}
