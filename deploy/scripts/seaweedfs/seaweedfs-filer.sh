#!/bin/sh
set -eu

exec weed filer \
  -ip="${SEAWEED_FILER_NAME:-filer}" \
  -master="${SEAWEED_MASTER_ADDRESS:-master:9333}" \
  -ip.bind=0.0.0.0 \
  -metricsPort="${SEAWEED_FILER_METRICS_PORT:-9326}"