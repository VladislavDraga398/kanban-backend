package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/VladislavDraga398/kanban-backend/internal/auth"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/board"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/column"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
	myhttp "github.com/VladislavDraga398/kanban-backend/internal/http"
	"github.com/VladislavDraga398/kanban-backend/internal/http/handlers"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
)

// --- stubs shared across tests ---

type stubUserRepo struct {
	createFn    func(ctx context.Context, u *user.User) error
	getByIDFn   func(ctx context.Context, id string) (*user.User, error)
	getByEmailF func(ctx context.Context, email string) (*user.User, error)
}

func (s *stubUserRepo) Create(ctx context.Context, u *user.User) error {
	if s.createFn != nil {
		return s.createFn(ctx, u)
	}
	return nil
}

func (s *stubUserRepo) GetByID(ctx context.Context, id string) (*user.User, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return nil, user.ErrNotFound
}

func (s *stubUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if s.getByEmailF != nil {
		return s.getByEmailF(ctx, email)
	}
	return nil, user.ErrNotFound
}

type stubBoardRepo struct {
	createFn func(ctx context.Context, b *board.Board) error
	updateFn func(ctx context.Context, b *board.Board) error
	getFn    func(ctx context.Context, id, ownerID string) (*board.Board, error)
	listFn   func(ctx context.Context, ownerID string) ([]*board.Board, error)
	deleteFn func(ctx context.Context, id, ownerID string) error
}

func (s *stubBoardRepo) Create(ctx context.Context, b *board.Board) error {
	if s.createFn != nil {
		return s.createFn(ctx, b)
	}
	return nil
}

func (s *stubBoardRepo) Update(ctx context.Context, b *board.Board) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, b)
	}
	return nil
}

func (s *stubBoardRepo) GetByID(ctx context.Context, id, ownerID string) (*board.Board, error) {
	if s.getFn != nil {
		return s.getFn(ctx, id, ownerID)
	}
	return nil, board.ErrNotFound
}

func (s *stubBoardRepo) ListByOwnerID(ctx context.Context, ownerID string) ([]*board.Board, error) {
	if s.listFn != nil {
		return s.listFn(ctx, ownerID)
	}
	return nil, nil
}

func (s *stubBoardRepo) Delete(ctx context.Context, id, ownerID string) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id, ownerID)
	}
	return nil
}

type stubColumnRepo struct {
	createFn      func(ctx context.Context, c *column.Column) error
	listFn        func(ctx context.Context, boardID string) ([]column.Column, error)
	updateFn      func(ctx context.Context, c *column.Column) error
	deleteFn      func(ctx context.Context, id, boardID string) error
	listByOwnerFn func(ctx context.Context, boardID, ownerID string) ([]*column.Column, error)
	createInFn    func(ctx context.Context, c *column.Column, boardID, ownerID string) error
}

func (s *stubColumnRepo) Create(ctx context.Context, c *column.Column) error {
	if s.createFn != nil {
		return s.createFn(ctx, c)
	}
	return nil
}

func (s *stubColumnRepo) ListByBoardID(ctx context.Context, boardID string) ([]column.Column, error) {
	if s.listFn != nil {
		return s.listFn(ctx, boardID)
	}
	return nil, nil
}

func (s *stubColumnRepo) Update(ctx context.Context, c *column.Column) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, c)
	}
	return nil
}

func (s *stubColumnRepo) Delete(ctx context.Context, id, boardID string) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id, boardID)
	}
	return nil
}

func (s *stubColumnRepo) ListByBoardOwner(ctx context.Context, boardID, ownerID string) ([]*column.Column, error) {
	if s.listByOwnerFn != nil {
		return s.listByOwnerFn(ctx, boardID, ownerID)
	}
	return nil, nil
}

func (s *stubColumnRepo) CreateInBoard(ctx context.Context, c *column.Column, boardID, ownerID string) error {
	if s.createInFn != nil {
		return s.createInFn(ctx, c, boardID, ownerID)
	}
	return nil
}

type stubTaskRepo struct {
	moveFn              func(ctx context.Context, t *task.Task, columnID string) error
	listByColumnOwnerFn func(ctx context.Context, boardID, columnID, ownerID string) ([]*task.Task, error)
	createInColumnFn    func(ctx context.Context, t *task.Task, boardID, columnID, ownerID string) error
	updateFn            func(ctx context.Context, t *task.Task) error
	deleteFn            func(ctx context.Context, id, boardID, columnID string) error
}

func (s *stubTaskRepo) Create(ctx context.Context, t *task.Task) error { return nil }
func (s *stubTaskRepo) MoveToColumn(ctx context.Context, t *task.Task, columnID string) error {
	return s.moveFn(ctx, t, columnID)
}
func (s *stubTaskRepo) ListByBoard(ctx context.Context, boardID string) ([]task.Task, error) {
	return nil, nil
}
func (s *stubTaskRepo) ListByColumn(ctx context.Context, columnID string) ([]task.Task, error) {
	return nil, nil
}
func (s *stubTaskRepo) ListByColumnOwner(ctx context.Context, boardID, columnID, ownerID string) ([]*task.Task, error) {
	if s.listByColumnOwnerFn != nil {
		return s.listByColumnOwnerFn(ctx, boardID, columnID, ownerID)
	}
	return nil, nil
}
func (s *stubTaskRepo) CreateInColumn(ctx context.Context, t *task.Task, boardID, columnID, ownerID string) error {
	if s.createInColumnFn != nil {
		return s.createInColumnFn(ctx, t, boardID, columnID, ownerID)
	}
	return nil
}
func (s *stubTaskRepo) Update(ctx context.Context, t *task.Task) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, t)
	}
	return nil
}
func (s *stubTaskRepo) Delete(ctx context.Context, id, boardID, columnID string) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id, boardID, columnID)
	}
	return nil
}

// --- helpers ---

func doJSONRequest(r http.Handler, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func bearer(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
}

const testSecret = "test-secret"

func mustToken(t *testing.T, userID string) string {
	t.Helper()
	token, err := auth.GenerateJWT(userID, []byte(testSecret), time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

// --- tests ---

func TestAuthRegisterSuccess(t *testing.T) {
	h := handlers.NewAuthHandler(&stubUserRepo{
		createFn: func(ctx context.Context, u *user.User) error {
			u.ID = "user-1"
			u.CreatedAt = time.Unix(1, 0)
			return nil
		},
	}, testSecret, time.Hour)

	r := chi.NewRouter()
	r.Post("/api/v1/auth/register", h.Register)

	reqBody := map[string]string{"email": "a@b.c", "password": "pass123"}
	rec := doJSONRequest(r, http.MethodPost, "/api/v1/auth/register", reqBody, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var resp struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != "user-1" || resp.Email != "a@b.c" || resp.Token == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAuthLoginInvalidPassword(t *testing.T) {
	hash, _ := auth.HashPassword("correct-pass")
	h := handlers.NewAuthHandler(&stubUserRepo{
		getByEmailF: func(ctx context.Context, email string) (*user.User, error) {
			return &user.User{ID: "u1", Email: email, PasswordHash: hash}, nil
		},
	}, testSecret, time.Hour)

	r := chi.NewRouter()
	r.Post("/api/v1/auth/login", h.Login)

	reqBody := map[string]string{"email": "a@b.c", "password": "wrong"}
	rec := doJSONRequest(r, http.MethodPost, "/api/v1/auth/login", reqBody, nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBoardListUnauthorized(t *testing.T) {
	h := handlers.NewBoardHandler(&stubBoardRepo{})
	r := chi.NewRouter()
	r.Use(middleware.Auth([]byte(testSecret)))
	r.Get("/api/v1/boards", h.List)

	rec := doJSONRequest(r, http.MethodGet, "/api/v1/boards", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBoardCreateSuccess(t *testing.T) {
	h := handlers.NewBoardHandler(&stubBoardRepo{
		createFn: func(ctx context.Context, b *board.Board) error {
			b.ID = "board-1"
			b.OwnerID = "owner-1"
			b.CreatedAt = time.Unix(1, 0)
			b.UpdatedAt = time.Unix(1, 0)
			return nil
		},
	})

	r := chi.NewRouter()
	r.Use(middleware.Auth([]byte(testSecret)))
	r.Post("/api/v1/boards", h.Create)

	token := mustToken(t, "owner-1")
	rec := doJSONRequest(r, http.MethodPost, "/api/v1/boards", map[string]string{"name": "My board"}, bearer(token))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var resp struct {
		ID      string `json:"id"`
		OwnerID string `json:"owner_id"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != "board-1" || resp.OwnerID != "owner-1" || resp.Name != "My board" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestColumnCreateNotFound(t *testing.T) {
	h := handlers.NewColumnHandler(&stubColumnRepo{
		createInFn: func(ctx context.Context, c *column.Column, boardID, ownerID string) error {
			return column.ErrNotFound
		},
	})

	r := chi.NewRouter()
	r.Use(middleware.Auth([]byte(testSecret)))
	r.Post("/api/v1/boards/{board_id}/columns", h.Create)

	token := mustToken(t, "owner-1")
	rec := doJSONRequest(r, http.MethodPost, "/api/v1/boards/board-1/columns", map[string]string{"name": "Todo"}, bearer(token))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTaskCreateValidation(t *testing.T) {
	h := handlers.NewTaskHandler(&stubTaskRepo{})

	r := chi.NewRouter()
	r.Use(middleware.Auth([]byte(testSecret)))
	r.Post("/api/v1/boards/{board_id}/columns/{column_id}/tasks", h.Create)

	token := mustToken(t, "owner-1")
	rec := doJSONRequest(r, http.MethodPost, "/api/v1/boards/b1/columns/c1/tasks", map[string]string{"title": "   "}, bearer(token))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTaskMoveSuccessThroughRouter(t *testing.T) {
	taskRepo := &stubTaskRepo{
		moveFn: func(ctx context.Context, t *task.Task, columnID string) error {
			t.ColumnID = columnID
			t.Position = 3
			t.Title = "moved"
			t.Description = "updated"
			t.CreatedAt = time.Unix(1, 0)
			t.UpdatedAt = time.Unix(2, 0)
			return nil
		},
	}

	router := myhttp.NewRouter(myhttp.Deps{
		UserRepo:   &stubUserRepo{},
		BoardRepo:  &stubBoardRepo{},
		ColumnRepo: &stubColumnRepo{},
		TaskRepo:   taskRepo,
		JWTSecret:  testSecret,
		JWTTTL:     time.Hour,
	})

	headers := bearer(mustToken(t, "owner-1"))
	rec := doJSONRequest(router, http.MethodPatch, "/api/v1/boards/b1/tasks/t1/move", map[string]string{"column_id": "col-2"}, headers)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		ColumnID string `json:"column_id"`
		Position int    `json:"position"`
		Title    string `json:"title"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ColumnID != "col-2" || resp.Position != 3 || resp.Title != "moved" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestTaskMoveNotFound(t *testing.T) {
	taskRepo := &stubTaskRepo{
		moveFn: func(ctx context.Context, t *task.Task, columnID string) error {
			return task.ErrNotFound
		},
	}

	router := myhttp.NewRouter(myhttp.Deps{
		UserRepo:   &stubUserRepo{},
		BoardRepo:  &stubBoardRepo{},
		ColumnRepo: &stubColumnRepo{},
		TaskRepo:   taskRepo,
		JWTSecret:  testSecret,
		JWTTTL:     time.Hour,
	})

	headers := bearer(mustToken(t, "owner-1"))
	rec := doJSONRequest(router, http.MethodPatch, "/api/v1/boards/b1/tasks/t1/move", map[string]string{"column_id": "col-x"}, headers)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "not found") {
		t.Fatalf("expected not found message, got %q", rec.Body.String())
	}
}
