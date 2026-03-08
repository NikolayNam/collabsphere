#!/bin/sh
set -eu

raw_secret="$(tr -d '\r\n' < "${STORAGE_S3_GARAGE_RPC_SECRET_FILE}")"
if printf '%s' "$raw_secret" | grep -Eq '^[0-9a-fA-F]{64}$'; then
  rpc_secret="$(printf '%s' "$raw_secret" | tr 'A-F' 'a-f')"
else
  rpc_secret="$(printf '%s' "$raw_secret" | sha256sum | awk '{print $1}')"
fi

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
  echo "garage-init: already initialized"
  exit 0
fi

echo "garage-init: waiting for garage status"
until /garage -c /etc/garage/garage.toml status >/dev/null 2>&1; do
  sleep 2
done

status_output="$(/garage -c /etc/garage/garage.toml status)"
node_id="$(printf '%s\n' "$status_output" | awk '/^[0-9a-f][0-9a-f]+[[:space:]]/ { print $1; exit }')"
if [ -z "$node_id" ]; then
  echo "garage-init: unable to extract node id from garage status" >&2
  printf '%s\n' "$status_output" >&2
  exit 1
fi

echo "garage-init: assigning layout to node ${node_id}"
/garage -c /etc/garage/garage.toml layout assign -z "${STORAGE_S3_GARAGE_ZONE}" -c "${STORAGE_S3_GARAGE_CAPACITY}" "$node_id" || true
/garage -c /etc/garage/garage.toml layout apply --version 1 || true

echo "garage-init: creating bucket ${STORAGE_S3_BUCKET}"
/garage -c /etc/garage/garage.toml bucket create "${STORAGE_S3_BUCKET}" || true

echo "garage-init: creating key ${STORAGE_S3_GARAGE_KEY_NAME}"
key_output="$(/garage -c /etc/garage/garage.toml key create "${STORAGE_S3_GARAGE_KEY_NAME}" 2>/dev/null || true)"
if [ -z "$key_output" ]; then
  echo "garage-init: key create returned no credentials; manual key bootstrap may be required" >&2
  exit 1
fi

access_key="$(printf '%s\n' "$key_output" | awk -F': ' '/^Key ID:/ {print $2; exit}')"
secret_key="$(printf '%s\n' "$key_output" | awk -F': ' '/^Secret key:/ {print $2; exit}')"
if [ -z "$access_key" ] || [ -z "$secret_key" ]; then
  echo "garage-init: unable to parse created key credentials" >&2
  printf '%s\n' "$key_output" >&2
  exit 1
fi

echo "garage-init: allowing key on bucket"
/garage -c /etc/garage/garage.toml bucket allow --read --write --owner "${STORAGE_S3_BUCKET}" --key "${STORAGE_S3_GARAGE_KEY_NAME}" || true
printf '%s' "$access_key" > /run/generated/s3/access_key
printf '%s' "$secret_key" > /run/generated/s3/secret_key
touch /run/generated/s3/.garage_initialized

echo "garage-init: completed"
