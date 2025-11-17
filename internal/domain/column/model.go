package column

import "time"

// Column - Колонка на доске (Task) (Progress) (To do)
type column struct {
	ID        string // UUID колонки
	BoardID   string // ID доски, к которой относится колонка
	Name      string // Название колонки
	Position  int    // Позиция колонки на доске, (Для сортировки слева на право)
	CreatedAt time.Time
	UpdatedAt time.Time
}
