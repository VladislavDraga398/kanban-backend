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
	// Обязательно проверяем, что колонка принадлежит доске владельца.
	Update(ctx context.Context, c *Column, ownerID string) error

	// Delete - Удаляем колонку с доски.
	// Обязательно проверяем принадлежность доски владельцу.
	Delete(ctx context.Context, id, boardID, ownerID string) error

	// ListByBoardOwner Список колонок конкретной доски пользователя.
	ListByBoardOwner(ctx context.Context, boardID, ownerID string) ([]*Column, error)

	// CreateInBoard Создаем колонку в конкретной доске пользователя.
	CreateInBoard(ctx context.Context, column *Column, boardID, ownerID string) error
}
