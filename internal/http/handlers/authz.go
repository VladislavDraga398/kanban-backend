package handlers

import (
	"net/http"

	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
	"github.com/VladislavDraga398/kanban-backend/internal/http/middleware"
)

// requireUserID извлекает userID из контекста запроса.
// При отсутствии авторизации пишет 401 и возвращает false.
func requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return "", false
	}
	return userID, true
}
