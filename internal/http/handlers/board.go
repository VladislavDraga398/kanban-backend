package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/board"
	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
)

// BoardHandler отвечает за операции с досками (boards).
type BoardHandler struct {
	boards board.Repository
}

// NewBoardHandler конструирует хэндлер досок.
func NewBoardHandler(boards board.Repository) *BoardHandler {
	return &BoardHandler{boards: boards}
}

// createBoardRequest — тело запроса при создании/обновлении доски.
type createBoardRequest struct {
	Name string `json:"name"`
}

// boardResponse — то, что мы отдаём наружу клиенту.
type boardResponse struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// writeBoard — маппинг доменной модели в DTO для ответа.
func writeBoard(b *board.Board) boardResponse {
	return boardResponse{
		ID:        b.ID,
		OwnerID:   b.OwnerID,
		Name:      b.Name,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

// List возвращает все доски пользователя. GET /api/v1/boards
func (h *BoardHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardsList, err := h.boards.ListByOwnerID(r.Context(), userID)
	if err != nil {
		log.Printf("failed to list boards: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := make([]boardResponse, 0, len(boardsList))
	for _, b := range boardsList {
		resp = append(resp, writeBoard(b))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// Get возвращает доску по её ID. GET /api/v1/boards/{id}
func (h *BoardHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	b, err := h.boards.GetByID(r.Context(), boardID, userID)
	if err != nil {
		if errors.Is(err, board.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "board not found")
			return
		}
		log.Printf("failed to get board: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := writeBoard(b)
	httputil.JSON(w, http.StatusOK, resp)
}

// Create создаёт новую доску. POST /api/v1/boards
func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	b := &board.Board{
		OwnerID: userID,
		Name:    req.Name,
	}

	if err := h.boards.Create(r.Context(), b); err != nil {
		log.Printf("failed to create board: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := writeBoard(b)
	httputil.JSON(w, http.StatusCreated, resp)
}

// Update обновляет название доски. PUT /api/v1/boards/{id}
func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req createBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}

	// Собираем доменную модель для репозитория
	b := &board.Board{
		ID:      boardID,
		OwnerID: userID,
		Name:    req.Name,
	}

	if err := h.boards.Update(r.Context(), b); err != nil {
		if errors.Is(err, board.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "board not found")
			return
		}
		log.Printf("failed to update board: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := writeBoard(b)
	httputil.JSON(w, http.StatusOK, resp)
}

// Delete удаляет доску. DELETE /api/v1/boards/{id}
func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	if err := h.boards.Delete(r.Context(), boardID, userID); err != nil {
		if errors.Is(err, board.ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "board not found")
			return
		}
		log.Printf("failed to delete board: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Для DELETE обычно 204 No Content без тела
	w.WriteHeader(http.StatusNoContent)
}
