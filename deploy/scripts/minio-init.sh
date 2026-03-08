#!/bin/sh
set -eu

access_key="$(tr -d '\r\n' < "${STORAGE_S3_ACCESS_KEY_FILE}")"
secret_key="$(tr -d '\r\n' < "${STORAGE_S3_SECRET_KEY_FILE}")"
alias_name="local"
endpoint="${STORAGE_S3_ENDPOINT}"

until mc alias set "$alias_name" "$endpoint" "$access_key" "$secret_key" >/dev/null 2>&1; do
  sleep 2
done

mc mb --ignore-existing "$alias_name/${STORAGE_S3_BUCKET}"
mc anonymous set private "$alias_name/${STORAGE_S3_BUCKET}" >/dev/null 2>&1 || true