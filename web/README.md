# Web

Retained Astro experiment for local audio conversion. It is not hosted and is
outside routine CLI builds, releases, and dependency upgrades. Update or rebuild
it only when browser work is explicitly requested.

The browser core requires threads for FFmpeg's audio pipeline. From the
repository root, build it with `./scripts/ffmpeg/build-browser-core.sh` before
starting the web workspace. Set `FFMPEG_WASM_REF` to an official ffmpeg.wasm
checkout. This requires Apple Container, Docker, or Podman and produces the
JavaScript, WebAssembly, and pthread worker files copied by `pnpm sync-core`.
The FFmpeg source archive and checksum are shared with the native build in
`scripts/ffmpeg/versions.env`. Browser AAC uses the NMR coder with PNS disabled.

## Run

```sh
pnpm install --frozen-lockfile
pnpm dev
```

Astro's development and preview servers send the required response headers:

```text
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Embedder-Policy: require-corp
```

When serving the static build elsewhere, configure the host to send these
headers and serve it over HTTPS (or localhost). All three core files must be
available under `/ffmpeg-core`, including `ffmpeg-core.worker.js`.
