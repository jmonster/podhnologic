# what?

- a tool to convert your music collection from A to B
- written for and used by me to convert my FLAC library to ALAC + 256 AAC

# why?

Apple's Music encoder is single-threaded and requires you to import your library into it before you can (very slowly) convert it.
Since it's considered the best, we're going to use Apple's encoder if it's available on your system.

# quality

This tool is simple and opinionated. I assume you want the best possible but practical quality.

- `alac` & `flac`: lossless
- `aac`: 256K w/Apple's encoder (where available)
- `ogg`: libvorbis -q:a 8 which is the edge of human perception
- `wav`: pcm_s16le (should we be doing something different?)
- `mp3`: 320kbps

The `--ipod` flag is shorthand for 256kbps AAC

# performance

Since iTunes is single-threaded, this script runs `X`-times faster while utilizing the same encoder on an machine with `X` cores

# how?

## requirements

- [ffmpeg](https://ffmpeg.org) must be installed, or at least located locally, such that you can specify with the path with `--ffmpeg` option. Uses your system path by default
- [node.js](https://nodejs.org) is used to executed and run this tool; they must be installed such that the `npx` command succeeds

There are no other dependencies.

```sh

npx github:jmonster/music-monstger --input "/path/to/input" --output "/path/to/output" --ipod
```

# options

```sh
node script.js \
  --input <inputDir> \
  --output <outputDir> \
  --codec [flac|alac|aac|wav|mp3|ogg] \
  [--ipod] \
  [--ffmpeg /opt/homebrew/bin/ffmpeg] \
  [--dry-run]
```
