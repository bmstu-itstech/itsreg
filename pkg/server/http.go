package server

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/bmstu-itstech/itsreg-bots/pkg/jwtauth"
	"github.com/bmstu-itstech/itsreg-bots/pkg/logs"
	"github.com/bmstu-itstech/itsreg-bots/pkg/logs/sl"
)

const corsMaxAge = 300

func RunHTTPServer(createHandler func(router chi.Router) http.Handler) {
	RunHTTPServerOnAddr(":"+os.Getenv("PORT"), createHandler)
}

func RunHTTPServerOnAddr(addr string, createHandler func(router chi.Router) http.Handler) {
	log := logs.DefaultLogger()

	apiRouter := chi.NewRouter()
	setMiddlewares(apiRouter, log)

	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api/v2", createHandler(apiRouter))

	log.Info("starting: HTTP server", "addr", addr)

	//nolint:gosec // Не имеется возможности переехать на другой фреймворк
	err := http.ListenAndServe(addr, rootRouter)
	if err != nil {
		log.Error("Unable to start HTTP server")
		panic(err)
	}
}

func setMiddlewares(router *chi.Mux, log *slog.Logger) {
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(sl.NewLoggerMiddleware(log))
	router.Use(middleware.Recoverer)

	addCorsMiddleware(router)

	router.Use(
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
	)
	router.Use(middleware.NoCache)
	router.Use(jwtauth.HTTPMiddleware)
}

func addCorsMiddleware(router *chi.Mux) {
	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ";")
	if len(allowedOrigins) == 0 {
		return
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           corsMaxAge,
	})
	router.Use(corsMiddleware.Handler)
}
