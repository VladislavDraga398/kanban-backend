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
	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
)

// ColumnHandler — хендлер для работы с колонками на досках.
type ColumnHandler struct {
	columns column.Repository
}

// NewColumnHandler конструирует хендлер колонок.
func NewColumnHandler(columns column.Repository) *ColumnHandler {
	return &ColumnHandler{columns: columns}
}

// createColumnRequest — тело запроса на создание/обновление колонки.
type createColumnRequest struct {
	Name string `json:"name"`
}

// columnResponse — то, что отдаём наружу клиенту.
type columnResponse struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// writeColumn — маппинг доменной модели в DTO.
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

// List возвращает все колонки доски текущего пользователя.
// GET /api/v1/boards/{board_id}/columns
func (h *ColumnHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "board_id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	cols, err := h.columns.ListByBoardOwner(r.Context(), boardID, userID)
	if err != nil {
		// В большинстве случаев List* возвращает пустой список и nil,
		// но если репозиторий вернёт ошибку — это уже системная проблема.
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

// Create создаёт новую колонку в доске пользователя.
// POST /api/v1/boards/{board_id}/columns
func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "board_id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req createColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid json")
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

	// CreateInBoard проверяет, что доска принадлежит пользователю,
	// выставляет position и timestamps.
	if err := h.columns.CreateInBoard(r.Context(), c, boardID, userID); err != nil {
		if errors.Is(err, column.ErrNotFound) {
			// Например, доска не найдена или не принадлежит пользователю.
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

// Update обновляет название (и при необходимости позицию) колонки.
// PUT /api/v1/boards/{board_id}/columns/{column_id}
func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	// Здесь предполагается, что репозиторий проверяет board_id/owner_id,
	// либо ты проверяешь доступ выше по уровню.
	c := &column.Column{
		ID:      columnID,
		BoardID: boardID,
		Name:    req.Name,
	}

	if err := h.columns.Update(r.Context(), c); err != nil {
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

// Delete удаляет колонку с доски.
// DELETE /api/v1/boards/{board_id}/columns/{column_id}
func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
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

	if err := h.columns.Delete(r.Context(), columnID, boardID); err != nil {
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
