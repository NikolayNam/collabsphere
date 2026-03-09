#!/bin/sh
set -eu

access_key_file="${STORAGE_S3_ACCESS_KEY_FILE:-/run/secrets/s3_access_key}"
secret_key_file="${STORAGE_S3_SECRET_KEY_FILE:-/run/secrets/s3_secret_key}"

if [ ! -f "$access_key_file" ]; then
  echo "seaweedfs-init: access key file not found: $access_key_file" >&2
  exit 1
fi

if [ ! -f "$secret_key_file" ]; then
  echo "seaweedfs-init: secret key file not found: $secret_key_file" >&2
  exit 1
fi

access_key="$(tr -d '\r\n' < "$access_key_file")"
secret_key="$(tr -d '\r\n' < "$secret_key_file")"
bucket="${STORAGE_S3_BUCKET:-collabsphere}"

until mc alias set local http://s3:8333 "$access_key" "$secret_key" >/dev/null 2>&1; do
  sleep 1
done

mc mb --ignore-existing "local/$bucket" >/dev/null