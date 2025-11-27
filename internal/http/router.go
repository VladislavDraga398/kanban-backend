package http

import (
	"net/http"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/board"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
	"github.com/VladislavDraga398/kanban-backend/internal/http/handlers"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Deps — зависимости HTTP-слоя (репозитории, сервисы и т.п.).
type Deps struct {
	UserRepo  user.Repository
	BoardRepo board.Repository
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
	// Board-handlers
	boardHandler := handlers.NewBoardHandler(deps.BoardRepo)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		// Защищенные ручки - нужно, чтобы middleware.Auth положил корректный UserID в контекст.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth)

			r.Route("/boards", func(r chi.Router) {
				r.Get("/", boardHandler.List)          // Список всех досок.
				r.Post("/", boardHandler.Create)       // Создание новой доски.
				r.Get("/{id}", boardHandler.Get)       // Получение доски по её ID.
				r.Put("/{id}", boardHandler.Update)    // Обновление названия доски.
				r.Delete("/{id}", boardHandler.Delete) // Удаление доски.
			})
		})
	})

	return r
}
