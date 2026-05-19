# Web

Minimal Astro app for browser-side audio conversion with ffmpeg.wasm.

## Run

```sh
pnpm install --frozen-lockfile
pnpm dev
```

## Core files

Set `PUBLIC_FFMPEG_CORE_BASE_URL` to the local directory that serves:

- `ffmpeg-core.js`
- `ffmpeg-core.wasm`
- `ffmpeg-core.worker.js` for the multithread build

Default: `/ffmpeg-core`

## Multithread headers

The multithread core needs a cross-origin isolated page. Send these headers from the web app host:

- `Cross-Origin-Opener-Policy: same-origin`
- `Cross-Origin-Embedder-Policy: require-corp`
