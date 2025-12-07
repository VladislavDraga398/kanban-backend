-- Расширение для генерации UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
                                     id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     email        TEXT NOT NULL UNIQUE,
                                     password_hash TEXT NOT NULL,
                                     created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Таблица для досок
CREATE TABLE IF NOT EXISTS boards (
                                      id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                      owner_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                      name       TEXT NOT NULL,
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS boards_owner_id_idx ON boards(owner_id);

-- Таблица колонок
CREATE TABLE IF NOT EXISTS columns (
                                       id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                       board_id   UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
                                       name       TEXT NOT NULL,
                                       position   INT NOT NULL,
                                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                       CONSTRAINT uq_columns_board_position UNIQUE (board_id, position)
);

CREATE INDEX IF NOT EXISTS columns_board_id_idx ON columns(board_id);

-- Таблица задач
CREATE TABLE IF NOT EXISTS tasks (
                                     id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     board_id    UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
                                     column_id   UUID NOT NULL REFERENCES columns(id) ON DELETE CASCADE,
                                     title       TEXT NOT NULL,
                                     description TEXT NOT NULL DEFAULT '',
                                     position    INT NOT NULL,
                                     created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                     updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                     CONSTRAINT uq_tasks_column_position UNIQUE (column_id, position)
);

CREATE INDEX IF NOT EXISTS tasks_board_id_idx  ON tasks(board_id);
CREATE INDEX IF NOT EXISTS tasks_column_id_idx ON tasks(column_id);