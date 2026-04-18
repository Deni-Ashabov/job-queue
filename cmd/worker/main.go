package main

import (
	"context"
	"job-queue/internal/config"
	"job-queue/internal/logger/handlers/setuplogger"
	"job-queue/internal/logger/sl"
	"job-queue/internal/repository/postgres"
	"job-queue/internal/worker"
	"log/slog"
	"os"
	"os/signal"

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

	log.Info("starting worker", slog.String("env", cfg.Env))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := postgres.New(ctx, cfg.DB.DBUrl)
	if err != nil {
		log.Error("failed to connect", sl.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	w := worker.New(storage, log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	w.Run(ctx, cfg)
}
