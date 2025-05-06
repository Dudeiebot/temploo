package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/hostrouter"
	"github.com/go-chi/httprate"

	"github.com/dudeiebot/ad-ly/config"
	"github.com/dudeiebot/ad-ly/controllers"
	"github.com/dudeiebot/ad-ly/helpers"
	customMiddleware "github.com/dudeiebot/ad-ly/middlewares"
)

func Routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	hr := hostrouter.New()

	apiHost := config.GetApiHost()

	hr.Map(apiHost, apiRoutes())

	r.Mount("/", hr)
	return r
}

func apiRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*"},
		AllowedMethods:   []string{"POST"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "stripe-signature"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// rate limit by IP Address
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Use(customMiddleware.AcceptJson)

	r.Use(customMiddleware.ValidateJson)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(helpers.Message("404 Not Found"))
		return
	})

	r.Group(func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(helpers.Response("ok", "API IS HEALTHY"))

			return
		})
	})

	r.Group(func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", controllers.Register)
			r.Get("/verify-email", controllers.VerifyUser)
			r.Post("/login", controllers.LoginUser)
			r.Post("/forgot-password", controllers.ForgotPassword)
			r.Post("/post-forgot", controllers.PostForgot)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthenticateUser)
		r.Route("/user", func(r chi.Router) {
			r.Get("/get-user", controllers.GetUser)
		})
	})

	return r
}
