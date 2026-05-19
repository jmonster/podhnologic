#!/usr/bin/env bash

set -euo pipefail

binary_path="${1:-}"
profile="${PODHNOLOGIC_NOTARY_PROFILE:-${NOTARY_PROFILE:-}}"
api_key="${PODHNOLOGIC_ASC_API_KEY:-${ASC_API_KEY:-${APP_STORE_CONNECT_API_KEY_KEY_ID:-}}}"
api_issuer="${PODHNOLOGIC_ASC_API_ISSUER:-${ASC_API_ISSUER:-${APP_STORE_CONNECT_API_KEY_ISSUER_ID:-}}}"
api_key_path="${PODHNOLOGIC_ASC_API_KEY_PATH:-${ASC_API_KEY_PATH:-${APP_STORE_CONNECT_API_KEY_KEY_FILEPATH:-}}}"

if [[ "$(uname -s)" != "Darwin" ]]; then
	printf 'macOS notarization requires Darwin, got %s\n' "$(uname -s)" >&2
	exit 1
fi

if [[ -z "$binary_path" || ! -f "$binary_path" ]]; then
	printf 'usage: %s <signed-binary>\n' "${0##*/}" >&2
	exit 2
fi

codesign --verify --strict --verbose=2 "$binary_path"

zip_path="${binary_path}.notary.zip"
rm -f "$zip_path"
ditto -c -k --keepParent "$binary_path" "$zip_path"

if [[ -n "$profile" ]]; then
	xcrun notarytool submit "$zip_path" --keychain-profile "$profile" --wait
elif [[ -n "$api_key" && -n "$api_key_path" ]]; then
	api_args=(
		--key "$api_key_path" \
		--key-id "$api_key"
	)
	if [[ -n "$api_issuer" ]]; then
		api_args+=(--issuer "$api_issuer")
	fi
	xcrun notarytool submit "$zip_path" "${api_args[@]}" --wait
else
	: "${APPLE_ID:?set APPLE_ID or PODHNOLOGIC_NOTARY_PROFILE}"
	: "${APPLE_TEAM_ID:?set APPLE_TEAM_ID or PODHNOLOGIC_NOTARY_PROFILE}"
	: "${APPLE_APP_SPECIFIC_PASSWORD:?set APPLE_APP_SPECIFIC_PASSWORD or PODHNOLOGIC_NOTARY_PROFILE}"
	xcrun notarytool submit "$zip_path" \
		--apple-id "$APPLE_ID" \
		--team-id "$APPLE_TEAM_ID" \
		--password "$APPLE_APP_SPECIFIC_PASSWORD" \
		--wait
fi

rm -f "$zip_path"
spctl --assess --type install --verbose=4 "$binary_path"
