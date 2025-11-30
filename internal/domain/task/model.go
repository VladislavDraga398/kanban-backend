package task

import (
	"time"
)

// Task - Задача/Карточка на доске.
type Task struct {
	ID          string // UUID задачи
	BoardID     string // ID доски (для выборки задач доски)
	ColumnID    string // ID колонки, в которой находится задача
	Title       string // Короткий заголовок
	Description string // Подробное описание
	Position    int    // Позиция задачи в колонке
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
