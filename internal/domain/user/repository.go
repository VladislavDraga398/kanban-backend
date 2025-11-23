package user

import (
	"context"
	"errors"
)

var (
	ErrNotFound         = errors.New("user not found")
	ErrEmailAlreadyUsed = errors.New("email already in use")
)

type Repository interface {
	// Create - создание нового пользователя
	Create(ctx context.Context, u *User) error
	// GetByID - получение пользователя по ID
	GetByID(ctx context.Context, id string) (*User, error)
	// GetByEmail - получение пользователя по email
	GetByEmail(ctx context.Context, email string) (*User, error)
}
