package config

import "os"

type Config struct {
	HTTPAddr string
	DBDSN    string
}

func Load() *Config {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = ":8038"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// дефол плд docker-compose (kanban/kanban@localhost:5432/kanban)
		dsn = "postgres://postgres:postgres@localhost:5432/kanban?sslmode=disable"
	}

	return &Config{HTTPAddr: port, DBDSN: dsn}
}
