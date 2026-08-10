#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [[ "${PODHNOLOGIC_APPSTORE_CONNECT_AUTH_READY:-0}" != "1" ]]; then
	exec "${repo_root}/scripts/with-appstore-connect-auth.sh" "$0" "$@"
fi

xcrun notarytool history \
	--key "${APPSTORE_CONNECT_API_KEY_PATH}" \
	--key-id "${APPSTORE_CONNECT_API_KEY_ID}" \
	--issuer "${APPSTORE_CONNECT_API_ISSUER_ID}" \
	--output-format json \
	--no-progress >/dev/null

printf 'appstore_connect_auth_preflight=ok source=p8_file\n'
