# Linked FFmpeg

podhnologic links FFmpeg into the native executable.

The old model of downloading static FFmpeg binaries, embedding them with Go build tags, and extracting them to `~/.podhnologic/bin` has been removed.

## Build Flow

```sh
./scripts/build-linked.sh darwin-arm64
```

`scripts/build-linked.sh`:

- builds pinned FFmpeg, LAME, and Opus sources under `scripts/ffmpeg/dist/<target>`
- archives the FFmpeg CLI bridge into `libpodhnologicffmpeg.a`
- links that archive into the Go binary with `linkedffmpeg_cgo linkedffmpeg_hidden`
- writes the binary to `build/podhnologic-<target>`

The app invokes FFmpeg and FFprobe through a hidden self-process bridge. That keeps FFmpeg's CLI behavior and argument semantics while avoiding any external binary.

## Native Targets

- `darwin-arm64`
- `darwin-amd64`
- `linux-amd64`
- `linux-arm64`
- `windows-amd64`
- `windows-arm64` when an ARM64 MinGW toolchain is available

## Source Versions

Pinned versions and checksums live in `scripts/ffmpeg/versions.env`.

## License

The linked native binary includes GPL-enabled FFmpeg components. Release builds need compatible project licensing and complete corresponding source access for FFmpeg, LAME, Opus, and the bridge code.
