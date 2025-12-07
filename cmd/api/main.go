package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	boardRepo, columnRepo, taskRepo := pg.NewBoardRepository(db), pg.NewColumnRepository(db), pg.NewTaskRepository(db)

	// 4. Собираем HTTP-роутер, передавая зависимости
	router := myhttp.NewRouter(myhttp.Deps{
		UserRepo:   userRepo,
		BoardRepo:  boardRepo,
		ColumnRepo: columnRepo,
		TaskRepo:   taskRepo,
		JWTSecret:  config.JWTSecret,
		JWTTTL:     config.JWTTTL,
	})

	// 5. Поднимаем HTTP-сервер
	server := myhttp.NewServer(config.HTTPAddr, router)

	// 6. Ловим сигналы и корректно гасим сервер
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("failed to start http server: %v", err)
		}
	}()

	<-stop

	// Плавное завершение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("failed to gracefully shutdown http server: %v", err)
	}
}
