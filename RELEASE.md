# Release

Tagged releases are built by `.github/workflows/release.yml`.

## Required Secrets

- `APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_BASE64`
- `APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_PASSWORD`
- `APPLE_CODESIGN_KEYCHAIN_PASSWORD`
- `APPLE_CODESIGN_IDENTITY`
- `APPLE_ID`
- `APPLE_TEAM_ID`
- `APPLE_APP_SPECIFIC_PASSWORD`

Instead of the Apple ID/app-specific password notarization secrets, CI can use:

- `APP_STORE_CONNECT_API_KEY_KEY_ID`
- `APP_STORE_CONNECT_API_KEY_ISSUER_ID`
- `APP_STORE_CONNECT_API_KEY_P8_BASE64`

## Tag A Release

```sh
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

The workflow builds linked FFmpeg binaries for Linux, Windows, and macOS. macOS binaries are signed with Developer ID, submitted to Apple notarization, assessed with `spctl`, then checksummed.

The installer downloads the matching release asset and `.sha256` file before installing.

## Local Signing Check

```sh
security find-identity -v -p codesigning | grep "Developer ID Application"
```

## Local macOS Signing

```sh
./scripts/build-linked.sh darwin-arm64
./scripts/sign-macos.sh build/podhnologic-darwin-arm64
```

The signing script auto-detects `Developer ID Application: Wabi Sabi Ware LLC (88M7JPMLS6)` and signs with the hardened runtime using `com.wabisabiware.podhnologic`.

## Local macOS Notarization

Using a stored notarytool profile:

```sh
PODHNOLOGIC_NOTARY_PROFILE=profile-name \
  ./scripts/notarize-macos.sh build/podhnologic-darwin-arm64
```

Using an App Store Connect API key:

```sh
PODHNOLOGIC_ASC_API_KEY=<ASC_API_KEY_ID> \
PODHNOLOGIC_ASC_API_ISSUER=<issuer-uuid> \
PODHNOLOGIC_ASC_API_KEY_PATH=~/Desktop/AuthKey_<ASC_API_KEY_ID>.p8 \
  ./scripts/notarize-macos.sh build/podhnologic-darwin-arm64
```

## Notarization Logs

```sh
xcrun notarytool log SUBMISSION_ID \
  --apple-id "$APPLE_ID" \
  --team-id "$APPLE_TEAM_ID" \
  --password "$APPLE_APP_SPECIFIC_PASSWORD"
```

## Licensing Check

Native release artifacts statically link GPL-enabled FFmpeg plus LAME and Opus. Before public release, make sure the project license and release assets satisfy the combined binary's license obligations.
