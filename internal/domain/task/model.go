package task

import (
	"time"
)

// Task - Задача/Карточка на доске.
type Task struct {
	ID          string // UUID Задачи
	BordID      string // ID доски (На случай выборки всех задач доски)
	ColumnID    string // ID колонки, в которой находится задача
	Title       string // Короткий заголовок
	Description string // Подробное описание
	Position    int    // Позиция задачи на колонке (Для сортировки сверху вниз)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
