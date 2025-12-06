package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/VladislavDraga398/kanban-backend/internal/auth"
	"github.com/VladislavDraga398/kanban-backend/internal/http/httputil"
)

type ctxKey string

const userIDKey ctxKey = "userID"

// Auth валидирует JWT из Authorization: Bearer <token> и кладёт userID в контекст.
func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				// Подсказка клиенту, что требуется Bearer токен
				w.Header().Set("WWW-Authenticate", "Bearer")
				httputil.Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := auth.ParseJWT(token, secret)
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Bearer error=\"invalid_token\"")
				httputil.Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext достает userID, который положил Auth.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	id, ok := v.(string)
	return id, ok && id != ""
}
