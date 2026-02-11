package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/column"
	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
)

// ColumnHandler обрабатывает эндпоинты колонок.
type ColumnHandler struct {
	columns columnStore
}

// NewColumnHandler создаёт хендлер колонок.
func NewColumnHandler(columns columnStore) *ColumnHandler {
	return &ColumnHandler{columns: columns}
}

type columnStore interface {
	ListByBoardOwner(ctx context.Context, boardID, ownerID string) ([]*column.Column, error)
	CreateInBoard(ctx context.Context, column *column.Column, boardID, ownerID string) error
	Update(ctx context.Context, c *column.Column, ownerID string) error
	Delete(ctx context.Context, id, boardID, ownerID string) error
}

type createColumnRequest struct {
	Name string `json:"name"`
}

type columnResponse struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func writeColumn(c *column.Column) columnResponse {
	return columnResponse{
		ID:        c.ID,
		BoardID:   c.BoardID,
		Name:      c.Name,
		Position:  c.Position,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// List обрабатывает GET /api/v1/boards/{board_id}/columns.
func (h *ColumnHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	boardID := chi.URLParam(r, "board_id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	cols, err := h.columns.ListByBoardOwner(r.Context(), boardID, userID)
	if err != nil {
		log.Printf("failed to list columns: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := make([]columnResponse, 0, len(cols))
	for _, c := range cols {
		resp = append(resp, writeColumn(c))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// Create обрабатывает POST /api/v1/boards/{board_id}/columns.
func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	boardID := chi.URLParam(r, "board_id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req createColumnRequest
	if !httputil.DecodeJSONOrError(w, r, &req, httputil.DefaultMaxJSONBodyBytes) {
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	c := &column.Column{
		Name: req.Name,
	}

	if err := h.columns.CreateInBoard(r.Context(), c, boardID, userID); err != nil {
		if errors.Is(err, column.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "board not found")
			return
		}

		log.Printf("failed to create column: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := writeColumn(c)
	httputil.JSON(w, http.StatusCreated, resp)
}

// Update обрабатывает PUT /api/v1/boards/{board_id}/columns/{column_id}.
func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	boardID := chi.URLParam(r, "board_id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	columnID := chi.URLParam(r, "column_id")
	if columnID == "" {
		httputil.Error(w, http.StatusBadRequest, "column id is required")
		return
	}

	var req createColumnRequest
	if !httputil.DecodeJSONOrError(w, r, &req, httputil.DefaultMaxJSONBodyBytes) {
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	c := &column.Column{
		ID:      columnID,
		BoardID: boardID,
		Name:    req.Name,
	}

	if err := h.columns.Update(r.Context(), c, userID); err != nil {
		if errors.Is(err, column.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "column not found")
			return
		}
		log.Printf("failed to update column: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := writeColumn(c)
	httputil.JSON(w, http.StatusOK, resp)
}

// Delete обрабатывает DELETE /api/v1/boards/{board_id}/columns/{column_id}.
func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	boardID := chi.URLParam(r, "board_id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	columnID := chi.URLParam(r, "column_id")
	if columnID == "" {
		httputil.Error(w, http.StatusBadRequest, "column id is required")
		return
	}

	if err := h.columns.Delete(r.Context(), columnID, boardID, userID); err != nil {
		if errors.Is(err, column.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "column not found")
			return
		}
		log.Printf("failed to delete column: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
