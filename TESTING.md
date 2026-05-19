# Testing

Run focused lanes for the touched scope.

## Go

```sh
go test ./...
```

This covers config, path handling, audio file discovery, FFmpeg argument construction, dry-run behavior, existing-output skips, and linked runner request handling.

## Linked FFmpeg

```sh
make linked-test
```

This builds the current host target through `scripts/build-linked.sh`, then runs tests with `linkedffmpeg_cgo linkedffmpeg_hidden`.

For an end-to-end smoke test, create a small input file with a system FFmpeg and run the linked podhnologic binary against it. The output should probe as the requested codec and preserve the allowed metadata.

## Web

```sh
cd web
pnpm build
```

The web build validates the Astro page and TypeScript bundle. Browser runtime validation requires FFmpeg WASM core assets at `/ffmpeg-core` and cross-origin isolation headers from the host.
