package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"goplex.kibonga/internal/data"
	"goplex.kibonga/internal/jsonlog"
	"goplex.kibonga/internal/mailer"
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
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type app struct {
	config  config
	logger  *jsonlog.Logger
	version string
	models  data.Models
	mailer  mailer.Mailer
	wg      sync.WaitGroup
}

const defaultMaxIdleTime int = 1000 * 60 * 15

func main() {
	var cfg config
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	flag.IntVar(&cfg.port, "port", 4000, "Server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "Postrges DSN")

	flag.IntVar(&cfg.db.maxOpenConn, "db-max-open-conns", 25, "Postgres max open connections")
	flag.IntVar(&cfg.db.maxIdleConn, "db-max-idle-conns", 25, "Postgres max idle connections")
	flag.IntVar(&cfg.db.maxIdleTime, "db-max-idle-time", defaultMaxIdleTime, "Postgres max idle time in ms")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "0902612716084e", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "c5865616978195", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "GOPLEX <no-reply@goplex.net>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Cors trusted origins", func(s string) error {
		cfg.cors.trustedOrigins = strings.Fields(s)
		return nil
	})

	flag.Parse()

	fmt.Printf("cors-trusted-origins=%v\n", cfg.cors.trustedOrigins)

	db, err := openDb(&cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &app{
		logger:  logger,
		version: version,
		config:  cfg,
		models:  data.NewModels(db),
		mailer:  mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	if err := app.serve(); err != nil {
		logger.PrintFatal(err, nil)
	}
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
