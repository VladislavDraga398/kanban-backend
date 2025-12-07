# Kanban Backend

REST API для досок/колонок/задач. Минимальная аутентификация через JWT.

## Требования
- Go (1.22+)
- Docker + Docker Compose (для локальной БД и интеграционных тестов)
- make

## Быстрый старт
1) Поднять Postgres через Docker Compose:
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
- `DB_DSN` — строка подключения к Postgres. Для docker-compose: `postgres://kanban:kanban@localhost:5432/kanban?sslmode=disable`.
- `JWT_SECRET` — секрет для подписи JWT, обязательно непустой.
- `JWT_TTL` — срок жизни токена, например `24h`.

Пример `env/dev.env`:
```env
HTTP_PORT=8083
DB_DSN=postgres://kanban:kanban@localhost:5432/kanban?sslmode=disable
JWT_SECRET=change-me-please
JWT_TTL=24h
```

Примечания:
- Рабочие файлы с секретами (`env/*.env`) не коммитим. В репозитории добавлен `env/.gitignore`, который игнорирует реальные `*.env` и оставляет только примеры `*.example.env`.
- Команда `make run` загружает переменные из `env/dev.env`, затем из `.env` (если существуют), и запускает приложение.

## Сборка
```bash
make build   # соберёт бинарник в ./bin/kanban-backend
```

## Полезные команды Makefile
```bash
make help            # список целей с описанием
make fmt vet tidy    # форматирование / проверка / tidy модулей
make run             # запуск (учитывает .env, если есть)
make test            # все тесты (нужен Docker для интеграции)
make test-integration# только интеграционный сценарий
make cover           # отчёт о покрытии тестами
make db-up           # поднять БД через docker-compose
make db-down         # остановить контейнеры
make migrate-up      # применить миграции через psql к DB_DSN
```

Docker Compose уже применяет миграции на старте контейнера (`migrations/` монтируются в init), так что для локального окружения обычно достаточно `make db-up`.

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
