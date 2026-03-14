#!/bin/sh
set -eu

bucket="${STORAGE_S3_BUCKET:-collabsphere}"
master_addr="${STORAGE_MASTER_ADDRESS:-master:9333}"
filer_addr="${STORAGE_FILER_ADDRESS:-filer:8888}"
max_attempts="${STORAGE_S3_BUCKET_INIT_MAX_ATTEMPTS:-20}"
retry_delay_seconds="${STORAGE_S3_BUCKET_INIT_RETRY_DELAY_SECONDS:-2}"
attempt=1

while [ "$attempt" -le "$max_attempts" ]; do
  set +e
  output="$(weed shell -master="$master_addr" -filer="$filer_addr" 2>&1 <<EOF
s3.bucket.create -name $bucket
exit
EOF
)"
  status=$?
  set -e

  case "$output" in
    *"already exists"*)
      printf '%s\n' "$output"
      exit 0
      ;;
  esac

  if [ "$status" -eq 0 ] && ! printf '%s\n' "$output" | grep -qi 'error:'; then
    printf '%s\n' "$output"
    exit 0
  fi

  printf '%s\n' "$output" >&2

  if [ "$attempt" -eq "$max_attempts" ]; then
    echo "seaweedfs-bucket-init: failed to create bucket '$bucket' after $max_attempts attempts" >&2
    exit 1
  fi

  attempt=$((attempt + 1))
  sleep "$retry_delay_seconds"
done
