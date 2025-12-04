package http

import (
	"log"
	"net/http"
	"time"
)

// Server Сервер структура
type Server struct {
	httpServer *http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

// Start Запуск сервера
func (s *Server) Start() error {
	log.Printf("srarting HTTP server on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Close Закрытие сервера
func (s *Server) Close() error {
	log.Println("shutnging down HTTP server...")
	return s.httpServer.Close()
}
