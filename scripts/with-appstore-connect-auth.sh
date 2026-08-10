#!/usr/bin/env bash

set -euo pipefail

if [[ "$#" -eq 0 ]]; then
	printf 'usage: %s <command> [args...]\n' "${0##*/}" >&2
	exit 64
fi

fail() {
	printf 'appstore_connect_auth_error=%s\n' "$1" >&2
	exit 69
}

for forbidden_name in \
	APP_STORE_CONNECT_API_KEY_P8_BASE64 \
	APPSTORE_CONNECT_API_PRIVATE_KEY_BASE64 \
	APP_STORE_CONNECT_API_KEY_KEY \
	APP_STORE_CONNECT_API_KEY_KEY_CONTENT \
	APP_STORE_CONNECT_API_KEY_IS_KEY_CONTENT_BASE64 \
	ASC_API_KEY_CONTENT \
	ASC_API_KEY_CONTENT_BASE64 \
	PODHNOLOGIC_NOTARY_PROFILE \
	NOTARY_PROFILE \
	PB_APPLE_CODESIGN_P12_PATH \
	PB_APPLE_CODESIGN_P12_PASSWORD \
	PB_APPLE_CODESIGN_KEYCHAIN_PATH \
	PB_APPLE_CODESIGN_KEYCHAIN_PASSWORD; do
	[[ -z "${!forbidden_name:-}" ]] || fail "legacy_inline_or_keychain_configuration_forbidden name=${forbidden_name}"
done

key_id="${APPSTORE_CONNECT_API_KEY_ID:-${PODHNOLOGIC_ASC_API_KEY:-${ASC_API_KEY:-${APP_STORE_CONNECT_API_KEY_KEY_ID:-}}}}"
issuer_id="${APPSTORE_CONNECT_API_ISSUER_ID:-${PODHNOLOGIC_ASC_API_ISSUER:-${ASC_API_ISSUER:-${APP_STORE_CONNECT_API_KEY_ISSUER_ID:-}}}}"
key_path="${APPSTORE_CONNECT_API_KEY_PATH:-${PODHNOLOGIC_ASC_API_KEY_PATH:-${ASC_API_KEY_PATH:-${APP_STORE_CONNECT_API_KEY_KEY_FILEPATH:-}}}}"

[[ "${key_id}" =~ ^[A-Z0-9]{10}$ ]] || fail invalid_or_missing_key_id
[[ "${issuer_id}" =~ ^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$ ]] || fail invalid_or_missing_issuer_id
[[ "${key_path}" == /* ]] || fail key_path_must_be_absolute
[[ "$(basename "${key_path}")" == "AuthKey_${key_id}.p8" ]] || fail key_filename_must_match_key_id
[[ -f "${key_path}" && ! -L "${key_path}" ]] || fail key_path_must_be_an_existing_non_symlink_file
key_path="$(cd "$(dirname "${key_path}")" && pwd -P)/$(basename "${key_path}")"
[[ -O "${key_path}" ]] || fail key_file_must_be_owned_by_current_user
if stat -f '%Lp' "${key_path}" >/dev/null 2>&1; then
	key_mode="$(stat -f '%Lp' "${key_path}")"
else
	key_mode="$(stat -c '%a' "${key_path}")"
fi
[[ "${key_mode}" =~ ^[46]00$ ]] || fail "key_file_permissions_must_be_0400_or_0600 actual=${key_mode}"
openssl pkey -in "${key_path}" -noout >/dev/null 2>&1 || fail key_file_is_not_a_valid_private_key

export APPSTORE_CONNECT_API_KEY_ID="${key_id}"
export APPSTORE_CONNECT_API_ISSUER_ID="${issuer_id}"
export APPSTORE_CONNECT_API_KEY_PATH="${key_path}"
export PODHNOLOGIC_ASC_API_KEY="${key_id}"
export PODHNOLOGIC_ASC_API_ISSUER="${issuer_id}"
export PODHNOLOGIC_ASC_API_KEY_PATH="${key_path}"
export ASC_API_KEY="${key_id}"
export ASC_API_ISSUER="${issuer_id}"
export ASC_API_KEY_PATH="${key_path}"
export APP_STORE_CONNECT_API_KEY_KEY_ID="${key_id}"
export APP_STORE_CONNECT_API_KEY_ISSUER_ID="${issuer_id}"
export APP_STORE_CONNECT_API_KEY_KEY_FILEPATH="${key_path}"
export PODHNOLOGIC_APPSTORE_CONNECT_AUTH_READY=1

exec "$@"
