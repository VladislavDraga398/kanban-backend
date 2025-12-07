package httputil

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse - единый формат ошибок во всём API
type ErrorResponse struct {
	Error string `json:"error"`
}

// JSON - отдать любой объект как JSON с нужным статусом HTTP
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// Error - отдаём ошибку в виде {"error": "some error"} с заданным статусом HTTP
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, ErrorResponse{Error: msg})
}
