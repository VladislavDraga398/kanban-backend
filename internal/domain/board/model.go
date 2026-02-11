package board

import (
	"time"
)

// Board описывает канбан-доску.
type Board struct {
	ID        string
	OwnerID   string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
