package http

import (
	"net/http"
	"time"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/board"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/column"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
	"github.com/VladislavDraga398/kanban-backend/internal/http/handlers"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Deps содержит зависимости HTTP-слоя.
type Deps struct {
	UserRepo   user.Repository
	BoardRepo  board.Repository
	ColumnRepo column.Repository
	TaskRepo   task.Repository
	JWTSecret  string
	JWTTTL     time.Duration
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	authHandler := handlers.NewAuthHandler(deps.UserRepo, deps.JWTSecret, deps.JWTTTL)
	boardHandler := handlers.NewBoardHandler(deps.BoardRepo)
	columnHandler := handlers.NewColumnHandler(deps.ColumnRepo)
	taskHandler := handlers.NewTaskHandler(deps.TaskRepo)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth([]byte(deps.JWTSecret)))

			r.Route("/boards", func(r chi.Router) {
				r.Get("/", boardHandler.List)
				r.Post("/", boardHandler.Create)
				r.Get("/{id}", boardHandler.Get)
				r.Put("/{id}", boardHandler.Update)
				r.Delete("/{id}", boardHandler.Delete)

				r.Route("/{board_id}/columns", func(r chi.Router) {
					r.Get("/", columnHandler.List)
					r.Post("/", columnHandler.Create)

					r.Put("/{column_id}", columnHandler.Update)
					r.Delete("/{column_id}", columnHandler.Delete)

					r.Route("/{column_id}/tasks", func(r chi.Router) {
						r.Get("/", taskHandler.List)
						r.Post("/", taskHandler.Create)

						r.Put("/{task_id}", taskHandler.Update)
						r.Delete("/{task_id}", taskHandler.Delete)
					})
				})

				r.Route("/{board_id}/tasks", func(r chi.Router) {
					r.Patch("/{task_id}/move", taskHandler.Move)
				})
			})
		})
	})
	return r
}
