package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
)

type sqlScanner interface {
	Scan(dest ...any) error
}

func labelsJSON(labels []string) (string, error) {
	if labels == nil {
		labels = []string{}
	}
	b, err := json.Marshal(labels)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func scanTask(s sqlScanner, t *task.Task) error {
	var (
		labelsRaw string
		dueDate   sql.NullTime
	)
	if err := s.Scan(
		&t.ID,
		&t.BoardID,
		&t.ColumnID,
		&t.Title,
		&t.Description,
		&t.Priority,
		&labelsRaw,
		&dueDate,
		&t.Position,
		&t.CreatedAt,
		&t.UpdatedAt,
	); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(labelsRaw), &t.Labels); err != nil {
		return err
	}
	if t.Labels == nil {
		t.Labels = []string{}
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	} else {
		t.DueDate = nil
	}
	return nil
}

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
		INSERT INTO tasks (board_id, column_id, title, description, priority, labels, due_date, position)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		RETURNING id, board_id, column_id, title, description, priority, labels::text, due_date, position, created_at, updated_at;
	`

	labels, err := labelsJSON(t.Labels)
	if err != nil {
		return err
	}
	if t.Priority == "" {
		t.Priority = "normal"
	}

	if err := scanTask(
		r.db.QueryRowContext(ctx, insert, t.BoardID, t.ColumnID, t.Title, t.Description, t.Priority, labels, t.DueDate, pos),
		t,
	); err != nil {
		return err
	}

	return nil
}

// ListByBoard — все задачи доски.
func (r *TaskRepository) ListByBoard(ctx context.Context, boardID string) ([]task.Task, error) {
	const (
		q = `
		SELECT id, board_id, column_id, title, description, priority, labels::text, due_date, position, created_at, updated_at
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
		if err := scanTask(rows, &t); err != nil {
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
		SELECT id, board_id, column_id, title, description, priority, labels::text, due_date, position, created_at, updated_at
		FROM tasks
		WHERE column_id = $1
		ORDER BY position, created_at;
	`

	rows, err := r.db.QueryContext(ctx, q, columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []task.Task
	for rows.Next() {
		var t task.Task
		if err := scanTask(rows, &t); err != nil {
			return nil, err
		}
		res = append(res, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// Update обновляет задачу, проверяя владельца доски.
func (r *TaskRepository) Update(ctx context.Context, t *task.Task, ownerID string) error {
	const q = `
		UPDATE tasks AS t
		SET column_id = $1,
		    title = $2,
		    description = $3,
		    priority = CASE WHEN $4 THEN $5 ELSE t.priority END,
		    labels = CASE WHEN $6 THEN $7::jsonb ELSE t.labels END,
		    due_date = CASE WHEN $8 THEN $9 ELSE t.due_date END,
		    position = COALESCE(NULLIF($10, 0), t.position),
		    updated_at = NOW()
		FROM boards b
		WHERE t.id = $11
		  AND t.board_id = $12
		  AND b.id = t.board_id
		  AND b.owner_id = $13
		RETURNING t.id, t.board_id, t.column_id, t.title, t.description, t.priority, t.labels::text, t.due_date, t.position, t.created_at, t.updated_at;
	`

	labels, err := labelsJSON(t.Labels)
	if err != nil {
		return err
	}

	if err := scanTask(r.db.QueryRowContext(
		ctx, q,
		t.ColumnID,
		t.Title,
		t.Description,
		t.PrioritySet,
		t.Priority,
		t.LabelsSet,
		labels,
		t.DueDateSet,
		t.DueDate,
		t.Position,
		t.ID,
		t.BoardID,
		ownerID,
	), t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task.ErrNotFound
		}
		return err
	}

	return nil
}

// Delete удаляет задачу по id, убеждаясь, что она принадлежит указанной доске и колонке, и доска принадлежит ownerID.
func (r *TaskRepository) Delete(ctx context.Context, id, boardID, columnID, ownerID string) error {
	const q = `
		DELETE FROM tasks AS t
		USING boards b
		WHERE t.id = $1
		  AND t.board_id = $2
		  AND t.column_id = $3
		  AND b.id = t.board_id
		  AND b.owner_id = $4;
	`

	res, err := r.db.ExecContext(ctx, q, id, boardID, columnID, ownerID)
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
		       t.priority,
		       t.labels::text,
		       t.due_date,
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
	defer rows.Close()

	var res []*task.Task
	for rows.Next() {
		var tt task.Task
		if err := scanTask(rows, &tt); err != nil {
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
	const insert = `
		WITH locked_column AS (
			SELECT c.id, c.board_id
			FROM columns c
			JOIN boards b ON c.board_id = b.id
			WHERE c.id = $1
			  AND c.board_id = $2
			  AND b.owner_id = $3
			FOR UPDATE
		),
		next_pos AS (
			SELECT COALESCE(MAX(t.position) + 1, 1) AS pos
			FROM tasks t
			JOIN locked_column lc ON t.column_id = lc.id
		)
		INSERT INTO tasks (board_id, column_id, title, description, priority, labels, due_date, position)
		SELECT lc.board_id, lc.id, $4, $5, $6, $7::jsonb, $8, np.pos
		FROM locked_column lc
		CROSS JOIN next_pos np
		RETURNING id, board_id, column_id, title, description, priority, labels::text, due_date, position, created_at, updated_at;
	`

	labels, err := labelsJSON(t.Labels)
	if err != nil {
		return err
	}
	if t.Priority == "" {
		t.Priority = "normal"
	}

	if err := scanTask(
		r.db.QueryRowContext(ctx, insert, columnID, boardID, ownerID, t.Title, t.Description, t.Priority, labels, t.DueDate),
		t,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task.ErrNotFound
		}
		return err
	}

	return nil
}

// MoveToColumn — переместить задачу в другую колонку атомарно с корректировкой позиций и проверкой владельца доски.
func (r *TaskRepository) MoveToColumn(ctx context.Context, t *task.Task, newColumnID, ownerID string) error {
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

	// 1) Убедиться, что доска принадлежит ownerID.
	const checkBoard = `
        SELECT 1 FROM boards WHERE id = $1 AND owner_id = $2 FOR UPDATE;
    `
	if err := tx.QueryRowContext(ctx, checkBoard, t.BoardID, ownerID).Scan(new(int)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return task.ErrNotFound
		}
		_ = tx.Rollback()
		return err
	}

	// 2) Прочитать задачу и залочить строку для корректного удаления из старой колонки.
	const selTask = `
        SELECT id, board_id, column_id, position, title, description, priority, labels::text, due_date, created_at, updated_at
        FROM tasks
        WHERE id = $1 AND board_id = $2
        FOR UPDATE;
    `
	var (
		curID, curBoardID, curColumnID string
		curPos                         int
	)
	var title, description, priority, labelsRaw string
	var createdAt, updatedAt sql.NullTime
	var dueDate sql.NullTime
	if err := tx.QueryRowContext(ctx, selTask, t.ID, t.BoardID).Scan(
		&curID, &curBoardID, &curColumnID, &curPos, &title, &description, &priority, &labelsRaw, &dueDate, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return task.ErrNotFound
		}
		_ = tx.Rollback()
		return err
	}

	// Если перенос в ту же колонку, оставляем порядок как есть и возвращаем текущее состояние задачи.
	if newColumnID == curColumnID {
		t.ID = curID
		t.BoardID = curBoardID
		t.ColumnID = curColumnID
		t.Title = title
		t.Description = description
		t.Priority = priority
		if err := json.Unmarshal([]byte(labelsRaw), &t.Labels); err != nil {
			_ = tx.Rollback()
			return err
		}
		if t.Labels == nil {
			t.Labels = []string{}
		}
		if dueDate.Valid {
			t.DueDate = &dueDate.Time
		}
		t.Position = curPos
		if createdAt.Valid {
			t.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			t.UpdatedAt = updatedAt.Time
		}
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return err
		}
		return nil
	}

	// 3) Проверить, что новая колонка относится к той же доске и залочить строку колонки.
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

	// 4) Освободить позицию в исходной колонке (сдвинуть вниз все задачи правее удаляемой).
	const compactSrc = `
        UPDATE tasks
        SET position = position - 1
        WHERE column_id = $1 AND position > $2;
    `
	if _, err := tx.ExecContext(ctx, compactSrc, curColumnID, curPos); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 5) Найти позицию для вставки в новую колонку (в конец).
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

	// 6) Обновить саму задачу: колонка, позиция и updated_at.
	const updTask = `
        UPDATE tasks
        SET column_id = $1,
            position  = $2,
            updated_at = NOW()
        WHERE id = $3 AND board_id = $4
        RETURNING id, board_id, column_id, title, description, priority, labels::text, due_date, position, created_at, updated_at;
    `
	if err := scanTask(tx.QueryRowContext(ctx, updTask, newColumnID, newPos, curID, curBoardID), t); err != nil {
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
