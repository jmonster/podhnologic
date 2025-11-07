# FFmpeg Binary Embedding

## Overview

As of version 3.0, podhnologic embeds static FFmpeg binaries directly into the executable, making it truly self-contained with **no external dependencies**.

## How It Works

### Build Process

1. **Download Binaries** (one-time setup for developers):
   ```sh
   make download-binaries
   ```
   This downloads static FFmpeg builds for all platforms (~1.3GB total) into `binaries/`

2. **Platform-Specific Embedding**:
   - When building for a specific platform, only that platform's binaries are embedded
   - Achieved using Go build tags in separate `embedded_*.go` files
   - Each platform file uses `//go:embed` to include only its binaries

3. **Build Outputs**:
   - macOS (Intel/ARM): ~162MB each
   - Linux (amd64): ~373MB
   - Linux (arm64): ~297MB
   - Windows (amd64): ~372MB

### Runtime Behavior

1. **First Run**:
   - On first execution, podhnologic checks if it has embedded binaries
   - If yes, extracts FFmpeg and FFprobe to `~/.podhnologic/bin/`
   - Extraction happens only once (checks if files already exist)

2. **Subsequent Runs**:
   - Uses already-extracted binaries from `~/.podhnologic/bin/`
   - No extraction overhead

3. **Fallback Chain**:
   - **Priority 1**: Embedded binaries (extracted to ~/.podhnologic/bin/)
   - **Priority 2**: System PATH
   - **Priority 3**: Custom path via `--ffmpeg` flag
   - **Priority 4**: Download from internet (if embedded not available)

## Files Involved

### Core Implementation
- `embedded.go` - Core extraction logic
- `embedded_darwin_amd64.go` - macOS Intel binaries
- `embedded_darwin_arm64.go` - macOS Apple Silicon binaries
- `embedded_linux_amd64.go` - Linux x86_64 binaries
- `embedded_linux_arm64.go` - Linux ARM64 binaries
- `embedded_windows_amd64.go` - Windows x86_64 binaries

### Build Infrastructure
- `scripts/download-ffmpeg.sh` - Downloads all platform binaries
- `Makefile` - Updated with `download-binaries` target and build checks
- `binaries/` - Directory containing static FFmpeg builds (gitignored)
- `binaries/README.md` - Documentation about binary sources

### Updated Logic
- `main.go:ensureFFmpeg()` - Updated to prioritize embedded binaries

## Benefits

✅ **Zero External Dependencies** - No need to install FFmpeg
✅ **Consistent Experience** - Same FFmpeg version across all platforms
✅ **Simplified Distribution** - Single executable to distribute
✅ **Offline Capable** - Works without internet connection
✅ **Version Control** - FFmpeg version bundled with each release

## Trade-offs

⚠️ **Larger Binary Size** - Executables are 150-380MB vs ~10-15MB without embedding
⚠️ **Build Complexity** - Requires downloading binaries before building
⚠️ **Update Cycle** - FFmpeg updates require rebuilding and redistributing

## Why This Approach?

We chose **embedded static binaries** over alternatives because:

### vs CGo + Dynamic Linking
- ✅ Maintains single-binary distribution
- ✅ Simple cross-compilation
- ✅ No platform-specific library dependencies

### vs External Download Only
- ✅ Works offline
- ✅ No network failure points
- ✅ Faster first-run experience

### vs No Embedding (Current v2.x)
- ✅ Better UX (no installation steps)
- ✅ More reliable (no PATH issues)
- ✅ Truly portable

## Developer Guide

### Building with Embedded Binaries

```sh
# First time setup
make download-binaries

# Build for current platform
make build

# Build for all platforms
make build-all
```

### Building without Binaries (Emergency)

If the `binaries/` directory doesn't exist, the build will fail with a helpful error.
The runtime code will gracefully fall back to other methods (PATH, download, custom path).

### Updating FFmpeg Version

```sh
# Remove old binaries
rm -rf binaries/

# Download latest
make download-binaries

# Rebuild
make build-all
```

### Testing

```sh
# Clean config to force re-extraction
rm -rf ~/.podhnologic

# Run with dry-run to test extraction
./build/podhnologic --input /tmp/in --output /tmp/out --codec aac --dry-run

# Verify extracted binaries
ls -lh ~/.podhnologic/bin/
~/.podhnologic/bin/ffmpeg -version
```

## Binary Sources

### macOS
- **Source**: [evermeet.cx](https://evermeet.cx/ffmpeg/)
- **Type**: Universal binaries (Intel + Apple Silicon)
- **License**: GPL v3
- **Size**: ~153MB (ffmpeg + ffprobe)

### Linux
- **Source**: [BtbN/FFmpeg-Builds](https://github.com/BtbN/FFmpeg-Builds)
- **Type**: Static builds with GPL codecs
- **License**: GPL v3
- **Size**: ~290-364MB per platform

### Windows
- **Source**: [BtbN/FFmpeg-Builds](https://github.com/BtbN/FFmpeg-Builds)
- **Type**: Static builds with GPL codecs
- **License**: GPL v3
- **Size**: ~363MB

## License Implications

podhnologic itself is licensed under CC BY-NC-ND 4.0.

The embedded FFmpeg binaries are licensed under **GPL v3** (due to included GPL codecs like x264, x265, etc.).

**Note**: Since FFmpeg is a separate binary (not linked as a library), podhnologic's license remains independent. The FFmpeg binaries are distributed alongside, not integrated into, the codebase.

## Future Improvements

Possible optimizations for the future:

1. **Compression**: Compress binaries with UPX or similar
2. **Lazy Download**: Only download when first needed (hybrid approach)
3. **Minimal Builds**: Create smaller FFmpeg builds with fewer codecs
4. **Platform Detection**: Only embed for detected platform during CI/CD

For now, we prioritize simplicity and user experience over binary size optimization.
