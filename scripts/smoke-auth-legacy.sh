#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
ACCOUNT_EMAIL="${ACCOUNT_EMAIL:-smoke+$(date +%s)@example.com}"
ACCOUNT_PASSWORD="${ACCOUNT_PASSWORD:-Secret123}"

signup_payload=$(printf '{"email":"%s","password":"%s","displayName":"Smoke User"}' "$ACCOUNT_EMAIL" "$ACCOUNT_PASSWORD")
login_payload=$(printf '{"email":"%s","password":"%s"}' "$ACCOUNT_EMAIL" "$ACCOUNT_PASSWORD")

curl -fsS \
  -H 'Content-Type: application/json' \
  -d "$signup_payload" \
  "${BASE_URL}/v1/accounts" >/dev/null

login_response=$(
  curl -fsS \
    -H 'Content-Type: application/json' \
    -d "$login_payload" \
    "${BASE_URL}/v1/auth/login"
)

access_token=$(printf '%s' "$login_response" | perl -0ne 'print $1 if /"accessToken"\s*:\s*"([^"]+)"/')
refresh_token=$(printf '%s' "$login_response" | perl -0ne 'print $1 if /"refreshToken"\s*:\s*"([^"]+)"/')

if [[ -z "${access_token}" || -z "${refresh_token}" ]]; then
  echo "failed to extract access/refresh tokens from login response" >&2
  exit 1
fi

curl -fsS \
  -H "Authorization: Bearer ${access_token}" \
  "${BASE_URL}/v1/auth/me" >/dev/null

refresh_response=$(
  printf '{"refreshToken":"%s"}' "$refresh_token" | \
    curl -fsS \
      -H 'Content-Type: application/json' \
      -d @- \
      "${BASE_URL}/v1/auth/refresh"
)

rotated_refresh_token=$(printf '%s' "$refresh_response" | perl -0ne 'print $1 if /"refreshToken"\s*:\s*"([^"]+)"/')

if [[ -z "${rotated_refresh_token}" ]]; then
  echo "failed to extract rotated refresh token" >&2
  exit 1
fi

printf '{"refreshToken":"%s"}' "$rotated_refresh_token" | \
  curl -fsS \
    -X POST \
    -H 'Content-Type: application/json' \
    -d @- \
    "${BASE_URL}/v1/auth/logout" >/dev/null

echo "legacy auth smoke passed for ${ACCOUNT_EMAIL}"
