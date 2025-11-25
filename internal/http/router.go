package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
	"github.com/VladislavDraga398/kanban-backend/internal/http/handlers"
)

// Deps — зависимости HTTP-слоя (репозитории, сервисы и т.п.).
type Deps struct {
	UserRepo user.Repository
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	// Healthcheck — чтобы k8s/docker могли проверять живой ли сервис.
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Auth-handlers
	authHandler := handlers.NewAuthHandler(deps.UserRepo)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})
	})

	return r
}
