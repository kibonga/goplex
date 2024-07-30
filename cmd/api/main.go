package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version string = "1.0.0"

type config struct {
	port int
	env  string
}

type app struct {
	config  config
	logger  *log.Logger
	version string
}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	port := flag.Int("port", 4000, "server port")
	env := flag.String("env", "development", "Environment (development|staging|production)")
	flag.Parse()

	app := &app{
		config: config{
			port: *port,
			env:  *env,
		},
		logger:  logger,
		version: version,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healtcheckHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("Starting server on addr: %d", app.config.port)
	err := srv.ListenAndServe()
	log.Fatal(err)
}
