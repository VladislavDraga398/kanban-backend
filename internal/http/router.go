package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter Роутер
func NewRouter() http.Handler {
	router := chi.NewRouter()
	// Пример
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return router
}
