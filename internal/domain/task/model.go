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
	Priority    string
	PrioritySet bool
	Labels      []string
	LabelsSet   bool
	DueDate     *time.Time
	DueDateSet  bool
	Position    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
