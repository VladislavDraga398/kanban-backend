package task

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("task not found")

// Repository описывает операции хранилища, необходимые домену задач.
type Repository interface {
	// ListByColumnOwner возвращает задачи колонки доски владельца ownerID.
	ListByColumnOwner(ctx context.Context, boardID, columnID, ownerID string) ([]*Task, error)
	// CreateInColumn создаёт задачу в колонке владельца ownerID.
	CreateInColumn(ctx context.Context, task *Task, boardID, columnID, ownerID string) error
	// Update обновляет задачу и проверяет владение доской.
	Update(ctx context.Context, task *Task, ownerID string) error
	// Delete удаляет задачу и проверяет владение доской.
	Delete(ctx context.Context, id, boardID, columnID, ownerID string) error
	// MoveToColumn переносит задачу в другую колонку и проверяет владение доской.
	MoveToColumn(ctx context.Context, task *Task, columnID, ownerID string) error
}
