#!/bin/sh
set -eu

export MINIO_ROOT_USER="$(tr -d '\r\n' < "${STORAGE_S3_ACCESS_KEY_FILE}")"
export MINIO_ROOT_PASSWORD="$(tr -d '\r\n' < "${STORAGE_S3_SECRET_KEY_FILE}")"

exec minio server /data \
  --address ":9000" \
  --console-address ":${STORAGE_S3_CONSOLE_PORT:-9001}"