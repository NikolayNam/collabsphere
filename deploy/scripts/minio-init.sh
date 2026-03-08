#!/bin/sh
set -eu

mkdir -p /run/generated/s3
password="$(tr -d '\r\n' < "${STORAGE_S3_ROOT_PASSWORD_FILE}")"

until mc alias set local "http://s3:${STORAGE_S3_PORT}" "${STORAGE_S3_ROOT_USER}" "$password" >/dev/null 2>&1; do
  sleep 2
done

mc mb --ignore-existing local/"${STORAGE_S3_BUCKET}"
mc anonymous set private local/"${STORAGE_S3_BUCKET}" || true
printf '%s' "${STORAGE_S3_ROOT_USER}" > /run/generated/s3/access_key
printf '%s' "$password" > /run/generated/s3/secret_key
