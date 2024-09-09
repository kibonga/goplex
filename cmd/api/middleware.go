package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"goplex.kibonga/internal/data"
	"goplex.kibonga/internal/validator"
)

func (app *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *app) limitRate(next http.Handler) http.Handler {

	if !app.config.limiter.enabled {
		return next
	}

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ip, _, err := net.SplitHostPort(r.RemoteAddr)

		// if err != nil {
		// 	app.serverErrorResponse(w, r, err)
		// }
		ip := realip.FromRequest(r)

		mu.Lock()

		// IP first time
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
		}

		clients[ip].lastSeen = time.Now()

		// Check if ip exceeded rate limit
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		// Unlock the mutex
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (app *app) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()
		data.ValidatePlaintextToken(v, token)

		if !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetByToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *app) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *app) requireAuthenticatedUser(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *app) requirePermissions(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		permissions, err := app.models.Permissions.GetAllForUser(user.Id)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireActivatedUser(fn)
}

func (app *app) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")
		fmt.Println("This is inside enable cors")

		if origin != "" && app.includes(origin, app.config.cors.trustedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)

			fmt.Println("This is inside origin")

			// Check if request is a preflight request
			// 1. http method is options
			// 2. has origin header
			// 3. has access-control-request-method header
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

				fmt.Println("This is inside preflight")

				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *app) metrics(next http.Handler) http.Handler {
	totalReqRecv := expvar.NewInt("total_request_received")
	totalRespSent := expvar.NewInt("total_reponses_sent")
	totalProcessingTimeMicrosec := expvar.NewInt("total_processing_time_microsec")
	totalRespSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// start := time.Now()
		totalReqRecv.Add(1)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		totalRespSent.Add(1)

		// duration := time.Since(start).Microseconds()
		totalProcessingTimeMicrosec.Add(metrics.Duration.Microseconds())

		totalRespSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
