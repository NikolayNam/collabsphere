#!/bin/sh
set -eu

exec weed master \
  -ip="${SEAWEED_MASTER_NAME:-master}" \
  -ip.bind=0.0.0.0 \
  -metricsPort="${SEAWEED_MASTER_METRICS_PORT:-9324}" \
  -volumeSizeLimitMB="${SEAWEED_VOLUME_SIZE_LIMIT_MB:-30000}"