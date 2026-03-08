#!/bin/sh
set -eu

rpc_secret="$(tr -d '\r\n' < "${STORAGE_S3_GARAGE_RPC_SECRET_FILE}")"
mkdir -p /etc/garage
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

exec garage -c /etc/garage/garage.toml server
