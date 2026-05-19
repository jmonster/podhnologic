#!/usr/bin/env bash

set -euo pipefail

P12_PATH="${PB_APPLE_CODESIGN_P12_PATH:-}"
KEYCHAIN_PATH="${PB_APPLE_CODESIGN_KEYCHAIN_PATH:-$HOME/Library/Keychains/podhnologic-ci.keychain-db}"
KEYCHAIN_PASSWORD="${PB_APPLE_CODESIGN_KEYCHAIN_PASSWORD:-${PB_APPLE_CODESIGN_P12_PASSWORD:-}}"

usage() {
	cat <<EOF
Usage: ${0##*/} setup|status

Environment:
  PB_APPLE_CODESIGN_P12_PASSWORD       Required for setup.
  PB_APPLE_CODESIGN_P12_PATH           Required path to Developer ID .p12.
  PB_APPLE_CODESIGN_KEYCHAIN_PATH      Default: ~/Library/Keychains/podhnologic-ci.keychain-db
  PB_APPLE_CODESIGN_KEYCHAIN_PASSWORD  Defaults to PB_APPLE_CODESIGN_P12_PASSWORD
EOF
}

die() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

list_keychains() {
	security list-keychains -d user | sed 's/[ "]*//g'
}

ensure_keychain_listed() {
	local existing=()
	local keychain

	while IFS= read -r keychain; do
		[[ -n "$keychain" ]] && existing+=("$keychain")
	done < <(list_keychains)

	for keychain in "${existing[@]}"; do
		[[ "$keychain" == "$KEYCHAIN_PATH" ]] && return
	done

	security list-keychains -d user -s "$KEYCHAIN_PATH" "${existing[@]}"
}

status() {
	printf 'keychain: %s\n' "$KEYCHAIN_PATH"
	if [[ ! -f "$KEYCHAIN_PATH" ]]; then
		die "keychain does not exist"
	fi

	security find-identity -v -p codesigning "$KEYCHAIN_PATH"
}

setup() {
	[[ -n "$P12_PATH" ]] || die "PB_APPLE_CODESIGN_P12_PATH is required"
	[[ -f "$P12_PATH" ]] || die "missing p12: $P12_PATH"
	[[ -n "${PB_APPLE_CODESIGN_P12_PASSWORD:-}" ]] || die "PB_APPLE_CODESIGN_P12_PASSWORD is required"
	[[ -n "$KEYCHAIN_PASSWORD" ]] || die "keychain password is required"

	if [[ ! -f "$KEYCHAIN_PATH" ]]; then
		security create-keychain -p "$KEYCHAIN_PASSWORD" "$KEYCHAIN_PATH"
	fi

	security set-keychain-settings -lut 21600 "$KEYCHAIN_PATH"
	security unlock-keychain -p "$KEYCHAIN_PASSWORD" "$KEYCHAIN_PATH"
	security import "$P12_PATH" \
		-P "$PB_APPLE_CODESIGN_P12_PASSWORD" \
		-A \
		-t cert \
		-f pkcs12 \
		-k "$KEYCHAIN_PATH"
	ensure_keychain_listed
	security set-key-partition-list \
		-S apple-tool:,apple:,codesign: \
		-s \
		-k "$KEYCHAIN_PASSWORD" \
		"$KEYCHAIN_PATH"
	status
}

case "${1:-}" in
	setup) setup ;;
	status) status ;;
	--help|-h) usage ;;
	*) usage >&2; exit 1 ;;
esac
