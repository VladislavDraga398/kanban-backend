package postgres

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	DB *sql.DB
}

// New Создание БД
func New(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &DB{db}, nil
}
