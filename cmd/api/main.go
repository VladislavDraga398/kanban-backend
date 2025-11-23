package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	cfg "github.com/VladislavDraga398/kanban-backend/internal/config"
	myhttp "github.com/VladislavDraga398/kanban-backend/internal/http"
	pg "github.com/VladislavDraga398/kanban-backend/internal/storage/postgres"
)

func main() {
	// 1. Загружаем конфиг (порт + DSN БД)
	config := cfg.Load()

	// 2. Подключаемся к Postgres
	db, err := pg.New(config.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Создаём репозитории поверх БД
	userRepo := pg.NewUserRepository(db)

	// 4. Собираем HTTP-роутер, передавая зависимости
	router := myhttp.NewRouter(myhttp.Deps{
		UserRepo: userRepo,
	})

	// 5. Поднимаем HTTP-сервер
	server := myhttp.NewServer(config.HTTPAddr, router)

	// 6. Ловим сигналы и корректно гасим сервер
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop

	if err := server.Close(); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
