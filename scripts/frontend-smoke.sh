#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_HOST="${FRONTEND_HOST:-127.0.0.1}"
FRONTEND_PORT="${FRONTEND_PORT:-5173}"
BACKEND_PORT="${BACKEND_PORT:-8083}"
FRONTEND_PID=""
FRONTEND_LOG=""

cleanup() {
  if [[ -n "${FRONTEND_PID}" ]] && kill -0 "${FRONTEND_PID}" >/dev/null 2>&1; then
    kill "${FRONTEND_PID}" >/dev/null 2>&1 || true
    wait "${FRONTEND_PID}" >/dev/null 2>&1 || true
  fi
  if [[ -n "${FRONTEND_LOG}" ]] && [[ -f "${FRONTEND_LOG}" ]]; then
    rm -f "${FRONTEND_LOG}"
  fi
  (cd "${ROOT_DIR}" && docker compose down >/dev/null 2>&1) || true
}
trap cleanup EXIT

wait_for_url() {
  local url="$1"
  local title="$2"
  local attempts="$3"
  for ((i = 1; i <= attempts; i += 1)); do
    if curl -fsS "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "ERROR: ${title} is not ready: ${url}" >&2
  return 1
}

json_field() {
  local field="$1"
  node -e '
let raw = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => {
  raw += chunk;
});
process.stdin.on("end", () => {
  try {
    const data = JSON.parse(raw || "{}");
    const value = data[process.argv[1]];
    if (value === undefined || value === null || value === "") {
      process.exit(2);
      return;
    }
    process.stdout.write(String(value));
  } catch {
    process.exit(1);
  }
});
' "${field}"
}

assert_board_in_list() {
  local board_id="$1"
  node -e '
let raw = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => {
  raw += chunk;
});
process.stdin.on("end", () => {
  try {
    const data = JSON.parse(raw || "[]");
    if (!Array.isArray(data)) {
      process.exit(2);
      return;
    }
    const found = data.some((item) => item && item.id === process.argv[1]);
    if (!found) {
      process.exit(3);
      return;
    }
  } catch {
    process.exit(1);
  }
});
' "${board_id}"
}

assert_task_in_list() {
  local task_id="$1"
  local column_id="$2"
  node -e '
let raw = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => {
  raw += chunk;
});
process.stdin.on("end", () => {
  try {
    const data = JSON.parse(raw || "[]");
    if (!Array.isArray(data)) {
      process.exit(2);
      return;
    }
    const found = data.some(
      (item) => item && item.id === process.argv[1] && item.column_id === process.argv[2],
    );
    if (!found) {
      process.exit(3);
      return;
    }
  } catch {
    process.exit(1);
  }
});
' "${task_id}" "${column_id}"
}

echo "[1/7] Starting backend stack with Docker Compose..."
(cd "${ROOT_DIR}" && docker compose up -d --build >/dev/null 2>&1)

echo "[2/7] Waiting for backend health..."
wait_for_url "http://127.0.0.1:${BACKEND_PORT}/healthz" "backend healthcheck" 120

echo "[3/7] Starting frontend dev server..."
FRONTEND_LOG="$(mktemp -t kanban-frontend-smoke.XXXXXX.log)"
npm --prefix "${ROOT_DIR}/frontend" run dev -- --host "${FRONTEND_HOST}" --port "${FRONTEND_PORT}" >"${FRONTEND_LOG}" 2>&1 &
FRONTEND_PID=$!

echo "[4/7] Waiting for frontend readiness..."
for ((i = 1; i <= 120; i += 1)); do
  if curl -fsS "http://${FRONTEND_HOST}:${FRONTEND_PORT}" >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 "${FRONTEND_PID}" >/dev/null 2>&1; then
    echo "ERROR: frontend process exited before readiness." >&2
    cat "${FRONTEND_LOG}" >&2
    exit 1
  fi
  if [[ "${i}" -eq 120 ]]; then
    echo "ERROR: frontend did not become ready in time." >&2
    cat "${FRONTEND_LOG}" >&2
    exit 1
  fi
  sleep 1
done

echo "[5/7] Running auth + board proxy checks..."
EMAIL="smoke.$(date +%s).${RANDOM}@example.com"
REGISTER_PAYLOAD="$(printf '{"email":"%s","password":"pass123"}' "${EMAIL}")"
REGISTER_RESPONSE="$(curl -fsS -X POST "http://${FRONTEND_HOST}:${FRONTEND_PORT}/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "${REGISTER_PAYLOAD}")"
TOKEN="$(printf '%s' "${REGISTER_RESPONSE}" | json_field token)"

CREATE_BOARD_RESPONSE="$(curl -fsS -X POST "http://${FRONTEND_HOST}:${FRONTEND_PORT}/api/v1/boards" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"name":"Smoke Board"}')"
BOARD_ID="$(printf '%s' "${CREATE_BOARD_RESPONSE}" | json_field id)"

LIST_BOARD_RESPONSE="$(curl -fsS "http://${FRONTEND_HOST}:${FRONTEND_PORT}/api/v1/boards" \
  -H "Authorization: Bearer ${TOKEN}")"
printf '%s' "${LIST_BOARD_RESPONSE}" | assert_board_in_list "${BOARD_ID}"

echo "[6/7] Running columns + tasks proxy checks..."
CREATE_COLUMN_A="$(curl -fsS -X POST "http://${FRONTEND_HOST}:${FRONTEND_PORT}/api/v1/boards/${BOARD_ID}/columns" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"name":"Todo"}')"
COLUMN_A_ID="$(printf '%s' "${CREATE_COLUMN_A}" | json_field id)"

CREATE_COLUMN_B="$(curl -fsS -X POST "http://${FRONTEND_HOST}:${FRONTEND_PORT}/api/v1/boards/${BOARD_ID}/columns" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"name":"Done"}')"
COLUMN_B_ID="$(printf '%s' "${CREATE_COLUMN_B}" | json_field id)"

CREATE_TASK_RESPONSE="$(curl -fsS -X POST "http://${FRONTEND_HOST}:${FRONTEND_PORT}/api/v1/boards/${BOARD_ID}/columns/${COLUMN_A_ID}/tasks" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"title":"Smoke task","description":"frontend-proxy"}')"
TASK_ID="$(printf '%s' "${CREATE_TASK_RESPONSE}" | json_field id)"

MOVE_TASK_RESPONSE="$(curl -fsS -X PATCH "http://${FRONTEND_HOST}:${FRONTEND_PORT}/api/v1/boards/${BOARD_ID}/tasks/${TASK_ID}/move" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"column_id\":\"${COLUMN_B_ID}\"}")"
MOVED_COLUMN_ID="$(printf '%s' "${MOVE_TASK_RESPONSE}" | json_field column_id)"
if [[ "${MOVED_COLUMN_ID}" != "${COLUMN_B_ID}" ]]; then
  echo "ERROR: moved task column mismatch, expected ${COLUMN_B_ID}, got ${MOVED_COLUMN_ID}" >&2
  exit 1
fi

LIST_DONE_TASKS_RESPONSE="$(curl -fsS "http://${FRONTEND_HOST}:${FRONTEND_PORT}/api/v1/boards/${BOARD_ID}/columns/${COLUMN_B_ID}/tasks" \
  -H "Authorization: Bearer ${TOKEN}")"
printf '%s' "${LIST_DONE_TASKS_RESPONSE}" | assert_task_in_list "${TASK_ID}" "${COLUMN_B_ID}"

echo "[7/7] Smoke checks completed."
echo "FRONTEND_PROXY_SMOKE=PASS"
