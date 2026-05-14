package httpapi

import (
	"job-queue/internal/api/http/handlers/job-queue/get"
	"job-queue/internal/api/http/handlers/job-queue/save"
	"job-queue/internal/api/http/middleware/logger"
	"job-queue/internal/config"
	"job-queue/internal/repository"
	"log/slog"
	"net/http"

	del "job-queue/internal/api/http/handlers/job-queue/delete"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(log *slog.Logger, storage *repository.Storage, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(logger.New(log))
	r.Use(middleware.Recoverer)

	r.Route("/job", func(r chi.Router) {
		r.Use(middleware.BasicAuth("job-queue", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{jobID}", del.New(log, storage))
	})

	r.Get("/job/{jobID}", get.New(log, storage))

	return r
}
