# podhnologic

Convert music libraries to iPod-friendly AAC/ALAC or to FLAC, MP3, Opus, and WAV.

- Interactive terminal UI and direct CLI flags
- FFmpeg linked into the native binary; no runtime download or install
- Parallel conversion across CPU cores
- Resumable output; existing files are skipped
- macOS release binaries are signed and notarized
- Browser build in `web/` for local file conversion with FFmpeg WASM

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/jmonster/podhnologic/main/install.sh | bash
```

The installer downloads the signed release binary for your platform, verifies its SHA-256 checksum, and installs it to `~/.local/bin`. FFmpeg is linked into podhnologic. Users do not install Node, npm packages, Homebrew FFmpeg, or any separate FFmpeg binary.

## Run

Interactive mode:

```sh
podhnologic
```

CLI mode:

```sh
podhnologic --input ~/Music --output ~/Converted --codec flac
```

iPod AAC:

```sh
podhnologic --input ~/Music --output ~/iPod --ipod
```

iPod ALAC, down-sampled for older devices:

```sh
podhnologic --input ~/Music --output ~/iPod --ipod --codec alac
```

Preview without writing files:

```sh
podhnologic --input ~/Music --output ~/Converted --codec flac --dry-run
```

## Flags

- `--input <dir>`: source directory
- `--output <dir>`: destination directory
- `--codec <format>`: `aac`, `alac`, `flac`, `mp3`, `opus`, or `wav`
- `--ipod`: use iPod settings; defaults to 256 kbps AAC when no codec is set
- `--no-lyrics`: drop lyrics metadata
- `--dry-run`: print planned work without converting
- `--interactive`: force the TUI
- `--version`: print the version

Settings save to `~/.podhnologic/config.json`.

## Audio Behavior

Output settings:

| Codec | Settings |
| --- | --- |
| AAC | 256 kbps |
| ALAC | Lossless; with `--ipod`, 16-bit 44.1 kHz |
| FLAC | Lossless |
| MP3 | V0 VBR with LAME |
| Opus | 128 kbps with libopus |
| WAV | 16-bit PCM |

Metadata handling is intentionally strict for iPod compatibility: podhnologic keeps title, artist, album, date, track, genre, disc, lyrics unless `--no-lyrics` is set, and album art. Other metadata is dropped.

## Linked FFmpeg

Production native builds use `scripts/build-linked.sh`. The script builds pinned FFmpeg/LAME/Opus sources, creates `libpodhnologicffmpeg.a`, and links it into the Go binary with cgo.

```sh
make build
```

There is no external `ffmpeg` binary lookup, no first-run extraction, and no runtime FFmpeg download. The app preserves the existing FFmpeg command behavior by invoking the linked FFmpeg CLI entry point through a hidden self-process bridge.

Build specific targets:

```sh
./scripts/build-linked.sh darwin-arm64
./scripts/build-linked.sh darwin-amd64
./scripts/build-linked.sh linux-amd64
./scripts/build-linked.sh linux-arm64
./scripts/build-linked.sh windows-amd64
```

Windows ARM64 is wired in the build scripts as `windows-arm64` when an `aarch64-w64-mingw32` toolchain is available.

## Web Development

The browser app is separate from the native installer. It is in `web/` and only needs Node/pnpm for local web development:

```sh
cd web
pnpm install --frozen-lockfile
pnpm build
```

Deploy the built Astro site with FFmpeg WASM core assets served from `/ffmpeg-core` or set `PUBLIC_FFMPEG_CORE_BASE_URL` before building. The hosting target must send cross-origin isolation headers:

```text
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Embedder-Policy: require-corp
```

## Tests

```sh
go test -count=1 .
make linked-test
cd web && pnpm build
```

## Release

Tagged releases are built by GitHub Actions. macOS artifacts are signed with Developer ID and notarized before checksums are created.

Because FFmpeg is statically linked with GPL-enabled components, release artifacts must ship with compatible licensing and complete corresponding source access for the linked FFmpeg/LAME/Opus build.
