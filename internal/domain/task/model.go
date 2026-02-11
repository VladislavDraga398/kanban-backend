package task

import (
	"time"
)

// Task описывает карточку задачи в колонке доски.
type Task struct {
	ID          string
	BoardID     string
	ColumnID    string
	Title       string
	Description string
	Position    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
