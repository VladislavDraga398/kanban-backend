package board

import (
	"time"
)

// Board - Кабан-Доска.
type Board struct {
	OwnerID   string    // UUID Доски
	Name      string    // ID пользователя владельца (User.ID)
	CreatedAt time.Time // Название доски
	UpdatedAt time.Time // Когда были последние изменения в доске
}
