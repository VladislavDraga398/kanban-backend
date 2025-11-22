package user

import (
	"context"
	"errors"
)

// ErrNotFound - ошибка, если пользователь не найден
var ErrNotFound = errors.New("user not found")

// Repository - описываем, что домен ждет от хранилища пользователей
type Repository interface {
	// Create - создание нового пользователя
	Create(ctx context.Context, user *User) error

	// GetByID - Находим пользователя по ID
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByEmail - Находим пользователя по Email (Вместо логина)
	GetByEmail(ctx context.Context, email string) (*User, error)
}
