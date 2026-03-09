#!/bin/sh
set -eu

read_secret() {
  path="$1"
  name="$2"
  if [ -z "$path" ]; then
    echo "seaweed-init: ${name} path is empty" >&2
    exit 1
  fi
  if [ ! -e "$path" ]; then
    echo "seaweed-init: ${name} file does not exist: $path" >&2
    exit 1
  fi
  if [ -d "$path" ]; then
    echo "seaweed-init: ${name} path points to a directory, expected file: $path" >&2
    exit 1
  fi
  tr -d '\r\n' < "$path"
}

access_key="$(read_secret "${STORAGE_S3_ACCESS_KEY_FILE}" "storage s3 access key")"
secret_key="$(read_secret "${STORAGE_S3_SECRET_KEY_FILE}" "storage s3 secret key")"
alias_name="local"
endpoint="${STORAGE_S3_ENDPOINT}"

i=0
until mc alias set "$alias_name" "$endpoint" "$access_key" "$secret_key"; do
  i=$((i+1))
  echo "seaweed-init: waiting for S3 endpoint $endpoint, attempt=$i" >&2
  if [ "$i" -ge 30 ]; then
    echo "seaweed-init: S3 endpoint did not become ready in time" >&2
    exit 1
  fi
  sleep 2
done

mc mb --ignore-existing "$alias_name/${STORAGE_S3_BUCKET}"
echo "seaweed-init: bucket ensured: ${STORAGE_S3_BUCKET}"