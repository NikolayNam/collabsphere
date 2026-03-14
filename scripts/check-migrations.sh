#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="${ROOT_DIR}/platform/internal/runtime/infrastructure/db/migrations"
TMP_DIR="$(mktemp -d)"
SNAPSHOT_DIR="${TMP_DIR}/migrations-before"

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

mkdir -p "${SNAPSHOT_DIR}"
cp -R "${MIGRATIONS_DIR}/." "${SNAPSHOT_DIR}/"

bash "${ROOT_DIR}/scripts/build-migrations.sh"

if ! diff -ru "${SNAPSHOT_DIR}" "${MIGRATIONS_DIR}" >/tmp/collabsphere-migrations.diff 2>&1; then
  cat /tmp/collabsphere-migrations.diff
  echo "bundled migrations are out of date; run 'make migrations-build' and commit updated files" >&2
  exit 1
fi
