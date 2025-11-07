<p align="center" >
  <b>podhnologic</b><br/>
  Convert your music collection to another format; e.g. iPod
<br/><br />
  <img alt="hero image" src="https://github.com/user-attachments/assets/a9383166-c1e6-432e-9658-9044b13725bc" width="256" height="256">
</p>

# ✨ Version 4.0 - The Ultimate Music Converter!

**New in v4.0:**
- 🎨 **Beautiful TUI** - Gorgeous terminal interface with ASCII art logo
- ⚡ **Quick Config** - See all settings at once with keyboard shortcuts
- 📁 **Visual Directory Picker** - Browse and select with arrow keys (Bubble Tea)
- 🔄 **Instant Toggles** - Press P for iPod mode, L for lyrics, etc.
- 💾 **Auto-Save** - Changes saved instantly, just press Enter to start

**Since v3.0:**
- 🚀 **Single binary** - No Node.js required!
- 📦 **Embedded ffmpeg** - FFmpeg binaries bundled inside, zero dependencies!
- 🔄 **Cross-platform** - Linux (amd64/arm64), macOS (Intel/Apple Silicon), Windows
- ⚡ **Efficient & resumable** - Multi-threaded Go performance

## Quick Start

### Interactive Mode (Recommended)

```sh
# Download and run - it will prompt you for everything
./podhnologic

# Or force interactive mode
./podhnologic --interactive
```

The interactive mode shows your current config with quick keys:
- 🎨 **Bold ASCII logo** with gradient colors
- 📋 **See all settings** - Input, output, codec, iPod mode, lyrics
- ⌨️ **Quick shortcuts** - `[I]` input, `[O]` output, `[C]` codec, `[P]` iPod, `[L]` lyrics
- ⚡ **Instant toggles** - Press P or L to toggle without prompts
- 📁 **Visual picker** - Beautiful Bubble Tea directory browser when changing paths
- 💾 **Auto-save** - All changes saved immediately
- 🚀 **Press Enter** to start - No more "start now?" prompts!

### Command-Line Mode

```sh
./podhnologic \
  --input "/path/to/input" \
  --output "/path/to/output" \
  --ipod \
  --dry-run  # Optional: preview what will be converted
```

# what?

- A self-contained tool to convert your music collection from A to B
- Written in Go for maximum performance and portability
- Originally created to convert FLAC libraries to ALAC + 256 AAC for iPods

# why?

Apple's Music encoder is single-threaded and requires you to import your library into it before you can (very slowly) convert it.
Since it's considered the best, we're still going to use Apple's encoder if it's available on your system.

# quality

This tool is simple and opinionated. I assume you want the best possible but practical quality.

- `alac` & `flac`: lossless (or down-sampled with `--ipod`)
- `aac`: 256kbps w/Apple's encoder (where available)
- `opus`: 128kbps
- `mp3`: 320kbps
- `wav`: pcm_s16le

# iPod

- `--ipod` is shorthand for 256 kbps AAC
- if `--codec alac` is _also_ specified, tracks will be down-sampled to 16-bit 44.1kHz.
  - This prevents track skipping on the iPod.
- Moves the moov atom at the beginning of the file, which is useful for streaming and playback compatibility.
- Eliminates all metadata except for `title`, `artist`, `album`, `date`, `track`, `genre`, `disc`, and `lyrics`
  - This helps increase the number of track you can fit in memory on an iPod
  - Optionally add `--no-lyrics` to squeeze even more space
- Album art is preserved

# performant

Runs `X`-times faster than iTunes while utilizing the same encoder on a machine with `X` idle cores

<img width="434" alt="image" src="https://github.com/jmonster/podhnologic/assets/368767/8a50948c-1e63-441d-8df8-ea3bebd75895">

# resumable

Checks if the output file already exists before converting. If a big job gets interrupted, just re-run the same command and the files that are already done will be skipped.

# installation

## Download Pre-built Binary

Download the latest release for your platform from the [Releases](https://github.com/jmonster/podhnologic/releases) page:

- **Linux (amd64)**: `podhnologic-linux-amd64`
- **Linux (arm64)**: `podhnologic-linux-arm64`
- **macOS (Intel)**: `podhnologic-darwin-amd64`
- **macOS (Apple Silicon)**: `podhnologic-darwin-arm64`
- **Windows**: `podhnologic-windows-amd64.exe`

Make it executable (Linux/macOS):
```sh
chmod +x podhnologic-*
sudo mv podhnologic-* /usr/local/bin/podhnologic
```

## Build from Source

Requires [Go 1.21+](https://golang.org/dl/)

```sh
git clone https://github.com/jmonster/podhnologic.git
cd podhnologic

# Download ffmpeg binaries for embedding (required for first build)
make download-binaries

# Build for current platform
make build

# Optional: Install to /usr/local/bin
sudo make install
```

### Build for all platforms

```sh
# Download binaries first (if not already done)
make download-binaries

# Build all platform binaries
make build-all
```

This creates binaries in the `build/` directory for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

Each binary is fully self-contained with embedded ffmpeg for its platform (~160-180MB per binary).

## ffmpeg

**FFmpeg is now embedded!** The tool comes with static ffmpeg binaries built-in. On first run, it will extract them to `~/.podhnologic/bin/`.

No installation needed! Just download and run.

The tool automatically uses embedded binaries:
1. **First run**: Extracts FFmpeg/FFprobe to `~/.podhnologic/bin/`
2. **Subsequent runs**: Uses already-extracted binaries (instant startup)
3. **Fallback**: If embedded binaries fail, checks system PATH

No configuration needed - just run it!

# usage

## Interactive Mode

Just run the binary without arguments:

```sh
podhnologic
```

You'll be prompted for:
- Input directory
- Output directory
- Target codec
- iPod optimizations
- Lyrics handling
- Custom ffmpeg path (optional)

Your choices are saved to `~/.podhnologic/config.json` and will be used as defaults next time.

## Command-Line Mode

### All Options

```sh
podhnologic \
  --input <inputDir> \
  --output <outputDir> \
  --codec [flac|alac|aac|wav|mp3|opus] \
  [--ipod] \
  [--no-lyrics] \
  [--ffmpeg /path/to/ffmpeg] \
  [--dry-run]
```

### Examples

Basic iPod conversion (256kbps AAC):
```sh
podhnologic \
  --input "/path/to/input" \
  --output "/path/to/output" \
  --ipod
```

iPod with ALAC (lossless, down-sampled):
```sh
podhnologic \
  --input "/path/to/input" \
  --output "/path/to/output" \
  --ipod \
  --codec alac
```

Preview before converting:
```sh
podhnologic \
  --input "/path/to/input" \
  --output "/path/to/output" \
  --codec aac \
  --dry-run
```

Strip lyrics to save iPod memory:
```sh
podhnologic \
  --input "/path/to/input" \
  --output "/path/to/output" \
  --ipod \
  --no-lyrics
```

## Configuration

Settings are stored in `~/.podhnologic/config.json`:

```json
{
  "input_dir": "/path/to/music",
  "output_dir": "/path/to/converted",
  "codec": "aac",
  "ipod": true,
  "no_lyrics": false,
  "ffmpeg_path": ""
}
```

You can edit this file manually or use interactive mode to update it.

## tips

For iPod, just use `--ipod` (256-kbps AAC) and be done with it.

- **battery life**
  - less storage accesses / wake ups
  - hardware is optimized for decoding aac
- **storage**
  - files are significantly smaller
- **transparent**
  - 256 AAC is **not** lossless, but it is transparent
    - _If you can tell the difference **and** care about that difference then, by all means, `--codec alac` is for you_

# migration from v2.x (Node.js)

If you were using the Node.js version:

1. **Build or download** the new Go binary
2. **Run it once** in interactive mode to set up your configuration
3. **Same directories work** - the conversion logic is identical
4. **Resume support** - existing converted files will be skipped automatically
5. **Uninstall old version** (optional): `npm uninstall -g podhnologic`

Your existing converted files are fully compatible. The new version will skip them just like the old one did.

# disclaimer

This software is provided "as is", without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose, and noninfringement. In no event shall the authors be liable for any claim, damages, or other liability, whether in an action of contract, tort, or otherwise, arising from, out of, or in connection with the software or the use or other dealings in the software.

# license

This project is licensed under the Creative Commons Attribution-NonCommercial-NoDerivatives 4.0 International License. See the [LICENSE](LICENSE.md) file for more details.
