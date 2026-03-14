#!/bin/sh
set -eu

access_key_file="${STORAGE_S3_ACCESS_KEY_FILE:-/run/secrets/s3_access_key}"
secret_key_file="${STORAGE_S3_SECRET_KEY_FILE:-/run/secrets/s3_secret_key}"

if [ ! -f "$access_key_file" ]; then
  echo "seaweedfs-s3: access key file not found: $access_key_file" >&2
  exit 1
fi

if [ ! -f "$secret_key_file" ]; then
  echo "seaweedfs-s3: secret key file not found: $secret_key_file" >&2
  exit 1
fi

access_key="$access_key_file"
secret_key="$secret_key_file"
config_file="/tmp/seaweedfs-s3.json"
master_addr="${STORAGE_MASTER_ADDRESS:-master:9333}"
filer_addr="${STORAGE_FILER_ADDRESS:-filer:8888}"
max_attempts="${STORAGE_S3_WAIT_MAX_ATTEMPTS:-30}"
retry_delay_seconds="${STORAGE_S3_WAIT_RETRY_DELAY_SECONDS:-2}"

wait_for_storage_ready() {
  attempt=1
  while [ "$attempt" -le "$max_attempts" ]; do
    set +e
    output="$(weed shell -master="$master_addr" -filer="$filer_addr" 2>&1 <<'EOF'
s3.bucket.list
exit
EOF
)"
    status=$?
    set -e

    if [ "$status" -eq 0 ] && ! printf '%s\n' "$output" | grep -qi 'error:'; then
      return 0
    fi

    if [ "$attempt" -eq "$max_attempts" ]; then
      printf '%s\n' "$output" >&2
      echo "seaweedfs-s3: storage endpoints are not ready after $max_attempts attempts (master=$master_addr filer=$filer_addr)" >&2
      return 1
    fi

    attempt=$((attempt + 1))
    sleep "$retry_delay_seconds"
  done
  return 1
}

cat >"$config_file" <<EOF
{
  "identities": [
    {
      "name": "collabsphere-app",
      "credentials": [
        {
          "accessKey": "$access_key",
          "secretKey": "$secret_key"
        }
      ],
      "actions": ["Admin", "Read", "Write", "List", "Tagging"]
    }
  ]
}
EOF

wait_for_storage_ready

exec weed s3 \
  -filer="$filer_addr" \
  -ip.bind=0.0.0.0 \
  -metricsPort="${STORAGE_S3_METRICS_PORT:-9327}" \
  -config="$config_file"