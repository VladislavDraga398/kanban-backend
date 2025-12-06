package tests

import (
	"errors"
	"testing"

	"github.com/VladislavDraga398/kanban-backend/internal/domain/board"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/column"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/task"
	"github.com/VladislavDraga398/kanban-backend/internal/domain/user"
	pg "github.com/VladislavDraga398/kanban-backend/internal/storage/postgres"
)

func TestDomainErrSentinels(t *testing.T) {
	if !errors.Is(board.ErrNotFound, board.ErrNotFound) ||
		!errors.Is(column.ErrNotFound, column.ErrNotFound) ||
		!errors.Is(task.ErrNotFound, task.ErrNotFound) ||
		!errors.Is(user.ErrNotFound, user.ErrNotFound) ||
		!errors.Is(user.ErrEmailAlreadyUsed, user.ErrEmailAlreadyUsed) {
		t.Fatalf("sentinel errors should be comparable with errors.Is")
	}
}

func TestPostgresNewFailsOnBadDSN(t *testing.T) {
	if _, err := pg.New("not-a-valid-dsn"); err == nil {
		t.Fatalf("expected error for invalid dsn")
	}
}

func TestRepositoryConstructors(t *testing.T) {
	db := &pg.DB{}
	if pg.NewBoardRepository(db) == nil ||
		pg.NewColumnRepository(db) == nil ||
		pg.NewTaskRepository(db) == nil ||
		pg.NewUserRepository(db) == nil {
		t.Fatalf("expected non-nil repositories")
	}
}
