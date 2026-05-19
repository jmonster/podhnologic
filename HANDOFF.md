# Handoff: FFmpeg 8 Native/WASM Work

## Current Truth

- Native host fallback lane: `go test -count=1 .` passes on macOS.
- Native production linked-FFmpeg lane: not freshly proven in this slice.
- Windows/Linux production lanes: not freshly proven in this slice.
- Web lane: not working yet. Astro loads the custom FFmpeg 8 WASM core with no `/lib/audio-ffmpeg` 404, but the first real conversion fails with `FFmpeg exited with code -6`.
- Full browser-core production build proof was delegated to subagent `019e3d3c-995d-77f0-8a05-3c46a09b05bf`; it did not return before shutdown and produced no accepted result.

## User Intent

- Native binary must link FFmpeg directly. No runtime FFmpeg download, no external binary lookup.
- FFmpeg flags/effects must preserve existing output behavior, especially iTunes/iPod compatibility.
- Browser/WASM must use FFmpeg 8+ and native FFmpeg `opus` encoder, not `libopus`.
- Native CLI may use `libopus`; FFmpeg docs describe native `opus` as experimental, while `libopus` is the mature external encoder.
- Audio-only work is CPU/SIMD, not GPU. Multicore value is file-level parallel conversion on native.

## Files Changed In This Slice

- `scripts/ffmpeg/browser-core.Dockerfile`
  - Builds FFmpeg from `FFMPEG_VERSION=8.1.1`.
  - Browser core uses native FFmpeg `opus` encoder.
  - Removed browser `libopus` build/link path.
  - Keeps LAME and zlib.
  - Generates missing FFmpeg 8 `fftools/resources/graph.css.c` and `graph.html.c`.
  - Renames FFmpeg CLI `main` to exported `ffmpeg`.
  - Builds both UMD and ESM core outputs.

- `scripts/ffmpeg/browser-ffmpeg-wasm.sh`
  - Links the FFmpeg 8 CLI source set into the wasm core.
  - Adds `-Icompat/stdbit` for FFmpeg 8 with Emscripten 3.1.40.
  - Exports `_ffmpeg`, `_abort`, `_malloc`.

- `scripts/ffmpeg/build-browser-core.sh`
  - Syncs custom browser core artifacts from `dist/esm` when available.
  - No browser `libopus` args.

- `web/src/lib/audio-ffmpeg.ts`
  - Opus browser output uses native `opus` plus `-strict experimental`.
  - Removed mandatory worker URL because the current browser core is single-threaded.
  - Version guard now accepts FFmpeg 8 by `libavutil 60+`, because the custom git build reports `ffmpeg version 239f2c7` instead of `ffmpeg version 8.x`.
  - Current attempted fix uses short-lived FFmpeg workers: one worker for version check, a fresh worker for each conversion. This did not fix exit `-6`.

- `web/src/lib/converter-ui.ts`
  - Processes selected files in series.
  - Filters visible `Aborted()` noise from the UI log.

- `web/scripts/sync-ffmpeg-core.mjs`
  - Copies custom core from `scripts/ffmpeg/dist/browser-core` into `web/public/ffmpeg-core`.
  - Removes stale `ffmpeg-core.worker.js`.

- `web/tests/converter.spec.ts`
  - Playwright E2E covers:
    - no missing core resource/404 path,
    - all browser output formats,
    - three files processed in series without visible `Aborted()`.
  - Current first test fails before the rest can pass.

## Proof Runs

- `bash -n scripts/ffmpeg/build-browser-core.sh`: passed earlier.
- `bash -n scripts/ffmpeg/browser-ffmpeg-wasm.sh`: passed earlier.
- `pnpm --dir web exec tsc -p tsconfig.json --noEmit`: passed after latest web edits.
- `node --check web/scripts/sync-ffmpeg-core.mjs`: passed.
- `pnpm --dir web sync-core`: passed.
- `go test -count=1 .`: passed.
- `pnpm --dir web exec playwright test tests/converter.spec.ts -g "converts a wav file without 404s"`: failed.

Failing Playwright result:

```text
Expected: "Ready: 1 output file."
Received: "FFmpeg exited with code -6."
Trace: web/test-results/converter-converts-a-wav-file-without-404s-chromium/trace.zip
Error context: web/test-results/converter-converts-a-wav-file-without-404s-chromium/error-context.md
```

## Browser Core Artifact State

Current synced artifact:

```text
scripts/ffmpeg/dist/browser-core/ffmpeg-core.js    83K
scripts/ffmpeg/dist/browser-core/ffmpeg-core.wasm  3.1M
web/public/ffmpeg-core/ffmpeg-core.js              83K
web/public/ffmpeg-core/ffmpeg-core.wasm            3.1M
```

`ffmpeg-core.js` in `scripts/ffmpeg/dist/browser-core` and `web/public/ffmpeg-core` is ESM and exports `createFFmpegCore`.

## Most Likely Problem

The custom FFmpeg 8 wasm wrapper can load and run `ffmpeg -version`, but aborts on the first real conversion. The app-level one-shot worker change did not change the failure, so the next owner should investigate the wasm CLI wrapper/link layer, not the Astro form.

Specific suspects:

- FFmpeg 8 CLI entrypoint may need more than a direct `main` to `ffmpeg` rename for wasm re-entry/exit behavior.
- The upstream `ffmpeg.wasm` bind/pre-js may be incompatible with FFmpeg 8 CLI lifecycle without patching `exit`, `abort`, or `ret` handling.
- Current logs only expose `FFmpeg exited with code -6`; add structured instrumentation before another Playwright rerun.

## Next Slice

1. Add focused instrumentation for the failing browser conversion.
   - Emit a keyed log before `ffmpeg.exec` with the exact args.
   - Capture the final 50 FFmpeg log lines on nonzero exit.
   - Surface a keyed event such as `WEB_FFMPEG_EXEC_FAIL` with `format`, `exitCode`, `args`, and tail logs.

2. Reproduce with the smallest browser command.
   - First try WAV to WAV: `-i tone.wav -map 0:a:0 -c:a pcm_s16le out.wav`.
   - If WAV works, isolate muxer/encoder failures per format.
   - If WAV also returns `-6`, debug the CLI wrapper/runtime lifecycle.

3. If wrapper lifecycle is the issue, patch the generated core build inputs, not the app UI.
   - Likely files: `scripts/ffmpeg/browser-core.Dockerfile` and `scripts/ffmpeg/browser-ffmpeg-wasm.sh`.
   - Rebuild/sync core, then rerun the same failing Playwright test.

4. Once the first browser test is green, run:

```sh
pnpm --dir web exec playwright test tests/converter.spec.ts
```

5. Then prove native production separately:

```sh
make linked-test
./scripts/build-linked.sh darwin-arm64
./scripts/build-linked.sh linux-amd64
./scripts/build-linked.sh windows-amd64
```

Do not claim Windows/Linux/web support until those lanes actually convert files, not just compile.

## Constraints For Next Owner

- Do not reintroduce browser `libopus`; user explicitly chose native `opus` in wasm.
- Do not add fallback external FFmpeg downloads.
- Do not weaken Playwright tests to hide the `-6` failure.
- Do not rerun the same failing Playwright lane without adding instrumentation or changing the hypothesis.
- Use `pnpm`, not `npm`.
- Keep changes local and small.
