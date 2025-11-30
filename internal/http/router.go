package http

import (
	"net/http"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/board"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/column"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
	"github.com/VladislavDraga398/kanban-backend/internal/http/handlers"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Deps — зависимости HTTP-слоя (репозитории, сервисы и т.п.).
type Deps struct {
	UserRepo   user.Repository
	BoardRepo  board.Repository
	ColumnRepo column.Repository
	TaskRepo   task.Repository
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	// Healthcheck — чтобы k8s/docker могли проверять живой ли сервис.
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	authHandler := handlers.NewAuthHandler(deps.UserRepo)
	boardHandler := handlers.NewBoardHandler(deps.BoardRepo)
	columnHandler := handlers.NewColumnHandler(deps.ColumnRepo)
	taskHandler := handlers.NewTaskHandler(deps.TaskRepo)

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

				r.Route("/{board_id}/columns", func(r chi.Router) {
					r.Get("/", columnHandler.List)    // GET /api/v1/boards/{board_id}/columns
					r.Post("/", columnHandler.Create) // POST /api/v1/boards/{board_id}/columns

					r.Put("/{column_id}", columnHandler.Update)
					r.Delete("/{column_id}", columnHandler.Delete)

					// задачи внутри колонки
					r.Route("/{column_id}/tasks", func(r chi.Router) {
						r.Get("/", taskHandler.List)    // GET  /api/v1/boards/{board_id}/columns/{column_id}/tasks
						r.Post("/", taskHandler.Create) // POST /api/v1/boards/{board_id}/columns/{column_id}/tasks

						r.Put("/{task_id}", taskHandler.Update)
						r.Delete("/{task_id}", taskHandler.Delete)
					})
				})
			})
		})
	})
	return r
}
