package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
)

type TaskHandler struct {
	tasks task.Repository
}

func NewTaskHandler(tasks task.Repository) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

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

// List — GET /api/v1/boards/{board_id}/columns/{column_id}/tasks
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	if boardID == "" || columnID == "" {
		http.Error(w, "board id and column id are required", http.StatusBadRequest)
		return
	}

	tasks, err := h.tasks.ListByColumnOwner(r.Context(), boardID, columnID, userID)
	if err != nil {
		log.Printf("failed to list tasks: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, writeTask(t))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Create — POST /api/v1/boards/{board_id}/columns/{column_id}/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	if boardID == "" || columnID == "" {
		http.Error(w, "board id and column id are required", http.StatusBadRequest)
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	t := &task.Task{
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.tasks.CreateInColumn(r.Context(), t, boardID, columnID, userID); err != nil {
		log.Printf("failed to create task: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := writeTask(t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Update — PUT /api/v1/boards/{board_id}/columns/{column_id}/tasks/{task_id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	taskID := chi.URLParam(r, "task_id")
	if boardID == "" || columnID == "" || taskID == "" {
		http.Error(w, "board id, column id and task id are required", http.StatusBadRequest)
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	t := &task.Task{
		ID:          taskID,
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.tasks.Update(r.Context(), t); err != nil {
		if errors.Is(err, task.ErrNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to update task: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := writeTask(t)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Delete — DELETE /api/v1/boards/{board_id}/columns/{column_id}/tasks/{task_id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	taskID := chi.URLParam(r, "task_id")
	if boardID == "" || columnID == "" || taskID == "" {
		http.Error(w, "board id, column id and task id are required", http.StatusBadRequest)
		return
	}

	// Удаляем задачу, валидируя принадлежность к доске и колонке из URL.
	if err := h.tasks.Delete(r.Context(), taskID, boardID, columnID); err != nil {
		if errors.Is(err, task.ErrNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to delete task: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
