#!/bin/sh
set -eu

mkdir -p /run/generated/s3
password="$(tr -d '\r\n' < "${STORAGE_S3_ROOT_PASSWORD_FILE}")"
if [ -n "${STORAGE_S3_ROOT_USER_FILE:-}" ]; then
  root_user="$(tr -d '\r\n' < "${STORAGE_S3_ROOT_USER_FILE}")"
else
  root_user="${STORAGE_S3_ROOT_USER}"
fi

until mc alias set local "http://s3:${STORAGE_S3_PORT}" "$root_user" "$password" >/dev/null 2>&1; do
  sleep 2
done

mc mb --ignore-existing local/"${STORAGE_S3_BUCKET}"
mc anonymous set private local/"${STORAGE_S3_BUCKET}" || true
printf '%s' "$root_user" > /run/generated/s3/access_key
printf '%s' "$password" > /run/generated/s3/secret_key
