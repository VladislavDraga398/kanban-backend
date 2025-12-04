package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/VladislavDraga398/kanban-backend/internal/auth"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
)

type AuthHandler struct {
	users user.Repository
}

func NewAuthHandler(users user.Repository) *AuthHandler {
	return &AuthHandler{users: users}
}

// loginRequest - запрос на авторизацию
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse - ответ на запрос на авторизацию
type loginResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Register - регистрация пользователя
type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// registerResponse - ответ на регистрацию пользователя
type registerResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	u := &user.User{
		Email:        req.Email,
		PasswordHash: hash,
	}

	if err := h.users.Create(r.Context(), u); err != nil {
		// бизнес-ошибка: email уже занят
		if errors.Is(err, user.ErrEmailAlreadyUsed) {
			http.Error(w, "email already in use", http.StatusConflict) // 409
			return
		}

		// всё остальное — системные ошибки
		log.Printf("failed to create user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := registerResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	// Идет поиск пользователя по email
	u, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		// Если пользователь не найден — возвращаем ошибку 401
		if errors.Is(err, user.ErrNotFound) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		// Любые другие ошибки — системные
		log.Printf("failed to get user: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Проверка пароля
	if err := auth.ComparePasswords(u.PasswordHash, req.Password); err != nil {
		// bcrypt.ErrMismatchedHashAndPassword — пароль не совпадает
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	resp := loginResponse{
		ID:    u.ID,
		Email: u.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
