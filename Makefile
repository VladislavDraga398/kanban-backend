GO ?= go
APP ?= kanban-backend
PKG ?= ./cmd/api

.PHONY: help fmt vet tidy build clean run test test-integration cover db-up db-down migrate-up migrate-down

## fmt: форматирование кода
fmt:
	$(GO) fmt ./...

## vet: статический анализ кода (go vet)
vet:
	$(GO) vet ./...

## tidy: приведение зависимостей в порядок
tidy:
	$(GO) mod tidy

## build: сборка бинарника ./bin/$(APP)
build: fmt vet tidy
	@mkdir -p bin
	$(GO) build -o bin/$(APP) $(PKG)

## clean: удалить артефакты сборки
clean:
	rm -rf bin coverage.out

## run: запуск приложения локально (переменные окружения берутся из env/dev.env и/или .env, если файлы существуют)
run:
	@bash -c 'set -a; [ -f env/dev.env ] && . env/dev.env; [ -f .env ] && . .env; set +a; $(GO) run $(PKG)'

## test: запустить все тесты (включая пакет tests)
test:
	$(GO) test ./... ./tests

## test-integration: только интеграционный сценарий
test-integration:
	$(GO) test ./tests -run TestIntegration_FullFlow -count=1

## cover: отчёт о покрытии тестами
cover:
	$(GO) test ./... ./tests -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

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

## help: список команд Makefile
help:
	@awk 'BEGIN{FS=":.*"} /^## /{desc=substr($$0,4); next} /^[a-zA-Z0-9_.-]+:/{name=$$1; if(desc!=""){printf "%-20s %s\n", name, desc; desc=""}}' $(MAKEFILE_LIST)
