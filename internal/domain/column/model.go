package column

import "time"

// Column описывает колонку доски.
type Column struct {
	ID        string
	BoardID   string
	Name      string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
