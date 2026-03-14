#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"

curl -fsS "${BASE_URL}/health" >/dev/null
curl -fsS "${BASE_URL}/v1/health" >/dev/null
curl -fsS "${BASE_URL}/v1/ready" >/dev/null
curl -fsS "${BASE_URL}/v1/openapi.json" >/dev/null

echo "bootstrap smoke passed for ${BASE_URL}"
