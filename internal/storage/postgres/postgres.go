package postgres

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB - обертка над *sql.DB, для навески модов/репозитории.
type DB struct {
	*sql.DB
}

// New открывает подключение к Postgres и проверяет его.
func New(dsn string) (*DB, error) {
	// sql.Open - не подключает сразу, для настройки драйвера.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// База настройки пула подключений.
	db.SetMaxOpenConns(10)               // Максимальное количество соединений в пуле.
	db.SetMaxIdleConns(5)                // Максимальное количество пустых соединений в пуле.
	db.SetConnMaxLifetime(1 * time.Hour) // Максимальное время жизни соединения в пуле.

	// Ping - проверяет подключение к БД.
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &DB{db}, nil
}
