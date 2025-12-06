package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VladislavDraga398/kanban-backend/internal/auth"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
	myhttp "github.com/VladislavDraga398/kanban-backend/internal/http"
)

func TestProtectedRouteRequiresJWT(t *testing.T) {
	secret := "secret"
	router := myhttp.NewRouter(myhttp.Deps{
		UserRepo:   &stubUserRepo{},
		BoardRepo:  &stubBoardRepo{},
		ColumnRepo: &stubColumnRepo{},
		TaskRepo:   &stubTaskRepo{moveFn: func(ctx context.Context, t *task.Task, columnID string) error { return nil }},
		JWTSecret:  secret,
		JWTTTL:     time.Hour,
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/boards", nil)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rr.Code)
	}

	token, err := auth.GenerateJWT("user-1", []byte(secret), time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/boards", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with token, got %d", rr.Code)
	}
}
