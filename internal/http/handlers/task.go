package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
)

// TaskHandler — хендлер для работы с задачами на доске.
type TaskHandler struct {
	tasks taskStore
}

// NewTaskHandler конструирует хендлер задач.
func NewTaskHandler(tasks taskStore) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

type taskStore interface {
	ListByColumnOwner(ctx context.Context, boardID, columnID, ownerID string) ([]*task.Task, error)
	CreateInColumn(ctx context.Context, task *task.Task, boardID, columnID, ownerID string) error
	Update(ctx context.Context, task *task.Task, ownerID string) error
	Delete(ctx context.Context, id, boardID, columnID, ownerID string) error
	MoveToColumn(ctx context.Context, task *task.Task, columnID, ownerID string) error
}

// createTaskRequest — тело запроса на создание/обновление задачи.
type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// moveTaskRequest — тело запроса на перенос задачи в другую колонку.
type moveTaskRequest struct {
	ColumnID string `json:"column_id"`
}

// taskResponse — DTO, который отдаём наружу.
type taskResponse struct {
	ID          string    `json:"id"`
	BoardID     string    `json:"board_id"`
	ColumnID    string    `json:"column_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// writeTask — маппинг доменной модели в DTO.
func writeTask(t *task.Task) taskResponse {
	return taskResponse{
		ID:          t.ID,
		BoardID:     t.BoardID,
		ColumnID:    t.ColumnID,
		Title:       t.Title,
		Description: t.Description,
		Position:    t.Position,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// List возвращает список задач в колонке.
// GET /api/v1/boards/{board_id}/columns/{column_id}/tasks
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	if boardID == "" || columnID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id and column id are required")
		return
	}

	// Используем твой метод ListByColumnOwner
	tasksList, err := h.tasks.ListByColumnOwner(r.Context(), boardID, columnID, userID)
	if err != nil {
		log.Printf("failed to list tasks: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := make([]taskResponse, 0, len(tasksList))
	for _, t := range tasksList {
		resp = append(resp, writeTask(t))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// Create создаёт задачу в колонке.
// POST /api/v1/boards/{board_id}/columns/{column_id}/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	if boardID == "" || columnID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id and column id are required")
		return
	}

	var req createTaskRequest
	if !httputil.DecodeJSONOrError(w, r, &req, httputil.DefaultMaxJSONBodyBytes) {
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		httputil.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	t := &task.Task{
		Title:       req.Title,
		Description: req.Description,
	}

	// Используем CreateInColumn из твоего репозитория
	if err := h.tasks.CreateInColumn(r.Context(), t, boardID, columnID, userID); err != nil {
		if errors.Is(err, task.ErrNotFound) {
			// например, колонка или доска не найдена / не принадлежит пользователю
			httputil.Error(w, http.StatusNotFound, "board or column not found")
			return
		}

		log.Printf("failed to create task: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := writeTask(t)
	httputil.JSON(w, http.StatusCreated, resp)
}

// Update обновляет заголовок/описание задачи.
// PUT /api/v1/boards/{board_id}/columns/{column_id}/tasks/{task_id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	taskID := chi.URLParam(r, "task_id")
	if boardID == "" || columnID == "" || taskID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id, column id and task id are required")
		return
	}

	var req createTaskRequest
	if !httputil.DecodeJSONOrError(w, r, &req, httputil.DefaultMaxJSONBodyBytes) {
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.Title == "" {
		httputil.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	t := &task.Task{
		ID:          taskID,
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       req.Title,
		Description: req.Description,
	}

	// У тебя Update(ctx, task *Task) — без ownerID, используем её
	if err := h.tasks.Update(r.Context(), t, userID); err != nil {
		if errors.Is(err, task.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "task not found")
			return
		}
		log.Printf("failed to update task: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := writeTask(t)
	httputil.JSON(w, http.StatusOK, resp)
}

// Delete удаляет задачу из колонки.
// DELETE /api/v1/boards/{board_id}/columns/{column_id}/tasks/{task_id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	taskID := chi.URLParam(r, "task_id")
	if boardID == "" || columnID == "" || taskID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id, column id and task id are required")
		return
	}

	// Используем Delete(ctx, id, boardID, columnID)
	if err := h.tasks.Delete(r.Context(), taskID, boardID, columnID, userID); err != nil {
		if errors.Is(err, task.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "task not found")
			return
		}
		log.Printf("failed to delete task: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Move переносит задачу в другую колонку той же доски.
// PATCH /api/v1/boards/{board_id}/tasks/{task_id}/move
func (h *TaskHandler) Move(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "board_id")
	taskID := chi.URLParam(r, "task_id")
	if boardID == "" || taskID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id and task id are required")
		return
	}

	var req moveTaskRequest
	if !httputil.DecodeJSONOrError(w, r, &req, httputil.DefaultMaxJSONBodyBytes) {
		return
	}
	req.ColumnID = strings.TrimSpace(req.ColumnID)
	if req.ColumnID == "" {
		httputil.Error(w, http.StatusBadRequest, "column_id is required")
		return
	}

	// Минимальная модель задачи: знаем только id и доску.
	t := &task.Task{
		ID:      taskID,
		BoardID: boardID,
	}

	// Используем MoveToColumn(ctx, task *Task, columnID string)
	if err := h.tasks.MoveToColumn(r.Context(), t, req.ColumnID, userID); err != nil {
		if errors.Is(err, task.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "task or column not found")
			return
		}
		log.Printf("failed to move task: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := writeTask(t)
	httputil.JSON(w, http.StatusOK, resp)
}
