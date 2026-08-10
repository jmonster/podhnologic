#!/usr/bin/env bash

set -euo pipefail

target_path="${1:-}"
team_id="${PODHNOLOGIC_APPLE_TEAM_ID:-88M7JPMLS6}"
identifier="${PODHNOLOGIC_BUNDLE_ID:-com.wabisabiware.podhnologic}"
identity="${PODHNOLOGIC_CODESIGN_IDENTITY:-${APPLE_CODESIGN_IDENTITY:-Developer ID Application: Wabi Sabi Ware LLC (88M7JPMLS6)}}"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ "$(uname -s)" != "Darwin" ]]; then
	printf 'macOS signing requires Darwin, got %s\n' "$(uname -s)" >&2
	exit 1
fi

if [[ -z "$target_path" || ! -f "$target_path" ]]; then
	printf 'usage: %s <binary>\n' "${0##*/}" >&2
	exit 2
fi

"${repo_root}/scripts/preflight-codesign-identity.sh" "${identity}"

printf 'signing %s\n' "$target_path"
printf 'identity: %s\n' "$identity"
printf 'identifier: %s\n' "$identifier"

codesign \
	--force \
	--options runtime \
	--timestamp \
	--sign "$identity" \
	--identifier "$identifier" \
	"$target_path"

codesign --verify --strict --verbose=2 "$target_path"

signature_info="$(codesign -dv --verbose=4 "$target_path" 2>&1)"
actual_identifier="$(printf '%s\n' "$signature_info" | awk -F= '/^Identifier=/{print $2; exit}')"
actual_team="$(printf '%s\n' "$signature_info" | awk -F= '/^TeamIdentifier=/{print $2; exit}')"
signature="$(printf '%s\n' "$signature_info" | awk -F= '/^Signature=/{print $2; exit}')"

printf 'signed: id=%s team=%s signature=%s\n' "$actual_identifier" "$actual_team" "$signature"

if [[ "$actual_identifier" != "$identifier" || "$actual_team" != "$team_id" ]]; then
	printf 'signature mismatch\n' >&2
	printf 'expected: id=%s team=%s\n' "$identifier" "$team_id" >&2
	printf 'actual:   id=%s team=%s\n' "$actual_identifier" "$actual_team" >&2
	exit 1
fi
