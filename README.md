# the `g` is silent.

## how?

```sh
pnpx github:jmonster/music-monstger --input "/path/to/input" --output "/path/to/output" --ipod
```

## what?

This is a little tooling I put together so I could convert my FLAC music collection to ALAC or 256 AAC for the Apple iPod.

## why?

Historically iTunes/Apple Music has been the preferred way to convert because of Apple's superior encoder. If you run this on macOS / if `aac_at` is available, then this will tool will also use Apple's encoder. You are not sacrificing quality using this.

## quality

This tool is simple and opinionated; I assume you want the best possible quality. Therefore, we us the following values:

- `alac` & `flac`: lossless
- `aac`: 256K w/Apple's encoder (where available)
- `ogg`: libvorbis -q:a 8 which is the edge of human perception
- `wav`: pcm_s16le (should we be doing something different?)
- `mp3`: 320kbps

The `--ipod` flag is a shortcut for 256 AAC since that's typically the most practical choice for an iPod

## performance

Runs as many threads as your machine; it's very fast. With 10 cores, it's 10x faster than Apple Music at converting ALAC to AAC.

## options

```sh
node script.js \
  --input <inputDir> \
  --output <outputDir> \
  --codec [flac|alac|aac|wav|mp3|ogg] \
  [--ipod] \
  [--ffmpeg /opt/homebrew/bin/ffmpeg] \
  [--dry-run]
```
