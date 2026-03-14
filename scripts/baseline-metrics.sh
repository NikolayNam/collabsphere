#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
METRICS_URL="${METRICS_URL:-${API_BASE_URL}/metrics}"
OUT_FILE="${OUT_FILE:-./logs/baseline-metrics.prom}"

mkdir -p "$(dirname "${OUT_FILE}")"

echo "[baseline] probing health endpoint..."
curl -fsS "${API_BASE_URL}/health" >/dev/null

echo "[baseline] collecting metrics snapshot from ${METRICS_URL}"
curl -fsS "${METRICS_URL}" >"${OUT_FILE}"

echo "[baseline] snapshot written to ${OUT_FILE}"
echo "[baseline] key series preview:"
grep -E '^(collabsphere_http_requests_total|collabsphere_http_request_duration_seconds_bucket|go_memstats_alloc_bytes)' "${OUT_FILE}" | head -n 15 || true
