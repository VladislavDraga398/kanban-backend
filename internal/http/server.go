package http

import (
	"context"
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
	log.Printf("starting HTTP server on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown Плавное завершение работы сервера с таймаутом из внешнего контекста
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
