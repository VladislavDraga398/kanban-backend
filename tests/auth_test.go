package tests

import (
	"testing"
	"time"

	"github.com/VladislavDraga398/kanban-backend/internal/auth"
)

func TestHashAndComparePassword(t *testing.T) {
	hash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if err := auth.ComparePasswords(hash, "secret"); err != nil {
		t.Fatalf("compare should succeed: %v", err)
	}
	if err := auth.ComparePasswords(hash, "wrong"); err == nil {
		t.Fatalf("expected compare to fail for wrong password")
	}
}

func TestJWTGenerateAndParse(t *testing.T) {
	secret := []byte("jwt-secret")
	token, err := auth.GenerateJWT("user-1", secret, time.Minute)
	if err != nil {
		t.Fatalf("generate jwt: %v", err)
	}

	userID, err := auth.ParseJWT(token, secret)
	if err != nil {
		t.Fatalf("parse jwt: %v", err)
	}
	if userID != "user-1" {
		t.Fatalf("unexpected user id: %s", userID)
	}
}

func TestJWTInvalidOnExpiry(t *testing.T) {
	secret := []byte("jwt-secret")
	token, err := auth.GenerateJWT("user-1", secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("generate jwt: %v", err)
	}

	if _, err := auth.ParseJWT(token, secret); err == nil {
		t.Fatalf("expected expired token to be invalid")
	}
}
