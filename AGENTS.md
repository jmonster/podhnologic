# Podhnologic

- `web/` is a retained, unhosted experiment. Routine CLI work and FFmpeg upgrades must not update, rebuild, or deploy it unless the operator explicitly requests browser work.
- macOS release artifacts are raw CLI executables, not application bundles.
- Ordinary implementation and tests do not require Apple release credentials.
- For an explicitly requested macOS release, follow `cmd/podhnologic/docs/release.md` and start with `make release-preflight`.
- Release automation may use the configured Developer ID identity and owner-only App Store Connect `.p8`, but must never create, unlock, import into, write, ACL-edit, select, or reconfigure a keychain.
- Validate the notarized raw CLI with `spctl --type install`; `--type execute` is not the applicable Gatekeeper assessment.
