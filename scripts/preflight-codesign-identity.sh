#!/usr/bin/env bash

set -euo pipefail

identity="${1:-}"
[[ -n "${identity}" ]] || {
	printf 'usage: %s <exact-identity-label>\n' "${0##*/}" >&2
	exit 64
}

identities="$(security find-identity -v -p codesigning 2>/dev/null || true)"
if ! grep -Fq "\"${identity}\"" <<<"${identities}"; then
	printf 'error: signing identity is unavailable through the logged-in user normal keychain search: %s\n' "${identity}" >&2
	printf 'install it once in the login keychain; release scripts never import or unlock it\n' >&2
	exit 1
fi

probe_dir="$(mktemp -d "${TMPDIR:-/tmp}/podhnologic-signing-preflight.XXXXXX")"
trap 'rm -rf -- "${probe_dir}"' EXIT
probe="${probe_dir}/probe"
printf '#!/bin/sh\nexit 0\n' >"${probe}"
chmod +x "${probe}"
codesign --force --sign "${identity}" --timestamp=none "${probe}" >/dev/null
codesign --verify --strict "${probe}"

printf 'codesign_preflight=ok source=login_keychain_search identity=%s\n' "${identity}"
