#!/bin/sh
set -eu

master_addr="${STORAGE_MASTER_ADDRESS:-master:9333}"
filer_addr="${STORAGE_FILER_ADDRESS:-filer:8888}"

set +e
output="$(weed shell -master="$master_addr" -filer="$filer_addr" 2>&1 <<'EOF'
s3.bucket.list
exit
EOF
)"
status=$?
set -e

if [ "$status" -ne 0 ]; then
  printf '%s\n' "$output" >&2
  exit 1
fi

if printf '%s\n' "$output" | grep -qi 'error:'; then
  printf '%s\n' "$output" >&2
  exit 1
fi
#!/bin/sh
set -eu

master_addr="${STORAGE_MASTER_ADDRESS:-master:9333}"
filer_addr="${STORAGE_FILER_ADDRESS:-filer:8888}"

set +e
output="$(weed shell -master="$master_addr" -filer="$filer_addr" 2>&1 <<'EOF'
s3.bucket.list
exit
EOF
)"
status=$?
set -e

if [ "$status" -ne 0 ]; then
  printf '%s\n' "$output" >&2
  exit 1
fi

if printf '%s\n' "$output" | grep -qi 'error:'; then
  printf '%s\n' "$output" >&2
  exit 1
fi
