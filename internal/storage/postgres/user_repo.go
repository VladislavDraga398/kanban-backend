package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *DB) user.Repository {
	return &UserRepository{db: db.DB}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at;
	`

	err := r.db.QueryRowContext(ctx, q, u.Email, u.PasswordHash).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		// Нарушение UNIQUE (email) → бизнес-ошибка домена
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.ErrEmailAlreadyUsed
		}
		return err
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1;
	`

	var u user.User
	err := r.db.QueryRowContext(ctx, q, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1;
	`

	var u user.User
	err := r.db.QueryRowContext(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}
