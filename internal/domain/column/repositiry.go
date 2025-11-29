package column

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("column not found")

// Repository - описываем, что домен ждет от хранилища колонок
type Repository interface {
	// Create - создание новой колонки
	Create(ctx context.Context, c *Column) error

	// ListByBoardID - Возвращаем все колонки доски, отсортированные по порядку.
	ListByBoardID(ctx context.Context, boardID string) ([]Column, error)

	// Update - Обновляем имя и позицию колонки.
	Update(ctx context.Context, c *Column) error

	// Delete - Удаляем колонку с доски.
	Delete(ctx context.Context, id, boardID string) error

	// ListByBoardOwner Список колодно к конкретной доске пользователя.
	ListByBoardOwner(ctx context.Context, boardID, ownerID string) ([]*Column, error)

	// CreateInBoard Создаем колонку в конкретной доске пользователя.
	CreateInBoard(ctx context.Context, column *Column, boardID, ownerID string) error
}
