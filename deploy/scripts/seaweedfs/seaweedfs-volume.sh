#!/bin/sh
set -eu

exec weed volume \
  -ip="${SEAWEED_VOLUME_NAME:-volume}" \
  -master="${SEAWEED_MASTER_ADDRESS:-master:9333}" \
  -ip.bind=0.0.0.0 \
  -port="${SEAWEED_VOLUME_INTERNAL_PORT:-8080}" \
  -metricsPort="${SEAWEED_VOLUME_METRICS_PORT:-9325}" \
  -dir="${SEAWEED_VOLUME_DIR:-/data}"