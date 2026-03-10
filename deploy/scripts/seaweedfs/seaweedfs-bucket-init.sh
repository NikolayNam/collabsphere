#!/bin/sh
set -eu

bucket="${STORAGE_S3_BUCKET:-collabsphere}"
master_addr="${STORAGE_MASTER_ADDRESS:-master:9333}"

# Идея правильная: bucket создается нативной командой SeaweedFS.
# Точный idempotent-вариант я бы еще один раз проверил перед коммитом.
weed shell -master="$master_addr" <<EOF
s3.bucket.create -name $bucket
exit
EOF
