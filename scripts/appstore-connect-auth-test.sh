#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
run_root="$(mktemp -d "${TMPDIR:-/tmp}/podhnologic-asc-auth.XXXXXX")"
trap 'rm -rf -- "${run_root}"' EXIT

key_id="ABCDEFGHIJ"
issuer_id="12345678-1234-1234-1234-1234567890ab"
key_path="${run_root}/AuthKey_${key_id}.p8"
openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out "${key_path}" >/dev/null 2>&1
chmod 600 "${key_path}"

APPSTORE_CONNECT_API_KEY_ID="${key_id}" \
APPSTORE_CONNECT_API_ISSUER_ID="${issuer_id}" \
APPSTORE_CONNECT_API_KEY_PATH="${key_path}" \
	"${repo_root}/scripts/with-appstore-connect-auth.sh" \
	/bin/bash -c '[[ "$PODHNOLOGIC_APPSTORE_CONNECT_AUTH_READY" == 1 && "$ASC_API_KEY_PATH" == "$APPSTORE_CONNECT_API_KEY_PATH" ]]'

expect_rejected() {
	local label="$1"
	shift
	set +e
	"$@" >"${run_root}/${label}.out" 2>"${run_root}/${label}.err"
	local status=$?
	set -e
	[[ "${status}" -eq 69 ]] || {
		printf 'expected %s to fail with 69; got %s\n' "${label}" "${status}" >&2
		exit 1
	}
}

chmod 644 "${key_path}"
expect_rejected permissive_mode \
	env APPSTORE_CONNECT_API_KEY_ID="${key_id}" \
		APPSTORE_CONNECT_API_ISSUER_ID="${issuer_id}" \
		APPSTORE_CONNECT_API_KEY_PATH="${key_path}" \
		"${repo_root}/scripts/with-appstore-connect-auth.sh" /usr/bin/true
chmod 600 "${key_path}"

expect_rejected legacy_keychain \
	env APPSTORE_CONNECT_API_KEY_ID="${key_id}" \
		APPSTORE_CONNECT_API_ISSUER_ID="${issuer_id}" \
		APPSTORE_CONNECT_API_KEY_PATH="${key_path}" \
		PB_APPLE_CODESIGN_KEYCHAIN_PATH="${run_root}/legacy.keychain-db" \
		"${repo_root}/scripts/with-appstore-connect-auth.sh" /usr/bin/true

expect_rejected inline_private_key \
	env APPSTORE_CONNECT_API_KEY_ID="${key_id}" \
		APPSTORE_CONNECT_API_ISSUER_ID="${issuer_id}" \
		APPSTORE_CONNECT_API_KEY_PATH="${key_path}" \
		APP_STORE_CONNECT_API_KEY_P8_BASE64="forbidden" \
		"${repo_root}/scripts/with-appstore-connect-auth.sh" /usr/bin/true

printf 'appstore-connect-auth-test.sh: PASS explicit_p8=accepted unsafe_inputs=rejected\n'
