package main

import (
	"job-queue/internal/config"
	del "job-queue/internal/http-server/handlers/job-queue/delete"
	"job-queue/internal/http-server/handlers/job-queue/get"
	"job-queue/internal/http-server/handlers/job-queue/save"
	"job-queue/internal/http-server/middleware/logger"
	"job-queue/internal/lib/logger/handlers/setuplogger"
	"job-queue/internal/lib/logger/sl"
	"job-queue/internal/repository/postgres"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	cfg := config.MustLoad()

	log := setuplogger.New(cfg.Env)

	log.Info("starting job queue", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := postgres.New(cfg.DB.DBUrl)
	if err != nil {
		log.Error("failed to connect", sl.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	log.Info("connected to db")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)

	router.Route("/job", func(r chi.Router) {
		r.Use(middleware.BasicAuth("job-queue", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{jobID}", del.New(log, storage))
	})

	router.Get("/job/{jobID}", get.New(log, storage))

	log.Info("server starting", slog.String("address", cfg.HTTPServer.Address))

	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}
