#!/bin/sh
set -eu

if [ -z "${TEST_POSTGRES_USER:-}" ] || [ -z "${TEST_POSTGRES_DB:-}" ]; then
  echo "TEST_POSTGRES_USER and TEST_POSTGRES_DB must be set"
  exit 1
fi

pg_isready -U "${TEST_POSTGRES_USER}" -d "${TEST_POSTGRES_DB}"
