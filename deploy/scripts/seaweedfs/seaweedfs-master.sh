#!/bin/sh
set -eu

exec weed master \
  -ip="${STORAGE_MASTER_NAME:-master}" \
  -ip.bind=0.0.0.0 \
  -metricsPort="${STORAGE_MASTER_METRICS_PORT:-9324}" \
  -volumeSizeLimitMB="${STORAGE_VOLUME_SIZE_LIMIT_MB:-30000}"