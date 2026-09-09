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
Requests travel as JSON over stdin, so large lyrics tags do not hit operating
system command-line limits. This transport shipped in v4.1.0.

## Native Targets

- `darwin-arm64`
- `darwin-amd64`
- `linux-amd64`
- `linux-arm64`
- `windows-amd64`
- `windows-arm64` when an ARM64 MinGW toolchain is available

## Source Versions

Pinned versions and checksums live in `scripts/ffmpeg/versions.env`.

Native and browser builds use upstream FFmpeg commit
`cc7a69a36a67acb661ceb2339a41488fb7409b57` (September 9, 2026), the same source
pin as PixelPrism. This is a development snapshot containing the new NMR AAC
encoder, not the FFmpeg 9.0 release branch. Podhnologic retains its own minimal
audio configuration and does not use PixelPrism's video dependencies or patches.

Native AAC explicitly selects `-aac_coder nmr`. macOS keeps AudioToolbox
(`aac_at`); iPod mode disables PNS only for the native encoder. The browser
uses NMR with PNS disabled. Playback compatibility should also be checked on
the affected iPod: successful encoding and decoding do not prove its firmware
will play every AAC feature correctly.

## License Policy

Native release builds use FFmpeg's LGPL configuration. GPL or nonfree FFmpeg components require a project license and release-process decision first.

Release assets that include linked FFmpeg must include third-party notices and source access for the exact FFmpeg, LAME, Opus, zlib, and bridge sources used to build the artifact.
