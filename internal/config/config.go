package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr  string
	DBDSN     string
	JWTSecret string
	JWTTTL    time.Duration
}

func Load() (*Config, error) {
	// Пробуем загрузить переменные из локальных env-файлов.
	// Ошибка нефатальная для отсутствующих файлов: подставляем только отсутствующие переменные.
	if err := loadEnvFiles(); err != nil {
		return nil, fmt.Errorf("load env files: %w", err)
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8083"
	}
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return nil, fmt.Errorf("invalid HTTP_PORT: %q", port)
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://kanban:kanban@localhost:5432/kanban?sslmode=disable"
	}

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	ttl := 24 * time.Hour
	if ttlStr := os.Getenv("JWT_TTL"); ttlStr != "" {
		parsed, err := time.ParseDuration(ttlStr)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_TTL: %w", err)
		}
		ttl = parsed
	}
	if ttl <= 0 {
		return nil, errors.New("JWT_TTL must be greater than 0")
	}

	return &Config{
		HTTPAddr:  ":" + port,
		DBDSN:     dsn,
		JWTSecret: jwtSecret,
		JWTTTL:    ttl,
	}, nil
}

// loadEnvFiles загружает переменные из .env и env/dev.env, если файлы существуют.
// Уже заданные в окружении переменные не перезаписываются.
func loadEnvFiles() error {
	var firstErr error

	// Проверяем текущую директорию и стандартный путь проекта.
	candidates := []string{
		".env",
		filepath.Join("env", "dev.env"),
	}

	for _, p := range candidates {
		if err := loadEnvFile(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			// Сохраняем первую ошибку, кроме "файл не существует".
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Простой парсер KEY=VALUE без обработки кавычек.
		if eq := strings.IndexByte(line, '='); eq != -1 {
			key := strings.TrimSpace(line[:eq])
			val := strings.TrimSpace(line[eq+1:])
			if key == "" {
				continue
			}
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
	}
	return scanner.Err()
}
