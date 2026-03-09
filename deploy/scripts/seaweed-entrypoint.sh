#!/bin/sh
set -eu

read_secret() {
  path="$1"
  name="$2"
  if [ -z "$path" ]; then
    echo "seaweed: ${name} path is empty" >&2
    exit 1
  fi
  if [ ! -e "$path" ]; then
    echo "seaweed: ${name} file does not exist: $path" >&2
    exit 1
  fi
  if [ -d "$path" ]; then
    echo "seaweed: ${name} path points to a directory, expected file: $path" >&2
    exit 1
  fi
  tr -d '\r\n' < "$path"
}

export AWS_ACCESS_KEY_ID="$(read_secret "${STORAGE_S3_ACCESS_KEY_FILE}" "storage s3 access key")"
export AWS_SECRET_ACCESS_KEY="$(read_secret "${STORAGE_S3_SECRET_KEY_FILE}" "storage s3 secret key")"

exec weed server \
  -dir=/data \
  -s3 \
  -s3.port=8333