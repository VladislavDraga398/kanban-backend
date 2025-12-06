package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/column"
)

// ColumnRepository — реализация column.Repository поверх *sql.DB.
type ColumnRepository struct {
	db *sql.DB
}

// NewColumnRepository создаёт репозиторий колонок.
func NewColumnRepository(db *DB) *ColumnRepository {
	return &ColumnRepository{db: db.DB}
}

// Create — простое создание колонки по board_id (без проверки владельца доски).
func (r *ColumnRepository) Create(ctx context.Context, c *column.Column) error {
	const q = `
		INSERT INTO columns (board_id, name, position)
		VALUES ($1, $2, COALESCE(
			(SELECT MAX(position) + 1 FROM columns WHERE board_id = $1),
			1
		))
		RETURNING id, position, created_at, updated_at;
	`

	return r.db.QueryRowContext(ctx, q, c.BoardID, c.Name).
		Scan(&c.ID, &c.Position, &c.CreatedAt, &c.UpdatedAt)
}

// ListByBoardID — все колонки по board_id (без проверки владельца).
func (r *ColumnRepository) ListByBoardID(ctx context.Context, boardID string) ([]column.Column, error) {
	const q = `
		SELECT id, board_id, name, position, created_at, updated_at
		FROM columns
		WHERE board_id = $1
		ORDER BY position, created_at;
	`

	rows, err := r.db.QueryContext(ctx, q, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []column.Column
	for rows.Next() {
		var c column.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Name, &c.Position, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// Update — обновляет имя и позицию колонки.
func (r *ColumnRepository) Update(ctx context.Context, c *column.Column) error {
	const q = `
		UPDATE columns
		SET name = $1,
		    position = COALESCE(NULLIF($2, 0), position),
		    updated_at = NOW()
		WHERE id = $3 AND board_id = $4
		RETURNING updated_at;
	`

	err := r.db.QueryRowContext(ctx, q, c.Name, c.Position, c.ID, c.BoardID).
		Scan(&c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return column.ErrNotFound
		}
		return err
	}

	return nil
}

// Delete — удаляет колонку по id и board_id.
func (r *ColumnRepository) Delete(ctx context.Context, id, boardID string) error {
	const q = `
		DELETE FROM columns
		WHERE id = $1 AND board_id = $2;
	`

	res, err := r.db.ExecContext(ctx, q, id, boardID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return column.ErrNotFound
	}

	return nil
}

// ListByBoardOwner — колонки доски, которая принадлежит ownerID.
func (r *ColumnRepository) ListByBoardOwner(ctx context.Context, boardID, ownerID string) ([]*column.Column, error) {
	const (
		q = `
		SELECT c.id, c.board_id, c.name, c.position, c.created_at, c.updated_at
		FROM columns c
		JOIN boards b ON c.board_id = b.id
		WHERE c.board_id = $1 AND b.owner_id = $2
		ORDER BY c.position, c.created_at;
	`
	)

	rows, err := r.db.QueryContext(ctx, q, boardID, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*column.Column
	for rows.Next() {
		var c column.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Name, &c.Position, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// CreateInBoard — создаёт колонку в доске конкретного пользователя.
func (r *ColumnRepository) CreateInBoard(ctx context.Context, c *column.Column, boardID, ownerID string) error {
	// 1) Проверяем, что доска существует и принадлежит ownerID.
	const checkBoard = `
		SELECT 1
		FROM boards
		WHERE id = $1 AND owner_id = $2;
	`

	var tmp int
	err := r.db.QueryRowContext(ctx, checkBoard, boardID, ownerID).Scan(&tmp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// доска не найдена или не принадлежит этому пользователю
			return column.ErrNotFound
		}
		return err
	}

	// 2) Находим следующую позицию в этой доске.
	const getPos = `
		SELECT COALESCE(MAX(position) + 1, 1)
		FROM columns
		WHERE board_id = $1;
	`

	var pos int
	if err := r.db.QueryRowContext(ctx, getPos, boardID).Scan(&pos); err != nil {
		return err
	}

	// 3) Вставляем колонку с рассчитанной позицией.
	const insert = `
		INSERT INTO columns (board_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at;
	`

	err = r.db.QueryRowContext(ctx, insert, boardID, c.Name, pos).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return err
	}

	c.BoardID = boardID
	c.Position = pos

	return nil
}
