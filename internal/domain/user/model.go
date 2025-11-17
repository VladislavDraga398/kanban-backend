package user

import "time"

// User Структура пользователя

type User struct {
	ID           int       // UUID Пользователя
	Email        string    // Уникальный Email, будет использоваться как логин
	PasswordHash string    // Хранит только хеш пароля, но не сам пароль
	CreatedAt    time.Time // Дата создания пользователя
}
