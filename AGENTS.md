# AGENTS.md

## macOS Release Signing

- macOS release binaries are raw CLI executables, not `.app` bundles.
- Before compiling a release, run `make release-preflight`. It must prove the
  exact Developer ID identity can sign through the logged-in user's normal
  keychain search and that the explicit owner-only App Store Connect `.p8` works
  against Apple. Do not start a release build after either check fails.
- Install the Developer ID certificate/private-key pair once in the normal login
  keychain. Repository scripts must never create, delete, unlock, lock, import
  into, write, ACL-edit, reconfigure, or replace a keychain or its search list.
- App Store Connect authentication uses an `AuthKey_<key-id>.p8` outside the
  repository with mode `0400` or `0600`. Pass its absolute path, key ID, and
  issuer ID; never store or transport its bytes through base64, a Keychain
  generic-password item, an environment variable, or a temporary decoded file.
- Sign with Developer ID:

```sh
./scripts/sign-macos.sh build/podhnologic-darwin-arm64
```

- The expected identity is:

```text
Developer ID Application: Wabi Sabi Ware LLC (88M7JPMLS6)
```

- The expected bundle identifier for signing is:

```text
com.wabisabiware.podhnologic
```

- Notarize with the App Store Connect API key from the shell environment:

```sh
./scripts/notarize-macos.sh build/podhnologic-darwin-arm64
```

- Verify raw CLI binaries with `spctl --type install`, not `--type execute`:

```sh
spctl -a -vvv -t install build/podhnologic-darwin-arm64
```

- Passing output should include:

```text
accepted
source=Notarized Developer ID
origin=Developer ID Application: Wabi Sabi Ware LLC (88M7JPMLS6)
```

- A `spctl --type execute` failure saying `the code is valid but does not seem to be an app` is not the right validation path for this raw CLI artifact.
