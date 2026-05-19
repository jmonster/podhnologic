# Linked FFmpeg

Native builds link the pinned FFmpeg toolchain into the podhnologic executable.

## Build Flow

```sh
./scripts/build-linked.sh darwin-arm64
```

`scripts/build-linked.sh`:

- builds pinned FFmpeg, LAME, Opus, and zlib sources under `scripts/ffmpeg/dist/<target>`
- archives the FFmpeg CLI bridge into `libpodhnologicffmpeg.a`
- links that archive into the Go command with `linkedffmpeg_cgo linkedffmpeg_hidden`
- writes the binary to `build/podhnologic-<target>`

The command invokes FFmpeg and FFprobe through a hidden self-process bridge to preserve FFmpeg CLI argument behavior.

## Native Targets

- `darwin-arm64`
- `darwin-amd64`
- `linux-amd64`
- `linux-arm64`
- `windows-amd64`
- `windows-arm64` when an ARM64 MinGW toolchain is available

## Source Versions

Pinned versions and checksums live in `scripts/ffmpeg/versions.env`.

## License Policy

Native release builds use FFmpeg's LGPL configuration. GPL or nonfree FFmpeg components require a project license and release-process decision first.

Release assets that include linked FFmpeg must include third-party notices and source access for the exact FFmpeg, LAME, Opus, zlib, and bridge sources used to build the artifact.
