# Web

Experimental Astro workspace for local audio conversion UI work.

The browser core requires threads for FFmpeg 8's audio pipeline. From the
repository root, build it with `./scripts/ffmpeg/build-browser-core.sh` before
starting the web workspace. This requires Docker or Podman and produces the
JavaScript, WebAssembly, and pthread worker files copied by `pnpm sync-core`.

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
