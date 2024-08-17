package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"goplex.kibonga/internal/data"
	"goplex.kibonga/internal/jsonlog"
)

const version string = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn         string
		maxOpenConn int
		maxIdleConn int
		maxIdleTime int
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type app struct {
	config  config
	logger  *jsonlog.Logger
	version string
	models  data.Models
}

const defaultMaxIdleTime int = 1000 * 60 * 15

func main() {
	var cfg config
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	flag.IntVar(&cfg.port, "port", 4000, "Server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "Data Source Name", os.Getenv("GOPLEX_DB_DSN"), "Postrges DSN")

	flag.IntVar(&cfg.db.maxOpenConn, "db-max-open-conns", 25, "Postgres max open connections")
	flag.IntVar(&cfg.db.maxIdleConn, "db-max-idle-conns", 25, "Postgres max idle connections")
	flag.IntVar(&cfg.db.maxIdleTime, "db-max-idle-time", defaultMaxIdleTime, "Postgres max idle time in ms")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	db, err := openDb(&cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	app := &app{
		logger:  logger,
		version: version,
		config:  cfg,
		models:  data.NewModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(logger, "", 0),
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.env,
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}

func openDb(cfg *config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConn)
	db.SetMaxIdleConns(cfg.db.maxIdleConn)
	db.SetConnMaxIdleTime(time.Millisecond * time.Duration(cfg.db.maxIdleTime))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
