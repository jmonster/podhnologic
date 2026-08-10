# Release

## Required Apple release inputs

Install the Developer ID certificate and matching private key once in the
logged-in user's normal login keychain. The expected identity is:

```text
Developer ID Application: Wabi Sabi Ware LLC (88M7JPMLS6)
```

Keep the App Store Connect API private key outside the repository as an
owner-only file. Export only its identity and path:

```sh
export APPSTORE_CONNECT_API_KEY_ID='YOUR_KEY_ID'
export APPSTORE_CONNECT_API_ISSUER_ID='YOUR_ISSUER_UUID'
export APPSTORE_CONNECT_API_KEY_PATH='/absolute/path/AuthKey_YOUR_KEY_ID.p8'
```

The `.p8` file must be owned by the current user and have mode `0400` or `0600`.
Do not put the private-key bytes in base64, a shell variable, a repository, a
temporary decoded file, or a Keychain generic-password item. Release scripts do
not create, unlock, import into, reconfigure, or edit ACLs on any keychain.

## Preflight before building

Run the full Apple preflight before spending time on a release build:

```sh
make release-preflight
```

This signs and verifies a disposable probe through the normal user keychain
search and performs a read-only `notarytool history` request with the `.p8`. A
missing identity or rejected API key is a hard stop.

## Build, sign, and notarize

```sh
./scripts/build-linked.sh darwin-arm64
./scripts/sign-macos.sh build/podhnologic-darwin-arm64
./scripts/notarize-macos.sh build/podhnologic-darwin-arm64
```

The signing script uses the hardened runtime and identifier
`com.wabisabiware.podhnologic`. The notarization script accepts only the explicit
file-backed App Store Connect API key; it does not use a stored `notarytool`
keychain profile or Apple ID password fallback.

## Tag a release

```sh
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

The installer downloads the matching release asset and `.sha256` file before
installing.

## Notarization logs

```sh
xcrun notarytool log SUBMISSION_ID \
  --key "$APPSTORE_CONNECT_API_KEY_PATH" \
  --key-id "$APPSTORE_CONNECT_API_KEY_ID" \
  --issuer "$APPSTORE_CONNECT_API_ISSUER_ID"
```

## Licensing check

Before publishing release assets, confirm the FFmpeg build is still
LGPL-configured:

```sh
rg -n --glob '!cache/**' --glob '!work/**' --glob '!dist/**' -- \
  '--enable-gpl|--enable-nonfree' scripts/ffmpeg
```

Rebuilt FFmpeg config headers should report
`FFMPEG_LICENSE "LGPL version 2.1 or later"`.
