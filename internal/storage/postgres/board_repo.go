package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/board"
)

// BoardRepository — реализация board.Repository поверх *sql.DB.
type BoardRepository struct {
	db *sql.DB
}

// NewBoardRepository создаёт репозиторий досок.
func NewBoardRepository(db *DB) *BoardRepository {
	return &BoardRepository{db: db.DB}
}

func (r *BoardRepository) ListByOwnerID(ctx context.Context, ownerID string) ([]*board.Board, error) {
	const q = `
        SELECT id, owner_id, name, created_at, updated_at
        FROM boards
        WHERE owner_id = $1
        ORDER BY created_at;
    `

	rows, err := r.db.QueryContext(ctx, q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*board.Board
	for rows.Next() {
		var b board.Board
		if err := rows.Scan(&b.ID, &b.OwnerID, &b.Name, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// Create создаёт доску для конкретного пользователя.
func (r *BoardRepository) Create(ctx context.Context, b *board.Board) error {
	const q = `
        INSERT INTO boards (owner_id, name)
        VALUES ($1, $2)
        RETURNING id, created_at, updated_at;
    `

	err := r.db.QueryRowContext(ctx, q, b.OwnerID, b.Name).
		Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// GetByID возвращает доску по id и владельцу.
func (r *BoardRepository) GetByID(ctx context.Context, id, ownerID string) (*board.Board, error) {
	const q = `
        SELECT id, owner_id, name, created_at, updated_at
        FROM boards
        WHERE id = $1 AND owner_id = $2;
    `

	var b board.Board
	err := r.db.QueryRowContext(ctx, q, id, ownerID).
		Scan(&b.ID, &b.OwnerID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, board.ErrNotFound
		}
		return nil, err
	}

	return &b, nil
}

// Update меняет название доски.
func (r *BoardRepository) Update(ctx context.Context, b *board.Board) error {
	const q = `
        UPDATE boards
        SET name = $1, updated_at = NOW()
        WHERE id = $2 AND owner_id = $3
        RETURNING updated_at;
    `

	err := r.db.QueryRowContext(ctx, q, b.Name, b.ID, b.OwnerID).
		Scan(&b.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return board.ErrNotFound
		}
		return err
	}

	return nil
}

// Delete удаляет доску пользователя.
func (r *BoardRepository) Delete(ctx context.Context, id, ownerID string) error {
	const q = `
        DELETE FROM boards
        WHERE id = $1 AND owner_id = $2;
    `

	res, err := r.db.ExecContext(ctx, q, id, ownerID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return board.ErrNotFound
	}

	return nil
}
