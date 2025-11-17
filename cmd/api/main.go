package main

import (
	"os"
	"os/signal"
	"syscall"

	myhttp "github.com/VladislavDraga398/kanban-backend/internal/http"
)

func main() {

	addr, router := ":8038", myhttp.NewRouter() // Потом внести в конфиг
	server := myhttp.NewServer(addr, router)    // Ловим сигналы

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Стартуем сервер
	go func() {
		if err := server.Start(); err != nil {
			panic(err)
		}
	}()

	<-stop
	if err := server.Close(); err != nil {
		panic(err)
	}
}
