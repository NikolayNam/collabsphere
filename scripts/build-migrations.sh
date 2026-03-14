#!/usr/bin/env bash
set -euo pipefail

mkdir -p /tmp/collabsphere-go-cache /tmp/collabsphere-go-modcache

GOCACHE=/tmp/collabsphere-go-cache \
GOMODCACHE=/tmp/collabsphere-go-modcache \
TMPDIR=/tmp \
  /usr/local/go/bin/go -C platform run ./internal/runtime/infrastructure/db/cmd/build-migrations
