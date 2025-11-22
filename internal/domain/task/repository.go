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

	// Update - Обновляем название задачи, (заголовок, описание, позицию, колонку)
	Update(ctx context.Context, task *Task) error

	// Delete - Удаляем задачу.
	Delete(ctx context.Context, id, broadID string) error
}
