package main

import (
	"context"
	httpapi "job-queue/internal/api/http"
	"job-queue/internal/config"
	"job-queue/internal/logger/handlers/setuplogger"
	"job-queue/internal/logger/sl"
	"job-queue/internal/repository"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	envPath := os.Getenv("ENV_FILE")
	if envPath == "" {
		envPath = ".env"
	}
	godotenv.Load(envPath)

	cfg := config.MustLoad()

	log := setuplogger.New(cfg.Env)

	log.Info("starting job queue", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := repository.New(ctx, cfg.DB.DBUrl)
	if err != nil {
		log.Error("failed to connect", sl.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	log.Info("connected to db")

	router := httpapi.NewRouter(log, storage, cfg)

	log.Info("server starting", slog.String("address", cfg.HTTPServer.Address))

	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		log.Info("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("failed to start server", sl.Err(err))
		os.Exit(1)
	}
}
