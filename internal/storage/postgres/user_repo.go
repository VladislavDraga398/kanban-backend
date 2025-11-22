package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
)

// UserRepo структура для хранения пользователей
type UserRepo struct {
	DB *sql.DB
}

// Create создает пользователя
func (r *UserRepo) Create(ctx context.Context, user *user.User) error {
	const query = `
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
 RETURNING id created_at`

	return r.DB.QueryRowContext(ctx, query, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)

}

// GetByID получаем пользователя по ID
func (r *UserRepo) GetByID(ctx context.Context, id string) (*user.User, error) {
	const query = `SELECT id, email, password_hash FROM users WHERE id = $1`
	var u user.User
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	const query = `SELECT id, email, password_hash FROM users WHERE email = $1`
	var u user.User
	err := r.DB.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
