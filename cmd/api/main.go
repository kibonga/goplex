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
}

type app struct {
	config  config
	logger  *log.Logger
	version string
	db      *sql.DB
}

const defaultMaxIdleTime int = 1000 * 60 * 15

func main() {
	var cfg config
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	flag.IntVar(&cfg.port, "port", 4000, "Server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "Data Source Name", os.Getenv("GOPLEX_DB_DSN"), "Postrges DSN")
	flag.IntVar(&cfg.db.maxOpenConn, "db-max-open-conns", 25, "Postgres max open connections")
	flag.IntVar(&cfg.db.maxIdleConn, "db-max-idle-conns", 25, "Postgres max idle connections")
	flag.IntVar(&cfg.db.maxIdleTime, "db-max-idle-time", defaultMaxIdleTime, "Postgres max idle time in ms")

	flag.Parse()

	app := &app{
		logger:  logger,
		version: version,
		config:  cfg,
	}

	db, err := openDb(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("database connection pool established\n")
	app.db = db

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("Starting server on addr: %d", app.config.port)
	err = srv.ListenAndServe()
	log.Fatal(err)
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
