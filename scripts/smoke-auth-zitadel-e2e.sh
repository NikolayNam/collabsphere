#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://api.localhost:8080}"
ZITADEL_BASE_URL="${ZITADEL_BASE_URL:-http://auth.localhost:8090}"
CALLBACK_URL="${CALLBACK_URL:-${BASE_URL%/}/auth/callback}"
OUTPUT_DIR="${OUTPUT_DIR:-$(pwd)/output/playwright}"
SESSION_PREFIX="${SESSION_PREFIX:-zitadel-e2e}"
ADMIN_EMAIL="${ZITADEL_ADMIN_EMAIL:-admin@collabsphere.ru}"
ADMIN_PASSWORD="${ZITADEL_ADMIN_PASSWORD:-}"
ADMIN_TOKEN="${AUTH_ZITADEL_ADMIN_TOKEN:-}"
ADMIN_TOKEN_FILE="${AUTH_ZITADEL_ADMIN_TOKEN_FILE:-deploy/secrets/identity/zitadel_admin_token}"
EXPECTED_UNVERIFIED_ERROR="${EXPECTED_UNVERIFIED_ERROR:-Verified email is required for first external login}"

need_cmd() {
	local name="$1"
	if ! command -v "$name" >/dev/null 2>&1; then
		echo "required command is not available: $name" >&2
		exit 1
	fi
}

print_npx_help_and_exit() {
	cat >&2 <<'EOF'
npx is required for the ZITADEL browser smoke.

# Verify Node/npm are installed
node --version
npm --version

# If missing, install Node.js/npm, then:
npm install -g @playwright/cli@latest
playwright-cli --help
EOF
	exit 1
}

trim_file() {
	local path="$1"
	tr -d '\r\n' <"$path"
}

resolve_admin_password() {
	if [[ -n "${ADMIN_PASSWORD}" ]]; then
		printf '%s' "${ADMIN_PASSWORD}"
		return 0
	fi
	if [[ -f deploy/secrets/identity/zitadel_org_human_password ]]; then
		trim_file deploy/secrets/identity/zitadel_org_human_password
		return 0
	fi
	if [[ -f deploy/secrets/identity/zitadel_init_steps.yaml ]]; then
		perl -ne 'if (/^\s*Password:\s*"([^"]+)"/) { print $1; exit 0 }' deploy/secrets/identity/zitadel_init_steps.yaml
		return 0
	fi
	return 1
}

resolve_admin_token() {
	if [[ -n "${ADMIN_TOKEN}" ]]; then
		printf '%s' "${ADMIN_TOKEN}"
		return 0
	fi
	if [[ -n "${ADMIN_TOKEN_FILE}" && -f "${ADMIN_TOKEN_FILE}" ]]; then
		trim_file "${ADMIN_TOKEN_FILE}"
		return 0
	fi
	return 1
}

json_string() {
	perl -MJSON::PP -e 'print encode_json($ARGV[0])' -- "$1"
}

json_get() {
	local path="$1"
	perl -MJSON::PP -e '
		my $path = shift @ARGV;
		my $json = do { local $/; <STDIN> };
		my $value = decode_json($json);
		for my $part (split /\./, $path) {
			if (ref $value eq "HASH") {
				$value = $value->{$part};
				next;
			}
			if (ref $value eq "ARRAY" && $part =~ /^\d+$/) {
				$value = $value->[$part];
				next;
			}
			exit 2;
		}
		exit 3 if !defined $value;
		if (ref $value eq "ARRAY") {
			print join("\n", @$value);
			exit 0;
		}
		if (ref $value eq "HASH") {
			print encode_json($value);
			exit 0;
		}
		print $value;
	' "$path"
}

curl_resolve_args() {
	local host="$1"
	local port="$2"
	case "${host}" in
		api.localhost|auth.localhost)
			printf -- "--resolve %s:%s:127.0.0.1" "${host}" "${port}"
			;;
		*)
			printf ''
			;;
	esac
}

api_curl() {
	local resolve_arg
	resolve_arg="$(curl_resolve_args api.localhost 8080)"
	if [[ -n "${resolve_arg}" ]]; then
		# shellcheck disable=SC2086
		curl --silent --show-error --fail ${resolve_arg} "$@"
		return
	fi
	curl --silent --show-error --fail "$@"
}

zitadel_curl() {
	local resolve_arg
	resolve_arg="$(curl_resolve_args auth.localhost 8090)"
	if [[ -n "${resolve_arg}" ]]; then
		# shellcheck disable=SC2086
		curl --silent --show-error --fail ${resolve_arg} "$@"
		return
	fi
	curl --silent --show-error --fail "$@"
}

pwcli() {
	local session="$1"
	shift
	(
		cd "${OUTPUT_DIR}/${session}"
		"${PWCLI}" --session "${session}" "$@"
	)
}

capture_artifacts() {
	local session
	for session in "${ADMIN_SESSION}" "${UNVERIFIED_SESSION}" "${VERIFIED_SESSION}"; do
		if [[ -z "${session}" ]]; then
			continue
		fi
		(
			cd "${OUTPUT_DIR}/${session}"
			"${PWCLI}" --session "${session}" screenshot >/dev/null 2>&1 || true
			"${PWCLI}" --session "${session}" snapshot > snapshot.txt 2>/dev/null || true
		)
	done
}

cleanup() {
	local status="$1"
	if [[ "${status}" -ne 0 ]]; then
		capture_artifacts
		echo "ZITADEL E2E smoke failed; Playwright artifacts are in ${OUTPUT_DIR}" >&2
	fi
}

wait_for_api_ready() {
	local attempt
	for attempt in $(seq 1 40); do
		if api_curl "${BASE_URL%/}/v1/ready" >/dev/null 2>&1; then
			return 0
		fi
		sleep 2
	done
	echo "API readiness probe did not succeed at ${BASE_URL%/}/v1/ready" >&2
	return 1
}

check_oidc_login_preflight() {
	local headers
	headers="$(
		api_curl \
			--max-time 20 \
			--include \
			--output /dev/null \
			"${BASE_URL%/}/v1/auth/zitadel/login?return_to=/auth/callback"
	)" || {
		echo "backend OIDC login preflight failed; check AUTH_ZITADEL_* config and local OIDC app setup" >&2
		return 1
	}
	if ! printf '%s' "${headers}" | grep -qi '^location: '; then
		echo "backend OIDC login did not return redirect headers" >&2
		return 1
	fi
}

create_unverified_user() {
	local email="$1"
	local password="$2"
	local username="$3"
	local display_name="$4"
	local password_json
	local payload
	local response
	local user_id

	password_json="$(json_string "${password}")"
	payload="$(
		cat <<EOF
{
  "username": "${username}",
  "human": {
    "profile": {
      "givenName": "Smoke",
      "familyName": "User",
      "displayName": "${display_name}"
    },
    "email": {
      "email": "${email}",
      "isVerified": false
    },
    "password": {
      "password": ${password_json},
      "changeRequired": false
    }
  }
}
EOF
	)"

	if ! response="$(
		zitadel_curl \
			-H "Authorization: Bearer ${ZITADEL_ADMIN_PAT}" \
			-H 'Content-Type: application/json' \
			-d "${payload}" \
			"${ZITADEL_BASE_URL%/}/v2/users/new"
	)"; then
		payload="$(
			cat <<EOF
{
  "username": "${username}",
  "profile": {
    "givenName": "Smoke",
    "familyName": "User",
    "displayName": "${display_name}"
  },
  "email": {
    "email": "${email}",
    "isVerified": false
  },
  "password": {
    "password": ${password_json},
    "changeRequired": false
  }
}
EOF
		)"
		response="$(
			zitadel_curl \
				-H "Authorization: Bearer ${ZITADEL_ADMIN_PAT}" \
				-H 'Content-Type: application/json' \
				-d "${payload}" \
				"${ZITADEL_BASE_URL%/}/v2/users/human"
		)"
	fi

	user_id="$(printf '%s' "${response}" | json_get id 2>/dev/null || true)"
	if [[ -z "${user_id}" ]]; then
		user_id="$(printf '%s' "${response}" | json_get userId 2>/dev/null || true)"
	fi
	if [[ -z "${user_id}" ]]; then
		echo "failed to extract ZITADEL user id from create-user response" >&2
		echo "${response}" >&2
		return 1
	fi
	printf '%s' "${user_id}"
}

login_via_browser() {
	local session="$1"
	local email="$2"
	local password="$3"
	local expect_error="$4"
	local email_js
	local password_js
	local callback_js
	local expected_error_js

	email_js="$(json_string "${email}")"
	password_js="$(json_string "${password}")"
	callback_js="$(json_string "${CALLBACK_URL}")"
	expected_error_js="$(json_string "${EXPECTED_UNVERIFIED_ERROR}")"

	pwcli "${session}" close >/dev/null 2>&1 || true
	pwcli "${session}" open "${CALLBACK_URL}" >/dev/null
	pwcli "${session}" run-code "$(cat <<EOF
	await page.locator('#login-zitadel').click();
const email = ${email_js};
const password = ${password_js};
const callbackURL = ${callback_js};

async function submitCurrentStep() {
  const roleButtons = [
    /next/i,
    /continue/i,
    /sign in/i,
    /login/i,
    /anmelden/i,
    /weiter/i,
    /fortfahren/i,
    /allow/i,
    /accept/i
  ];
  for (const pattern of roleButtons) {
    const button = page.getByRole('button', { name: pattern }).first();
    if (await button.count()) {
      try {
        await button.click({ timeout: 1500 });
        return;
      } catch (_) {}
    }
  }
  const submit = page.locator('button[type=\"submit\"], input[type=\"submit\"]').first();
  if (await submit.count()) {
    try {
      await submit.click({ timeout: 1500 });
      return;
    } catch (_) {}
  }
  await page.keyboard.press('Enter');
}

const emailInput = page.locator('input[type=\"email\"], input[autocomplete=\"username\"], input[name=\"loginName\"], input[name=\"username\"], input[type=\"text\"]').first();
await emailInput.waitFor({ state: 'visible', timeout: 30000 });
await emailInput.fill(email);
await submitCurrentStep();

const passwordInput = page.locator('input[type=\"password\"], input[autocomplete=\"current-password\"]').first();
await passwordInput.waitFor({ state: 'visible', timeout: 30000 });
await passwordInput.fill(password);
await submitCurrentStep();

for (let attempt = 0; attempt < 45; attempt += 1) {
  if (page.url().startsWith(callbackURL)) {
    break;
  }
  const consent = page.getByRole('button', { name: /allow|accept|continue|weiter|anmelden/i }).first();
  if (await consent.count()) {
    try {
      await consent.click({ timeout: 1000 });
    } catch (_) {}
  }
  await page.waitForTimeout(1000);
}

if (!page.url().startsWith(callbackURL)) {
  throw new Error('browser did not return to callback page; current URL: ' + page.url());
}
EOF
)" >/dev/null

	if [[ "${expect_error}" == "yes" ]]; then
		pwcli "${session}" run-code "$(cat <<EOF
const expected = ${expected_error_js};
await page.waitForFunction(
  (message) => document.body.textContent && document.body.textContent.includes(message),
  expected,
  { timeout: 45000 }
);
EOF
)" >/dev/null
		return 0
	fi

	pwcli "${session}" run-code "$(cat <<'EOF'
await page.waitForFunction(() => {
  const access = document.getElementById('access-token');
  const refresh = document.getElementById('refresh-token');
  return Boolean(access && refresh && access.value.length > 0 && refresh.value.length > 0);
}, { timeout: 45000 });
EOF
)" >/dev/null
}

extract_token() {
	local session="$1"
	local field_id="$2"
	pwcli "${session}" eval "document.getElementById('${field_id}').value" | tr -d '\r'
}

read_status_text() {
	local session="$1"
	pwcli "${session}" eval "document.getElementById('status').textContent" | tr -d '\r'
}

need_cmd curl
need_cmd perl

if ! command -v npx >/dev/null 2>&1; then
	print_npx_help_and_exit
fi

export CODEX_HOME="${CODEX_HOME:-$HOME/.codex}"
PWCLI="${PWCLI:-$CODEX_HOME/skills/playwright/scripts/playwright_cli.sh}"
if [[ ! -x "${PWCLI}" ]]; then
	echo "Playwright wrapper not found: ${PWCLI}" >&2
	exit 1
fi

mkdir -p "${OUTPUT_DIR}"

ADMIN_PASSWORD="$(resolve_admin_password)" || {
	echo "could not resolve ZITADEL admin password; set ZITADEL_ADMIN_PASSWORD or provide deploy/secrets/identity/zitadel_org_human_password" >&2
	exit 1
}
ZITADEL_ADMIN_PAT="$(resolve_admin_token)" || {
	echo "could not resolve ZITADEL admin PAT; set AUTH_ZITADEL_ADMIN_TOKEN or provide ${ADMIN_TOKEN_FILE}" >&2
	exit 1
}

ADMIN_SESSION="${SESSION_PREFIX}-admin"
UNVERIFIED_SESSION="${SESSION_PREFIX}-unverified"
VERIFIED_SESSION="${SESSION_PREFIX}-verified"
for session in "${ADMIN_SESSION}" "${UNVERIFIED_SESSION}" "${VERIFIED_SESSION}"; do
	mkdir -p "${OUTPUT_DIR}/${session}"
done

trap 'status=$?; cleanup "$status"; exit "$status"' EXIT

wait_for_api_ready
check_oidc_login_preflight

suffix="$(date +%s)-$RANDOM"
UNVERIFIED_EMAIL="zitadel-e2e+${suffix}@example.com"
UNVERIFIED_PASSWORD="Smoke-${suffix}-Secret!9"
UNVERIFIED_USERNAME="smoke-${suffix}"
UNVERIFIED_DISPLAY_NAME="Smoke ${suffix}"

UNVERIFIED_USER_ID="$(create_unverified_user "${UNVERIFIED_EMAIL}" "${UNVERIFIED_PASSWORD}" "${UNVERIFIED_USERNAME}" "${UNVERIFIED_DISPLAY_NAME}")"

login_via_browser "${ADMIN_SESSION}" "${ADMIN_EMAIL}" "${ADMIN_PASSWORD}" "no"
ADMIN_ACCESS_TOKEN="$(extract_token "${ADMIN_SESSION}" "access-token")"
ADMIN_REFRESH_TOKEN="$(extract_token "${ADMIN_SESSION}" "refresh-token")"
if [[ -z "${ADMIN_ACCESS_TOKEN}" || -z "${ADMIN_REFRESH_TOKEN}" ]]; then
	echo "admin browser login did not yield backend access/refresh tokens" >&2
	exit 1
fi

platform_access_response="$(
	api_curl \
		-H "Authorization: Bearer ${ADMIN_ACCESS_TOKEN}" \
		"${BASE_URL%/}/v1/platform/access/me"
)"
if ! printf '%s' "${platform_access_response}" | json_get effectiveRoles | grep -qx 'platform_admin'; then
	echo "admin backend token does not resolve to effective platform_admin" >&2
	echo "${platform_access_response}" >&2
	exit 1
fi

login_via_browser "${UNVERIFIED_SESSION}" "${UNVERIFIED_EMAIL}" "${UNVERIFIED_PASSWORD}" "yes"
status_text="$(read_status_text "${UNVERIFIED_SESSION}")"
if [[ "${status_text}" != *"${EXPECTED_UNVERIFIED_ERROR}"* ]]; then
	echo "unverified browser login did not surface the expected rejection" >&2
	echo "${status_text}" >&2
	exit 1
fi

force_verify_response="$(
	api_curl \
		-X POST \
		-H "Authorization: Bearer ${ADMIN_ACCESS_TOKEN}" \
		"${BASE_URL%/}/v1/platform/users/${UNVERIFIED_USER_ID}/email/force-verify"
)"
verified_flag="$(printf '%s' "${force_verify_response}" | json_get verified 2>/dev/null || true)"
returned_user_id="$(printf '%s' "${force_verify_response}" | json_get userId 2>/dev/null || true)"
if [[ "${verified_flag}" != "true" || "${returned_user_id}" != "${UNVERIFIED_USER_ID}" ]]; then
	echo "force-verify response did not confirm the expected user" >&2
	echo "${force_verify_response}" >&2
	exit 1
fi

login_via_browser "${VERIFIED_SESSION}" "${UNVERIFIED_EMAIL}" "${UNVERIFIED_PASSWORD}" "no"
USER_ACCESS_TOKEN="$(extract_token "${VERIFIED_SESSION}" "access-token")"
USER_REFRESH_TOKEN="$(extract_token "${VERIFIED_SESSION}" "refresh-token")"
if [[ -z "${USER_ACCESS_TOKEN}" || -z "${USER_REFRESH_TOKEN}" ]]; then
	echo "verified browser login did not yield backend access/refresh tokens" >&2
	exit 1
fi

echo "ZITADEL browser E2E smoke passed for ${UNVERIFIED_EMAIL} (${UNVERIFIED_USER_ID})"
