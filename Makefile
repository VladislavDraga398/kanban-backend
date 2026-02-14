GO ?= go
APP ?= kanban-backend
PKG ?= ./cmd/api
GOCACHE ?= $(CURDIR)/.gocache
GOENV = env GOCACHE=$(GOCACHE)

.PHONY: help fmt vet tidy build clean run test test-integration cover db-up db-down migrate-up migrate-down docker-up docker-down docker-logs docker-rebuild frontend-install frontend-dev frontend-build frontend-lint frontend-test frontend-smoke

## fmt: форматирование кода
fmt:
	$(GOENV) $(GO) fmt ./...

## vet: статический анализ кода (go vet)
vet:
	$(GOENV) $(GO) vet ./...

## tidy: приведение зависимостей в порядок
tidy:
	$(GOENV) $(GO) mod tidy

## build: сборка бинарника ./bin/$(APP)
build: fmt vet tidy
	@mkdir -p bin
	$(GOENV) $(GO) build -o bin/$(APP) $(PKG)

## clean: удалить артефакты сборки
clean:
	rm -rf bin coverage.out

## run: запуск приложения локально (переменные окружения берутся из env/dev.env и/или .env, если файлы существуют)
run:
	@bash -c 'set -a; [ -f env/dev.env ] && . env/dev.env; [ -f .env ] && . .env; set +a; env GOCACHE=$(GOCACHE) $(GO) run $(PKG)'

## test: запустить все тесты (включая пакет tests)
test:
	$(GOENV) $(GO) test ./... ./tests

## test-integration: только интеграционный сценарий
test-integration:
	$(GOENV) $(GO) test ./tests -run TestIntegration_FullFlow -count=1

## cover: отчёт о покрытии тестами
cover:
	$(GOENV) $(GO) test ./tests -coverpkg=./... -coverprofile=coverage.out
	$(GOENV) $(GO) tool cover -func=coverage.out

## db-up: поднять Postgres через docker-compose
db-up:
	docker compose up -d db

## db-down: остановить Postgres из docker-compose
db-down:
	docker compose down

## migrate-up: применить миграции (нужен установленный psql и переменная DB_DSN)
migrate-up:
	@[ -n "$$DB_DSN" ] || (echo "DB_DSN is not set" && exit 1)
	psql "$$DB_DSN" -f migrations/0001_init.sql

## migrate-down: откат миграций (если предусмотрены down-скрипты)
migrate-down:
	@echo "No down migrations provided. Add a down SQL and update this target if needed."

## docker-up: запустить весь стек (БД + приложение) через docker-compose
docker-up:
	docker compose up -d --build

## docker-down: остановить весь стек и удалить контейнеры
docker-down:
	docker compose down

## docker-logs: показать логи всех сервисов
docker-logs:
	docker compose logs -f

## docker-rebuild: пересобрать и перезапустить приложение
docker-rebuild:
	docker compose up -d --build app

## frontend-install: установить зависимости фронтенда
frontend-install:
	npm --prefix frontend install

## frontend-dev: запустить фронтенд (Vite)
frontend-dev:
	npm --prefix frontend run dev

## frontend-build: собрать фронтенд
frontend-build:
	npm --prefix frontend run build

## frontend-lint: проверить фронтенд линтером
frontend-lint:
	npm --prefix frontend run lint

## frontend-test: запустить frontend unit/integration тесты
frontend-test:
	npm --prefix frontend run test

## frontend-smoke: fullstack smoke (frontend proxy + backend docker)
frontend-smoke:
	bash scripts/frontend-smoke.sh

## help: список команд Makefile
help:
	@awk 'BEGIN{FS=":.*"} /^## /{desc=substr($$0,4); next} /^[a-zA-Z0-9_.-]+:/{name=$$1; if(desc!=""){printf "%-20s %s\n", name, desc; desc=""}}' $(MAKEFILE_LIST)
