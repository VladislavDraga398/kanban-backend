package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/board"
	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
)

// BoardHandler обрабатывает эндпоинты досок.
type BoardHandler struct {
	boards boardStore
}

// NewBoardHandler создаёт хендлер досок.
func NewBoardHandler(boards boardStore) *BoardHandler {
	return &BoardHandler{boards: boards}
}

type boardStore interface {
	ListByOwnerID(ctx context.Context, ownerID string) ([]*board.Board, error)
	GetByID(ctx context.Context, id, ownerID string) (*board.Board, error)
	Create(ctx context.Context, b *board.Board) error
	Update(ctx context.Context, b *board.Board) error
	Delete(ctx context.Context, id, ownerID string) error
}

type createBoardRequest struct {
	Name string `json:"name"`
}

type boardResponse struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func writeBoard(b *board.Board) boardResponse {
	return boardResponse{
		ID:        b.ID,
		OwnerID:   b.OwnerID,
		Name:      b.Name,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

// List обрабатывает GET /api/v1/boards.
func (h *BoardHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
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

// Get обрабатывает GET /api/v1/boards/{id}.
func (h *BoardHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
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

// Create обрабатывает POST /api/v1/boards.
func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req createBoardRequest
	if !httputil.DecodeJSONOrError(w, r, &req, httputil.DefaultMaxJSONBodyBytes) {
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

// Update обрабатывает PUT /api/v1/boards/{id}.
func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	boardID := chi.URLParam(r, "id")
	if boardID == "" {
		httputil.Error(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req createBoardRequest
	if !httputil.DecodeJSONOrError(w, r, &req, httputil.DefaultMaxJSONBodyBytes) {
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name is required")
		return
	}

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

// Delete обрабатывает DELETE /api/v1/boards/{id}.
func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
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

	w.WriteHeader(http.StatusNoContent)
}
