package task

import (
	"time"
)

// Task - Задача/Карточка на доске.
type Task struct {
	ID          string // UUID Задачи
	Bord        string // ID доски (На случай выборки всех задач доски)
	Column      string // ID колонки, в которой находится задача
	Title       string // Короткий заголовок
	Description string // Подробное описание
	Position    int    // Позиция задачи на колонке (Для сортировки сверху вниз)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
