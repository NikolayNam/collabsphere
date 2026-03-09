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

access_key="$(tr -d '\r\n' < "$access_key_file")"
secret_key="$(tr -d '\r\n' < "$secret_key_file")"
config_file="/tmp/seaweedfs-s3.json"

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

exec weed s3 \
  -filer="${STORAGE_FILER_ADDRESS:-filer:8888}" \
  -ip.bind=0.0.0.0 \
  -metricsPort="${STORAGE_S3_METRICS_PORT:-9327}" \
  -config="$config_file"