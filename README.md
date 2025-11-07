<p align="center" >
  <b>podhnologic</b><br/>
  Convert your music collection to another format; e.g. iPod
<br/><br />
  <img alt="hero image" src="https://github.com/user-attachments/assets/a9383166-c1e6-432e-9658-9044b13725bc" width="256" height="256">
</p>

# ✨ Version 3.0 - Now Self-Contained!

**New in v3.0:**
- 🚀 **Single binary** - No Node.js required!
- 💬 **Interactive mode** - Friendly prompts guide you through setup
- 💾 **Saves your settings** - Configuration persisted in `~/.podhnologic/config.json`
- 📦 **Auto-downloads ffmpeg** - Manages ffmpeg automatically in `~/.podhnologic/bin/`
- 🔄 **Cross-platform** - Supports Linux (amd64/arm64), macOS (Intel/Apple Silicon), and Windows
- ⚡ **Efficient & resumable** - Same great performance, now in Go

## Quick Start

### Interactive Mode (Recommended)

```sh
# Download and run - it will prompt you for everything
./podhnologic

# Or force interactive mode
./podhnologic --interactive
```

The interactive mode will:
1. Prompt for input/output directories
2. Let you choose your target codec
3. Ask about iPod optimizations
4. Save your preferences for next time
5. Optionally start conversion immediately

### Command-Line Mode

```sh
./podhnologic \
  --input "/path/to/input" \
  --output "/path/to/output" \
  --ipod
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
make build

# Optional: Install to /usr/local/bin
sudo make install
```

### Build for all platforms

```sh
make build-all
```

This creates binaries in the `build/` directory for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## ffmpeg

The tool will automatically:
1. Check if ffmpeg is installed in your PATH
2. If not found, download a static build to `~/.podhnologic/bin/`
3. Use the downloaded version for all conversions

You can also specify a custom ffmpeg path:
```sh
./podhnologic --ffmpeg "/opt/homebrew/bin/ffmpeg"
```

**Note for macOS users**: If auto-download doesn't work, install ffmpeg via Homebrew:
```sh
brew install ffmpeg
```

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

Custom ffmpeg path:
```sh
podhnologic \
  --input "/path/to/input" \
  --output "/path/to/output" \
  --ipod \
  --ffmpeg "/opt/homebrew/bin/ffmpeg"
```

Dry run (preview what will happen):
```sh
podhnologic \
  --input "/path/to/input" \
  --output "/path/to/output" \
  --codec aac \
  --dry-run
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
