package tests

import (
	"testing"
	"time"

	"github.com/VladislavDraga398/kanban-backend/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_TTL", "")

	cfg := config.Load()
	if cfg.HTTPAddr != ":8083" {
		t.Fatalf("expected default port, got %s", cfg.HTTPAddr)
	}
	if cfg.DBDSN == "" {
		t.Fatalf("expected default DSN")
	}
	if cfg.JWTSecret != "secret" {
		t.Fatalf("expected jwt secret from env, got %s", cfg.JWTSecret)
	}
	if cfg.JWTTTL != 24*time.Hour {
		t.Fatalf("expected default ttl 24h, got %s", cfg.JWTTTL)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("HTTP_PORT", "9999")
	t.Setenv("DB_DSN", "postgres://u:p@host/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_TTL", "2h")

	cfg := config.Load()
	if cfg.HTTPAddr != ":9999" {
		t.Fatalf("expected :9999, got %s", cfg.HTTPAddr)
	}
	if cfg.DBDSN != "postgres://u:p@host/db" {
		t.Fatalf("dsn override not applied")
	}
	if cfg.JWTSecret != "secret" {
		t.Fatalf("jwt secret override not applied")
	}
	if cfg.JWTTTL != 2*time.Hour {
		t.Fatalf("jwt ttl override not applied: %s", cfg.JWTTTL)
	}
}
