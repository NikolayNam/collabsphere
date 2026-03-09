#!/bin/sh
set -eu

exec weed volume \
  -ip="${STORAGE_VOLUME_NAME:-volume}" \
  -master="${STORAGE_MASTER_ADDRESS:-master:9333}" \
  -ip.bind=0.0.0.0 \
  -port="${STORAGE_VOLUME_INTERNAL_PORT:-8080}" \
  -metricsPort="${STORAGE_VOLUME_METRICS_PORT:-9325}" \
  -dir="${STORAGE_VOLUME_DIR:-/data}"