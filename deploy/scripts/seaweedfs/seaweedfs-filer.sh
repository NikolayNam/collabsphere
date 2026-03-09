#!/bin/sh
set -eu

exec weed filer \
  -ip="${STORAGE_FILER_NAME:-filer}" \
  -master="${STORAGE_MASTER_ADDRESS:-master:9333}" \
  -ip.bind=0.0.0.0 \
  -metricsPort="${STORAGE_FILER_METRICS_PORT:-9326}"