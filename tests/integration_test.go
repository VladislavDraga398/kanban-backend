package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	myhttp "github.com/VladislavDraga398/kanban-backend/internal/http"
	pg "github.com/VladislavDraga398/kanban-backend/internal/storage/postgres"
)

func startPostgres(t *testing.T) (dsn string, stop func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "kanban",
			"POSTGRES_PASSWORD": "kanban",
			"POSTGRES_DB":       "kanban",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("mapped port: %v", err)
	}

	dsn = fmt.Sprintf("postgres://kanban:kanban@%s:%s/kanban?sslmode=disable", host, mappedPort.Port())

	stop = func() {
		_ = container.Terminate(ctx)
	}
	return dsn, stop
}

func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
	sqlBytes, err := os.ReadFile(filepath.Join(root, "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}

type authResp struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func doJSON(t *testing.T, client *http.Client, method, url string, body any, token string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return out
}

// Integration: full HTTP flow against real Postgres in container.
func TestIntegration_FullFlow(t *testing.T) {
	dsn, stop := startPostgres(t)
	defer stop()

	db, err := pg.New(dsn)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	defer db.Close()
	applyMigrations(t, db.DB)

	router := myhttp.NewRouter(myhttp.Deps{
		UserRepo:   pg.NewUserRepository(db),
		BoardRepo:  pg.NewBoardRepository(db),
		ColumnRepo: pg.NewColumnRepository(db),
		TaskRepo:   pg.NewTaskRepository(db),
		JWTSecret:  "integration-secret",
		JWTTTL:     time.Hour,
	})

	srv := httptest.NewServer(router)
	defer srv.Close()
	client := srv.Client()

	// register
	resp := doJSON(t, client, http.MethodPost, srv.URL+"/api/v1/auth/register", map[string]string{
		"email":    "user@example.com",
		"password": "password123",
	}, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register status: %d", resp.StatusCode)
	}
	register := decode[authResp](t, resp)
	if register.Token == "" || register.ID == "" {
		t.Fatalf("register response missing fields: %+v", register)
	}

	token := register.Token

	// create board
	resp = doJSON(t, client, http.MethodPost, srv.URL+"/api/v1/boards", map[string]string{"name": "Board 1"}, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create board status: %d", resp.StatusCode)
	}
	board := decode[struct {
		ID string `json:"id"`
	}](t, resp)

	// create columns
	var columns []string
	for _, name := range []string{"Todo", "In Progress"} {
		url := fmt.Sprintf("%s/api/v1/boards/%s/columns", srv.URL, board.ID)
		resp = doJSON(t, client, http.MethodPost, url, map[string]string{"name": name}, token)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create column status: %d", resp.StatusCode)
		}
		col := decode[struct {
			ID string `json:"id"`
		}](t, resp)
		columns = append(columns, col.ID)
	}

	// create task in first column
	taskURL := fmt.Sprintf("%s/api/v1/boards/%s/columns/%s/tasks", srv.URL, board.ID, columns[0])
	resp = doJSON(t, client, http.MethodPost, taskURL, map[string]string{"title": "Task 1", "description": "desc"}, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create task status: %d", resp.StatusCode)
	}
	taskResp := decode[struct {
		ID string `json:"id"`
	}](t, resp)

	// move task to second column
	moveURL := fmt.Sprintf("%s/api/v1/boards/%s/tasks/%s/move", srv.URL, board.ID, taskResp.ID)
	resp = doJSON(t, client, http.MethodPatch, moveURL, map[string]string{"column_id": columns[1]}, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("move task status: %d", resp.StatusCode)
	}
	moveResp := decode[struct {
		ColumnID string `json:"column_id"`
		Position int    `json:"position"`
	}](t, resp)
	if moveResp.ColumnID != columns[1] {
		t.Fatalf("task column mismatch: %s", moveResp.ColumnID)
	}
	if moveResp.Position == 0 {
		t.Fatalf("expected position to be set")
	}

	// list tasks in second column
	listURL := fmt.Sprintf("%s/api/v1/boards/%s/columns/%s/tasks", srv.URL, board.ID, columns[1])
	resp = doJSON(t, client, http.MethodGet, listURL, nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list tasks status: %d", resp.StatusCode)
	}
	tasks := decode[[]struct {
		ID string `json:"id"`
	}](t, resp)
	if len(tasks) != 1 || tasks[0].ID != taskResp.ID {
		t.Fatalf("unexpected tasks list: %+v", tasks)
	}
}
