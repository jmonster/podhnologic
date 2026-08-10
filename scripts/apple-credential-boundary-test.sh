#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

dangerous="$(rg -n \
	--glob '*.sh' \
	--glob '!**/*test.sh' \
	'security[[:space:]]+(create-keychain|delete-keychain|unlock-keychain|lock-keychain|set-keychain-settings|set-key-partition-list|import|add-generic-password|delete-generic-password|list-keychains[[:space:]].*-s|default-keychain[[:space:]].*-s)|allowProvisioningUpdates|allowProvisioningDeviceRegistration|notarytool[[:space:]]+store-credentials|--keychain-profile' \
	ci scripts || true)"
if [[ -n "${dangerous}" ]]; then
	printf '%s\n' "${dangerous}" >&2
	exit 1
fi

[[ ! -e ci/apple_codesign_setup.sh ]] || {
	printf 'retired keychain setup remains reachable\n' >&2
	exit 1
}

inline_consumers="$(rg -n \
	--glob '*.sh' \
	--glob '!**/*test.sh' \
	--glob '!scripts/with-appstore-connect-auth.sh' \
	'APP_STORE_CONNECT_API_KEY_P8_BASE64|APPSTORE_CONNECT_API_PRIVATE_KEY_BASE64|APP_STORE_CONNECT_API_KEY_KEY_CONTENT|ASC_API_KEY_CONTENT|PB_APPLE_CODESIGN_P12|PB_APPLE_CODESIGN_KEYCHAIN' \
	ci scripts || true)"
if [[ -n "${inline_consumers}" ]]; then
	printf '%s\n' "${inline_consumers}" >&2
	exit 1
fi

printf 'apple-credential-boundary-test.sh: PASS keychain_mutators=0 appstore_connect=p8_file codesign=login_keychain_search\n'
