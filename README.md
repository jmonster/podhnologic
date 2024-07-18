# jTunes

Convert your music collection to another format; e.g. iPod

```sh
npx github:jmonster/jTunes \
--input "/path/to/input" \
--output "/path/to/output" \
--ipod
```

# what?

- A tool to convert your music collection from A to B
- Written for and used by me to convert my FLAC library to ALAC + 256 AAC

# why?

Apple's Music encoder is single-threaded and requires you to import your library into it before you can (very slowly) convert it.
Since it's considered the best, we're still going to use Apple's encoder if it's available on your system.

# quality

This tool is simple and opinionated. I assume you want the best possible but practical quality.

- `alac` & `flac`: lossless (or down-sampled with `--ipod`)
- `aac`: 256kbps w/Apple's encoder (where available)
- `opus`: 128kbps
- `mp3`: 320kbps
- `wav`: pcm_s16le (should we be doing something different?)

# iPod

- `--ipod` is shorthand for 256-kbps AAC
- if `--codec alac` is also specified, it'll be down-sampled to 16-bit 44.1kHz to (allegedly) prevent track skipping.
- Moves the moov atom at the beginning of the file, which is useful for streaming and playback compatibility.
- Eliminates all metadata except for `title`, `artist`, `album`, `date`, `track`, `genre`, and `disc`
  - This increases the number of track you can fit in memory on an iPod

# performannt

Runs `X`-times faster than iTunes while utilizing the same encoder on a machine with `X` idle cores

<img width="434" alt="image" src="https://github.com/jmonster/jTunes/assets/368767/8a50948c-1e63-441d-8df8-ea3bebd75895">

# resumable

Checks if the output file already exists before converting. If a big job gets interrupted, just re-run the same command and the files that are already done will be skipped.

# how?

## requirements

- [ffmpeg](https://ffmpeg.org) must be installed, or at least located locally, such that you can specify with the path via `--ffmpeg`.
- [node.js](https://nodejs.org) is used to execute and run this tool

There are no other dependencies.

# options

```sh
npx github:jmonster/jTunes \
  --input <inputDir> \
  --output <outputDir> \
  --codec [flac|alac|aac|wav|mp3|opus] \
  [--ipod] \
  [--ffmpeg /opt/homebrew/bin/ffmpeg] \
  [--dry-run]
```

### examples

```sh
npx github:jmonster/jTunes \
--input "/path/to/input" \
--output "/path/to/output" \
--ipod
```

```sh
npx github:jmonster/jTunes \
--input "/path/to/input" \
--output "/path/to/output" \
--codec alac
```

```sh
npx github:jmonster/jTunes \
--input "/path/to/input" \
--output "/path/to/output" \
--ipod \
--ffmpeg "/opt/homebrew/bin/ffmpeg"
```

## node.js Installation

Follow the instructions for your operating system to [install Node.js](https://nodejs.org/en/download/prebuilt-installer/)

## ffmpeg Installation

Follow the instructions for your operating system to install FFmpeg:

- **Linux**: [Install FFmpeg on Linux](https://ffmpeg.org/download.html#build-linux)
- **Windows**: [Install FFmpeg on Windows](https://ffmpeg.org/download.html#build-windows)
- **macOS**: [Install FFmpeg on macOS](https://ffmpeg.org/download.html#build-mac)

# disclaimer

This software is provided "as is", without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose, and noninfringement. In no event shall the authors be liable for any claim, damages, or other liability, whether in an action of contract, tort, or otherwise, arising from, out of, or in connection with the software or the use or other dealings in the software.
