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
	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
)

type AuthHandler struct {
	users     user.Repository
	jwtSecret []byte
	jwtTTL    time.Duration
}

func NewAuthHandler(users user.Repository, jwtSecret string, jwtTTL time.Duration) *AuthHandler {
	return &AuthHandler{users: users, jwtSecret: []byte(jwtSecret), jwtTTL: jwtTTL}
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
	Token string `json:"token"`
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
	Token     string    `json:"token"`
}

// Register обрабатывает POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		// Невалидный JSON в теле запроса
		httputil.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		// Бизнес-ошибка валидации: не хватает данных
		httputil.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		// Системная ошибка при хешировании
		log.Printf("failed to hash password: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	u := &user.User{
		Email:        req.Email,
		PasswordHash: hash,
	}

	if err := h.users.Create(r.Context(), u); err != nil {
		// бизнес-ошибка: email уже занят
		if errors.Is(err, user.ErrEmailAlreadyUsed) {
			// 409 Conflict — логично для "email уже используется"
			httputil.Error(w, http.StatusConflict, "email already in use")
			return
		}

		// всё остальное — системные ошибки
		log.Printf("failed to create user: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	token, err := auth.GenerateJWT(u.ID, h.jwtSecret, h.jwtTTL)
	if err != nil {
		log.Printf("failed to sign token: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := registerResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		Token:     token,
	}

	// Успех — возвращаем 201 + JSON
	httputil.JSON(w, http.StatusCreated, resp)
}

// Login обрабатывает POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Идём в репозиторий — ищем пользователя по email
	u, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		// Если пользователь не найден — 401, но не раскрываем, что именно не так
		if errors.Is(err, user.ErrNotFound) {
			httputil.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		// Любые другие ошибки — системные
		log.Printf("failed to get user: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Проверка пароля
	if err := auth.ComparePasswords(u.PasswordHash, req.Password); err != nil {
		// Неверный пароль — тоже 401, тот же текст, чтобы не палить, существует ли email
		httputil.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateJWT(u.ID, h.jwtSecret, h.jwtTTL)
	if err != nil {
		log.Printf("failed to sign token: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := loginResponse{
		ID:    u.ID,
		Email: u.Email,
		Token: token,
	}

	// Успешный логин — 200 + JSON с данными пользователя (пока без токена)
	httputil.JSON(w, http.StatusOK, resp)
}
