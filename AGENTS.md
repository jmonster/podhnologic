# AGENTS.md

## macOS Release Signing

- macOS release binaries are raw CLI executables, not `.app` bundles.
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
