#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [[ "${PODHNOLOGIC_APPSTORE_CONNECT_AUTH_READY:-0}" != "1" ]]; then
	exec "${repo_root}/scripts/with-appstore-connect-auth.sh" "$0" "$@"
fi

binary_path="${1:-}"

if [[ "$(uname -s)" != "Darwin" ]]; then
	printf 'macOS notarization requires Darwin, got %s\n' "$(uname -s)" >&2
	exit 1
fi

if [[ -z "${binary_path}" || ! -f "${binary_path}" ]]; then
	printf 'usage: %s <signed-binary>\n' "${0##*/}" >&2
	exit 2
fi

"${repo_root}/scripts/preflight-appstore-connect-auth.sh"
codesign --verify --strict --verbose=2 "${binary_path}"

notary_dir="$(mktemp -d "${TMPDIR:-/tmp}/podhnologic-notary.XXXXXX")"
trap 'rm -rf -- "${notary_dir}"' EXIT
zip_path="${notary_dir}/$(basename "${binary_path}").zip"
ditto -c -k --keepParent "${binary_path}" "${zip_path}"

xcrun notarytool submit "${zip_path}" \
	--key "${APPSTORE_CONNECT_API_KEY_PATH}" \
	--key-id "${APPSTORE_CONNECT_API_KEY_ID}" \
	--issuer "${APPSTORE_CONNECT_API_ISSUER_ID}" \
	--wait

spctl --assess --type install --verbose=4 "${binary_path}"
