package column

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("column not found")

// Repository описывает операции хранилища, необходимые домену колонок.
type Repository interface {
	// ListByBoardOwner возвращает колонки доски владельца ownerID.
	ListByBoardOwner(ctx context.Context, boardID, ownerID string) ([]*Column, error)
	// CreateInBoard создаёт колонку в доске владельца ownerID.
	CreateInBoard(ctx context.Context, column *Column, boardID, ownerID string) error
	// Update обновляет колонку и проверяет владение доской.
	Update(ctx context.Context, c *Column, ownerID string) error
	// Delete удаляет колонку и проверяет владение доской.
	Delete(ctx context.Context, id, boardID, ownerID string) error
}
