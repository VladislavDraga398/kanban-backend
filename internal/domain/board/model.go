package board

import (
	"time"
)

// Board - Кабан-Доска.
type Board struct {
	ID        string    // UUID Доски
	OwnerID   string    // UUID Пользователя-владельца
	Name      string    // Название доски
	CreatedAt time.Time // Когда мы создали доску
	UpdatedAt time.Time // Когда были последние изменения в доске
}
