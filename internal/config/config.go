package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr  string
	DBDSN     string
	JWTSecret string
	JWTTTL    time.Duration
}

func Load() *Config {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		// тут ставим твой дефолтный порт
		port = "8083"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://kanban:kanban@localhost:5432/kanban?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is required")
	}

	ttl := 24 * time.Hour
	if ttlStr := os.Getenv("JWT_TTL"); ttlStr != "" {
		if parsed, err := time.ParseDuration(ttlStr); err == nil {
			ttl = parsed
		}
	}

	return &Config{
		HTTPAddr:  ":" + port, // вот тут формируется ":8083"
		DBDSN:     dsn,
		JWTSecret: jwtSecret,
		JWTTTL:    ttl,
	}
}
