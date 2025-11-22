package board

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("board not found")

// Repository - описываем, что домен ждет от хранилища досок.
type Repository interface {
	// Create - создание новой доски
	Create(ctx context.Context, b *Board) error

	// GetByID - Возвращает доску по ID и владельцу.
	// OwnerID - нужен для проверки прав доступа к доске.
	GetByID(ctx context.Context, id, ownerID string) (*Board, error)

	// ListByOwnerID - Возвращаем доски конкретного пользователя.
	ListByOwnerID(ctx context.Context, ownerID string) ([]*Board, error)

	//Delete - Удаляем доску по ID и владельцу.
	Delete(ctx context.Context, id, ownerID string) error
}
