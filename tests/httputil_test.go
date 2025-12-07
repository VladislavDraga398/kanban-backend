package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
)

func TestJSONWritesBodyAndStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	type payload struct {
		Name string `json:"name"`
	}
	httputil.JSON(rr, http.StatusCreated, payload{Name: "ok"})

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
	if rr.Body.String() == "" {
		t.Fatalf("expected body")
	}
}

func TestErrorUsesJSONEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	httputil.Error(rr, http.StatusBadRequest, "bad request")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Body.String() == "" {
		t.Fatalf("expected error body")
	}
}
