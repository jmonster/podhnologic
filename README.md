# podhnologic

Convert music libraries to iPod-friendly AAC/ALAC or to FLAC, MP3, Opus, and WAV.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/jmonster/podhnologic/main/install.sh | bash
```

The installer verifies the download checksum and installs podhnologic to `~/.local/bin`.

## Use

Interactive mode:

```sh
podhnologic
```

Convert a folder:

```sh
podhnologic --input ~/Music --output ~/Converted --codec flac
```

Prepare files for an iPod:

```sh
podhnologic --input ~/Music --output ~/iPod --ipod
```

Use ALAC for older iPods:

```sh
podhnologic --input ~/Music --output ~/iPod --ipod --codec alac
```

Preview planned conversions:

```sh
podhnologic --input ~/Music --output ~/Converted --codec flac --dry-run
```

## Flags

- `--input <dir>`: source directory
- `--output <dir>`: destination directory
- `--codec <format>`: `aac`, `alac`, `flac`, `mp3`, `opus`, or `wav`
- `--ipod`: use iPod settings; defaults to 256 kbps AAC when no codec is set
- `--no-lyrics`: drop lyrics metadata
- `--dry-run`: show planned conversions
- `--interactive`: force the terminal UI
- `--version`: print the version

Settings save to `~/.podhnologic/config.json`.

## Output

| Codec | Settings |
| --- | --- |
| AAC | 256 kbps |
| ALAC | Lossless; with `--ipod`, 16-bit 44.1 kHz |
| FLAC | Lossless |
| MP3 | V0 VBR with LAME |
| Opus | 128 kbps with libopus |
| WAV | 16-bit PCM |

podhnologic keeps title, artist, album, date, track, genre, disc, lyrics unless `--no-lyrics` is set, and album art. Other metadata is dropped for iPod compatibility.
