# Kanban Frontend

React + TypeScript + Vite клиент для `kanban-backend`.

## Что реализовано

- auth: `register/login/logout`;
- boards: list/create/update/delete;
- board details:
  - columns: list/create/update/delete;
  - tasks: list/create/update/delete;
  - drag-and-drop задач между колонками (`PATCH /boards/{board_id}/tasks/{task_id}/move`).

## Быстрый старт

```bash
npm install
npm run dev
```

Frontend поднимается на `http://localhost:5173`.

## Конфиг

Скопируй пример:

```bash
cp .env.example .env
```

Переменные:

- `VITE_API_URL` — базовый путь API (по умолчанию `/api/v1`).
- `VITE_BACKEND_ORIGIN` — backend origin для dev proxy (по умолчанию `http://localhost:8083`).

## Скрипты

```bash
npm run dev
npm run build
npm run lint
npm run test
npm run test:watch
npm run preview
```

## Проверка совместимости с backend

Из корня репозитория можно прогнать полный smoke-сценарий через Docker:

```bash
make frontend-smoke
```
