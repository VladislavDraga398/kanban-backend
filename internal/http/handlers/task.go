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
)

// TaskHandler обрабатывает эндпоинты задач.
type TaskHandler struct {
	tasks taskStore
}

// NewTaskHandler создаёт хендлер задач.
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

type createTaskRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    *string  `json:"priority"`
	Labels      []string `json:"labels"`
	DueDate     *string  `json:"due_date"`
}

type moveTaskRequest struct {
	ColumnID string `json:"column_id"`
}

type taskResponse struct {
	ID          string    `json:"id"`
	BoardID     string    `json:"board_id"`
	ColumnID    string    `json:"column_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	Labels      []string  `json:"labels"`
	DueDate     *string   `json:"due_date"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func writeTask(t *task.Task) taskResponse {
	var dueDate *string
	if t.DueDate != nil {
		formatted := t.DueDate.Format("2006-01-02")
		dueDate = &formatted
	}
	labels := t.Labels
	if labels == nil {
		labels = []string{}
	}

	return taskResponse{
		ID:          t.ID,
		BoardID:     t.BoardID,
		ColumnID:    t.ColumnID,
		Title:       t.Title,
		Description: t.Description,
		Priority:    t.Priority,
		Labels:      labels,
		DueDate:     dueDate,
		Position:    t.Position,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func normalizePriority(priority *string, defaultWhenMissing bool) (string, bool, bool) {
	if priority == nil {
		if defaultWhenMissing {
			return "normal", true, true
		}
		return "", false, true
	}

	trimmed := strings.TrimSpace(*priority)
	if trimmed == "" {
		if defaultWhenMissing {
			return "normal", true, true
		}
		return "", false, true
	}

	switch trimmed {
	case "low", "normal", "high":
		return trimmed, true, true
	default:
		return "", false, false
	}
}

func normalizeLabels(labels []string) []string {
	if len(labels) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(labels))
	out := make([]string, 0, len(labels))
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		if len([]rune(label)) > 24 {
			label = string([]rune(label)[:24])
		}
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		out = append(out, label)
		if len(out) == 5 {
			break
		}
	}
	return out
}

func parseDueDate(value *string) (*time.Time, bool, error) {
	if value == nil {
		return nil, false, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, true, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, true, err
	}
	return &parsed, true, nil
}

// List обрабатывает GET /api/v1/boards/{board_id}/columns/{column_id}/tasks.
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	if boardID == "" || columnID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id and column id are required")
		return
	}

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

// Create обрабатывает POST /api/v1/boards/{board_id}/columns/{column_id}/tasks.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
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
	priority, prioritySet, ok := normalizePriority(req.Priority, true)
	if !ok {
		httputil.Error(w, http.StatusBadRequest, "priority must be low, normal or high")
		return
	}
	dueDate, dueDateSet, err := parseDueDate(req.DueDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "due_date must be YYYY-MM-DD")
		return
	}

	if req.Title == "" {
		httputil.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	t := &task.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		PrioritySet: prioritySet,
		Labels:      normalizeLabels(req.Labels),
		LabelsSet:   req.Labels != nil,
		DueDate:     dueDate,
		DueDateSet:  dueDateSet,
	}

	if err := h.tasks.CreateInColumn(r.Context(), t, boardID, columnID, userID); err != nil {
		if errors.Is(err, task.ErrNotFound) {
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

// Update обрабатывает PUT /api/v1/boards/{board_id}/columns/{column_id}/tasks/{task_id}.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
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
	priority, prioritySet, ok := normalizePriority(req.Priority, false)
	if !ok {
		httputil.Error(w, http.StatusBadRequest, "priority must be low, normal or high")
		return
	}
	dueDate, dueDateSet, err := parseDueDate(req.DueDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "due_date must be YYYY-MM-DD")
		return
	}
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
		Priority:    priority,
		PrioritySet: prioritySet,
		Labels:      normalizeLabels(req.Labels),
		LabelsSet:   req.Labels != nil,
		DueDate:     dueDate,
		DueDateSet:  dueDateSet,
	}

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

// Delete обрабатывает DELETE /api/v1/boards/{board_id}/columns/{column_id}/tasks/{task_id}.
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	taskID := chi.URLParam(r, "task_id")
	if boardID == "" || columnID == "" || taskID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id, column id and task id are required")
		return
	}

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

// Move обрабатывает PATCH /api/v1/boards/{board_id}/tasks/{task_id}/move.
func (h *TaskHandler) Move(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
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

	t := &task.Task{
		ID:      taskID,
		BoardID: boardID,
	}

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
