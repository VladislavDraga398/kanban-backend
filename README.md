# Kanban Backend

REST API для досок/колонок/задач. Минимальная аутентификация через JWT.

## Требования
- Go (1.25+)
- Docker + Docker Compose (для локальной БД и интеграционных тестов)
- make

## Быстрый старт

### Вариант 1: Полный стек в Docker (рекомендуется)
Самый простой способ — запустить всё через Docker Compose:

```bash
make docker-up
```

Эта команда:
- Соберёт Docker-образ приложения
- Поднимет PostgreSQL
- Применит миграции автоматически
- Запустит приложение с правильной конфигурацией

Проверка healthcheck:
```bash
curl -s http://localhost:8083/healthz
```

Просмотр логов:
```bash
make docker-logs
```

Остановка всего стека:
```bash
make docker-down
```

### Вариант 2: Локальная разработка (Go на хосте)
Если хотите запускать приложение напрямую через Go:

1) Поднять только PostgreSQL:
```bash
make db-up
```

2) Создать файл с переменными окружения — рекомендуется `env/dev.env` (или корневой `.env`). Есть пример `env/dev.example.env`, можно скопировать: `cp env/dev.example.env env/dev.env`. Приложение автоматически подхватит `env/dev.env`, а затем (при наличии) `.env`.

3) Запустить сервер локально:
```bash
make run
```

Проверка healthcheck:
```bash
curl -s http://localhost:8083/healthz
```

## Конфигурация (переменные окружения)
- `HTTP_PORT` — порт HTTP (по умолчанию `8083`).
- `DB_DSN` — строка подключения к Postgres:
  - **Для локальной разработки** (Go на хосте): `postgres://kanban:kanban@localhost:5432/kanban?sslmode=disable`
  - **Для Docker Compose**: `postgres://kanban:kanban@db:5432/kanban?sslmode=disable` (использует имя сервиса "db")
- `JWT_SECRET` — секрет для подписи JWT, обязательно непустой.
- `JWT_TTL` — срок жизни токена, например `24h`.

Пример `env/dev.env` для локальной разработки:
```env
HTTP_PORT=8083
DB_DSN=postgres://kanban:kanban@localhost:5432/kanban?sslmode=disable
JWT_SECRET=change-me-please
JWT_TTL=24h
```

**Важно:** При использовании `make docker-up` переменные окружения настроены автоматически в `docker-compose.yml`, создавать `env/dev.env` не нужно.

Примечания:
- Рабочие файлы с секретами (`env/*.env`) не коммитим. В репозитории добавлен `env/.gitignore`, который игнорирует реальные `*.env` и оставляет только примеры `*.example.env`.
- Команда `make run` загружает переменные из `env/dev.env`, затем из `.env` (если существуют), и запускает приложение.

## Сборка
```bash
make build   # соберёт бинарник в ./bin/kanban-backend
```

## Полезные команды Makefile

### Docker Compose команды (полный стек)
```bash
make docker-up       # запустить весь стек (БД + приложение)
make docker-down     # остановить весь стек и удалить контейнеры
make docker-logs     # показать логи всех сервисов (с -f для follow)
make docker-rebuild  # пересобрать и перезапустить приложение
```

### Локальная разработка
```bash
make help            # список целей с описанием
make fmt vet tidy    # форматирование / проверка / tidy модулей
make build           # собрать бинарник в ./bin/kanban-backend
make run             # запуск (учитывает .env, если есть)
make db-up           # поднять только БД через docker-compose
make db-down         # остановить контейнеры БД
make migrate-up      # применить миграции через psql к DB_DSN
```

### Тестирование
```bash
make test            # все тесты (нужен Docker для интеграции)
make test-integration# только интеграционный сценарий
make cover           # отчёт о покрытии тестами
```

**Примечание:** Docker Compose автоматически применяет миграции при старте БД (`migrations/` монтируются в `/docker-entrypoint-init-db.d`).

## Аутентификация
1. Зарегистрироваться: `POST /api/v1/auth/register` → в ответе придёт `token`.
2. Авторизоваться: `POST /api/v1/auth/login` → вернёт `token`.
3. Передавать `Authorization: Bearer <token>` ко всем защищённым ручкам.

Примеры запросов:
```bash
# Регистрация
curl -s -X POST http://localhost:8083/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"pass123"}' | jq

# Логин
curl -s -X POST http://localhost:8083/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"pass123"}' | jq
```

## Основные маршруты
- `GET /api/v1/boards`, `POST /api/v1/boards`, `GET/PUT/DELETE /api/v1/boards/{id}`
- `GET/POST /api/v1/boards/{board_id}/columns`, `PUT/DELETE /api/v1/boards/{board_id}/columns/{column_id}`
- `GET/POST /api/v1/boards/{board_id}/columns/{column_id}/tasks`, `PUT/DELETE /api/v1/boards/{board_id}/columns/{column_id}/tasks/{task_id}`
- `PATCH /api/v1/boards/{board_id}/tasks/{task_id}/move`

## Тесты
- Все тесты: `make test` (для интеграционных тестов требуется Docker, контейнер Postgres поднимется автоматически через testcontainers).
- Только интеграция: `make test-integration`.
- Покрытие: `make cover`.

## Остановка сервисов
```bash
make db-down
```

## Примечания
- В директории `internal/http` используется пакет с именем `http`, что корректно, но может конфликтовать в импортах со стандартным `net/http`. Это известное решение; возможен рефакторинг в будущем.
- Обязательно задайте надёжный `JWT_SECRET` в продакшене.
