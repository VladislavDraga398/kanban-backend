package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
)

// TaskRepository — реализация task.Repository поверх *sql.DB.
type TaskRepository struct {
	db *sql.DB
}

// Create создает задачу.
func (r *TaskRepository) Create(ctx context.Context, t *task.Task) error {
	const getPos = `
		SELECT COALESCE(MAX(position) + 1, 1)
		FROM tasks
		WHERE column_id = $1;
	`

	var pos int
	if err := r.db.QueryRowContext(ctx, getPos, t.ColumnID).Scan(&pos); err != nil {
		return err
	}

	const insert = `
		INSERT INTO tasks (board_id, column_id, title, description, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at;
	`

	if err := r.db.QueryRowContext(ctx, insert, t.BoardID, t.ColumnID, t.Title, t.Description, pos).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return err
	}

	t.Position = pos
	return nil
}

// ListByBoard — все задачи доски.
func (r *TaskRepository) ListByBoard(ctx context.Context, boardID string) ([]task.Task, error) {
	const (
		q = `
		SELECT id, board_id, column_id, title, description, position, created_at, updated_at
		FROM tasks
		WHERE board_id = $1
		ORDER BY position, created_at;
	`
	)
	rows, err := r.db.QueryContext(ctx, q, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []task.Task
	for rows.Next() {
		var t task.Task
		if err := rows.Scan(
			&t.ID,
			&t.BoardID,
			&t.ColumnID,
			&t.Title,
			&t.Description,
			&t.Position,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// ListByColumn — все задачи колонки.
func (r *TaskRepository) ListByColumn(ctx context.Context, columnID string) ([]task.Task, error) {
	const q = `
		SELECT id, board_id, column_id, title, description, position, created_at, updated_at
		FROM tasks
		WHERE column_id = $1
		ORDER BY position, created_at;
	`

	rows, err := r.db.QueryContext(ctx, q, columnID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var res []task.Task
	for rows.Next() {
		var t task.Task
		if err := rows.Scan(
			&t.ID,
			&t.BoardID,
			&t.ColumnID,
			&t.Title,
			&t.Description,
			&t.Position,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// Update обновляет задачу.
func (r *TaskRepository) Update(ctx context.Context, t *task.Task) error {
	const q = `
		UPDATE tasks
		SET column_id = $1,
		    title = $2,
		    description = $3,
		    position = COALESCE(NULLIF($4, 0), position),
		    updated_at = NOW()
		WHERE id = $5 AND board_id = $6
		RETURNING updated_at;
	`

	if err := r.db.QueryRowContext(
		ctx,
		q,
		t.ColumnID,
		t.Title,
		t.Description,
		t.Position,
		t.ID,
		t.BoardID,
	).Scan(&t.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task.ErrNotFound
		}
		return err
	}

	return nil
}

// Delete удаляет задачу по id, убеждаясь, что она принадлежит указанной доске и колонке.
func (r *TaskRepository) Delete(ctx context.Context, id, boardID, columnID string) error {
	const q = `
		DELETE FROM tasks
		WHERE id = $1 AND board_id = $2 AND column_id = $3;
	`

	res, err := r.db.ExecContext(ctx, q, id, boardID, columnID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return task.ErrNotFound
	}

	return nil
}

// NewTaskRepository создаёт репозиторий задач.
func NewTaskRepository(db *DB) *TaskRepository {
	return &TaskRepository{db: db.DB}
}

// ListByColumnOwner — все задачи колонки, если доска принадлежит ownerID.
func (r *TaskRepository) ListByColumnOwner(ctx context.Context, boardID, columnID, ownerID string) ([]*task.Task, error) {
	const q = `
		SELECT t.id,
		       t.board_id,
		       t.column_id,
		       t.title,
		       t.description,
		       t.position,
		       t.created_at,
		       t.updated_at
		FROM tasks t
		JOIN boards b ON t.board_id = b.id
		WHERE t.board_id = $1
		  AND t.column_id = $2
		  AND b.owner_id = $3
		ORDER BY t.position, t.created_at;
	`

	rows, err := r.db.QueryContext(ctx, q, boardID, columnID, ownerID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var res []*task.Task
	for rows.Next() {
		var tt task.Task
		if err := rows.Scan(
			&tt.ID,
			&tt.BoardID,
			&tt.ColumnID,
			&tt.Title,
			&tt.Description,
			&tt.Position,
			&tt.CreatedAt,
			&tt.UpdatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, &tt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// CreateInColumn — создать задачу в колонке конкретного пользователя.
func (r *TaskRepository) CreateInColumn(ctx context.Context, t *task.Task, boardID, columnID, ownerID string) error {
	// 1. Проверяем, что колонка принадлежит доске, а доска — пользователю.
	const check = `
		SELECT 1
		FROM columns c
		JOIN boards b ON c.board_id = b.id
		WHERE c.id = $1
		  AND c.board_id = $2
		  AND b.owner_id = $3;
	`

	var tmp int
	if err := r.db.QueryRowContext(ctx, check, columnID, boardID, ownerID).Scan(&tmp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// либо нет такой колонки, либо чужая доска
			return task.ErrNotFound
		}
		return err
	}

	// 2. Считаем позицию в колонке.
	const getPos = `
		SELECT COALESCE(MAX(position) + 1, 1)
		FROM tasks
		WHERE column_id = $1;
	`

	var pos int
	if err := r.db.QueryRowContext(ctx, getPos, columnID).Scan(&pos); err != nil {
		return err
	}

	// 3. Вставляем задачу.
	const insert = `
		INSERT INTO tasks (board_id, column_id, title, description, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at;
	`

	if err := r.db.QueryRowContext(ctx, insert, boardID, columnID, t.Title, t.Description, pos).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return err
	}

	t.BoardID = boardID
	t.ColumnID = columnID
	t.Position = pos

	return nil
}

// MoveToColumn — переместить задачу в другую колонку атомарно с корректировкой позиций.
func (r *TaskRepository) MoveToColumn(ctx context.Context, t *task.Task, newColumnID string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		// В случае паники откатываем транзакцию
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// 1) Прочитать задачу и залочить строку для корректного удаления из старой колонки.
	const selTask = `
        SELECT id, board_id, column_id, position, title, description, created_at, updated_at
        FROM tasks
        WHERE id = $1 AND board_id = $2
        FOR UPDATE;
    `
	var (
		curID, curBoardID, curColumnID string
		curPos                         int
	)
	var title, description string
	var createdAt, updatedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, selTask, t.ID, t.BoardID).Scan(
		&curID, &curBoardID, &curColumnID, &curPos, &title, &description, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return task.ErrNotFound
		}
		_ = tx.Rollback()
		return err
	}

	// Если целевая колонка совпадает с текущей — просто ставим в конец этой же колонки.
	// (Можно оптимизировать, но оставим логикой перемещения.)

	// 2) Проверить, что новая колонка относится к той же доске и залочить строку колонки.
	const checkCol = `
        SELECT id FROM columns WHERE id = $1 AND board_id = $2 FOR UPDATE;
    `
	var lockedColumnID string
	if err := tx.QueryRowContext(ctx, checkCol, newColumnID, curBoardID).Scan(&lockedColumnID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return task.ErrNotFound
		}
		_ = tx.Rollback()
		return err
	}

	// 3) Освободить позицию в исходной колонке (сдвинуть вниз все задачи правее удаляемой).
	const compactSrc = `
        UPDATE tasks
        SET position = position - 1
        WHERE column_id = $1 AND position > $2;
    `
	if _, err := tx.ExecContext(ctx, compactSrc, curColumnID, curPos); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 4) Найти позицию для вставки в новую колонку (в конец).
	const getDstPos = `
        SELECT COALESCE(MAX(position) + 1, 1)
        FROM tasks
        WHERE column_id = $1;
    `
	var newPos int
	if err := tx.QueryRowContext(ctx, getDstPos, newColumnID).Scan(&newPos); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 5) Обновить саму задачу: колонка, позиция, updated_at.
	const updTask = `
        UPDATE tasks
        SET column_id = $1,
            position  = $2,
            updated_at = NOW()
        WHERE id = $3 AND board_id = $4
        RETURNING id, board_id, column_id, title, description, position, created_at, updated_at;
    `
	if err := tx.QueryRowContext(ctx, updTask, newColumnID, newPos, curID, curBoardID).Scan(
		&t.ID,
		&t.BoardID,
		&t.ColumnID,
		&t.Title,
		&t.Description,
		&t.Position,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return task.ErrNotFound
		}
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}
