#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [[ "${PODHNOLOGIC_APPSTORE_CONNECT_AUTH_READY:-0}" != "1" ]]; then
	exec "${repo_root}/scripts/with-appstore-connect-auth.sh" "$0" "$@"
fi

identity="${PODHNOLOGIC_CODESIGN_IDENTITY:-${APPLE_CODESIGN_IDENTITY:-Developer ID Application: Wabi Sabi Ware LLC (88M7JPMLS6)}}"

"${repo_root}/scripts/preflight-codesign-identity.sh" "${identity}"
"${repo_root}/scripts/preflight-appstore-connect-auth.sh"

printf 'apple_release_preflight=ok codesign_source=login_keychain_search appstore_connect_auth=p8_file\n'
