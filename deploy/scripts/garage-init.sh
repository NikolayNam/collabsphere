#!/bin/sh
set -eu

rpc_secret="$(tr -d '\r\n' < "${STORAGE_S3_GARAGE_RPC_SECRET_FILE}")"
mkdir -p /etc/garage /run/generated/s3
cat > /etc/garage/garage.toml <<EOF
metadata_dir = "/var/lib/garage/meta"
data_dir = "/var/lib/garage/data"
db_engine = "sqlite"
replication_factor = 1

rpc_bind_addr = "0.0.0.0:${STORAGE_S3_GARAGE_RPC_PORT}"
rpc_public_addr = "s3:${STORAGE_S3_GARAGE_RPC_PORT}"
rpc_secret = "$rpc_secret"

[s3_api]
api_bind_addr = "0.0.0.0:${STORAGE_S3_PORT}"
s3_region = "${STORAGE_S3_REGION}"
root_domain = ""
EOF

if [ -f /run/generated/s3/.garage_initialized ]; then
  exit 0
fi

until /garage -c /etc/garage/garage.toml status >/dev/null 2>&1; do
  sleep 2
done

node_id="$(/garage -c /etc/garage/garage.toml node id | head -n 1 | tr -d '\r\n' | xargs)"
[ -n "$node_id" ]

/garage -c /etc/garage/garage.toml layout assign -z "${STORAGE_S3_GARAGE_ZONE}" -c "${STORAGE_S3_GARAGE_CAPACITY}" "$node_id" || true
/garage -c /etc/garage/garage.toml layout apply --version 1 || true
/garage -c /etc/garage/garage.toml bucket create "${STORAGE_S3_BUCKET}" || true

key_output="$(/garage -c /etc/garage/garage.toml key create "${STORAGE_S3_GARAGE_KEY_NAME}" 2>/dev/null || true)"
if [ -z "$key_output" ]; then
  echo "garage key create did not return credentials; manual key bootstrap is required" >&2
  exit 1
fi

access_key="$(printf '%s\n' "$key_output" | awk '/Key ID:/ {print $3; exit}')"
secret_key="$(printf '%s\n' "$key_output" | awk '/Secret key:/ {print $3; exit}')"
[ -n "$access_key" ]
[ -n "$secret_key" ]

/garage -c /etc/garage/garage.toml bucket allow --read --write --owner "${STORAGE_S3_BUCKET}" --key "${STORAGE_S3_GARAGE_KEY_NAME}" || true
printf '%s' "$access_key" > /run/generated/s3/access_key
printf '%s' "$secret_key" > /run/generated/s3/secret_key
touch /run/generated/s3/.garage_initialized
