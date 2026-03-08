#!/bin/sh
set -eu

read_file_secret() {
  file_path="$1"
  tr -d '\r\n' < "$file_path"
}

if [ -n "${STORAGE_S3_ROOT_USER_FILE:-}" ]; then
  export MINIO_ROOT_USER="$(read_file_secret "${STORAGE_S3_ROOT_USER_FILE}")"
else
  export MINIO_ROOT_USER="${STORAGE_S3_ROOT_USER}"
fi
export MINIO_ROOT_PASSWORD="$(read_file_secret "${STORAGE_S3_ROOT_PASSWORD_FILE}")"
export MINIO_BROWSER_REDIRECT_URL="${STORAGE_S3_BROWSER_REDIRECT_URL}"

exec minio server /data --console-address ":${STORAGE_S3_CONSOLE_PORT}"
