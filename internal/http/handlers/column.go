package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/column"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
)

type ColumnHandler struct {
	columns column.Repository
}

func NewColumnHandler(columns column.Repository) *ColumnHandler {
	return &ColumnHandler{columns: columns}
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

// List — GET /api/v1/boards/{board_id}/columns
func (h *ColumnHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	boardID := chi.URLParam(r, "board_id")
	if boardID == "" {
		http.Error(w, "board id is required", http.StatusBadRequest)
		return
	}

	cols, err := h.columns.ListByBoardOwner(r.Context(), boardID, userID)
	if err != nil {
		log.Printf("failed to list columns: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]columnResponse, 0, len(cols))
	for _, c := range cols {
		resp = append(resp, writeColumn(c))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Create — POST /api/v1/boards/{board_id}/columns
func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	boardID := chi.URLParam(r, "board_id")
	if boardID == "" {
		http.Error(w, "board id is required", http.StatusBadRequest)
		return
	}

	var req createColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	c := &column.Column{
		Name: req.Name,
	}

	if err := h.columns.CreateInBoard(r.Context(), c, boardID, userID); err != nil {
		log.Printf("failed to create column: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := writeColumn(c)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Update — PUT /api/v1/boards/{board_id}/columns/{column_id}
func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	broardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	if broardID == "" || columnID == "" {
		http.Error(w, "board id is required", http.StatusBadRequest)
		return
	}

	var req createColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	c := &column.Column{
		ID:      columnID,
		BoardID: broardID,
		Name:    req.Name,
	}

	if err := h.columns.Update(r.Context(), c); err != nil {
		if errors.Is(err, column.ErrNotFound) {
			http.Error(w, "column not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to update column: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := writeColumn(c)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Delete — DELETE /api/v1/boards/{board_id}/columns/{column_id}
func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	boardID := chi.URLParam(r, "board_id")
	columnID := chi.URLParam(r, "column_id")
	if boardID == "" || columnID == "" {
		http.Error(w, "board id is required", http.StatusBadRequest)
		return
	}

	if err := h.columns.Delete(r.Context(), columnID, boardID); err != nil {
		if errors.Is(err, column.ErrNotFound) {
			http.Error(w, "column not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to delete column: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
