#!/bin/sh
set -eu

exec weed s3 \
  -filer="${SEAWEED_FILER_ADDRESS:-filer:8888}" \
  -ip.bind=0.0.0.0 \
  -metricsPort="${SEAWEED_S3_METRICS_PORT:-9327}"