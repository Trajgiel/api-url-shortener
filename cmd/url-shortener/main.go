package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Trajgiel/api-url-shortener/internal/config"
	"github.com/Trajgiel/api-url-shortener/internal/http-server/handlers/redirect"
	"github.com/Trajgiel/api-url-shortener/internal/http-server/handlers/url/delete"
	"github.com/Trajgiel/api-url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/Trajgiel/api-url-shortener/internal/http-server/middleware/logger"
	"github.com/Trajgiel/api-url-shortener/internal/lib/logger/sl"
	"github.com/Trajgiel/api-url-shortener/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envDev  = "dev"
	envTest = "test"
	envProd = "prod"
)

func main() {
	// TODO: init config: cleanenv
	cfg := config.MustLoad()

	// TODO: init logger: slog
	log := setupLogger(cfg.Env)
	log.Info("starting server", slog.String("env", cfg.Env))

	// TODO: init storage: sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed ti init storage", sl.Err(err))
		os.Exit(1)
	}

	// TODO: init router: chi, chi render
	router := chi.NewRouter()

	// TODO: middleware
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)

	// disabled: trims alias by point (e.g. "google.com" -> "google"),
	// why GetURL did not find existing aliases
	// router.Use(middleware.URLFormat)

	// TODO: handlers
	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{id}", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	// TODO: run server
	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed ti start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDev:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envTest:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
