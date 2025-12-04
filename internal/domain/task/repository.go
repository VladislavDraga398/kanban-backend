package task

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("task not found")

// Repository - интерфейс для хранилища задач
type Repository interface {
	// Create - создание новой задачи
	Create(ctx context.Context, task *Task) error

	// ListByBoard - Возвращаем все задачи доски, для быстрого отображения.
	ListByBoard(ctx context.Context, boardID string) ([]Task, error)

	// ListByColumn - Возвращаем задачи для конкретной колонки.
	ListByColumn(ctx context.Context, columnID string) ([]Task, error)

	// ListByColumnOwner - Возвращаем задачи для конкретной колонки и владельца.
	ListByColumnOwner(ctx context.Context, columnID, ownerID, id string) ([]*Task, error)

	// CreateInColumn - Создаем задачу в конкретной колонке для конкретного пользователя.
	CreateInColumn(ctx context.Context, task *Task, columnID, ownerID, id string) error

	// Update - Обновляем название задачи, (заголовок, описание, позицию, колонку)
	Update(ctx context.Context, task *Task) error

	// Delete - Удаляем задачу.
	Delete(ctx context.Context, id, boardID, columnID string) error
}
