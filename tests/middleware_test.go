package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VladislavDraga398/kanban-backend/internal/auth"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
)

func TestAuthMiddlewareAllowsValidToken(t *testing.T) {
	secret := []byte("secret")
	token, err := auth.GenerateJWT("user-123", secret, time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	handler := middleware.Auth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := middleware.UserIDFromContext(r.Context()); !ok || id != "user-123" {
			t.Fatalf("unexpected user id: %q ok=%v", id, ok)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	secret := []byte("secret")

	handler := middleware.Auth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
