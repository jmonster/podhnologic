# Changelog

All notable changes to this project will be documented in this file.

## [4.0.0] - 2024-11-07

### 🎨 Major UX Overhaul

Complete redesign of the interactive mode with beautiful TUI, embedded binaries, and instant configuration.

### Added

- **Beautiful ASCII Art Logo**: Bold PODHNOLOGIC title with gradient colors (cyan → purple → blue)
- **Quick Config Interface**: See all settings at once with keyboard shortcuts
  - `[I]` - Change Input Directory
  - `[O]` - Change Output Directory
  - `[C]` - Change Codec
  - `[P]` - Toggle iPod Mode (instant!)
  - `[L]` - Toggle Lyrics Stripping (instant!)
  - `[Enter]` - Start Conversion
  - `[Q]` - Quit
- **Visual Directory Picker**: Bubble Tea-powered browser with arrow key navigation
- **Embedded FFmpeg Binaries**: Static binaries bundled in executable
  - Platform-specific embedding (only target platform included)
  - Automatic extraction on first run to `~/.podhnologic/bin/`
  - Binary sizes: macOS ~162MB, Linux ~297-373MB, Windows ~372MB
- **Auto-Save**: All configuration changes saved immediately
- **Colored Output**: Success (green), error (red), info (cyan), warning (yellow)
- **Rainbow Music Notes**: Colorful musical decorations in banner

### Changed

- **Interactive Mode Redesigned**: Shows current config instead of prompting every time
- **Removed "Start now?" Prompt**: Just press Enter to begin
- **Simplified FFmpeg**: Uses embedded binaries exclusively (no custom paths)
- **Instant Toggles**: Press P or L to toggle iPod/Lyrics without confirmation prompts
- **Go 1.24 Required**: Updated from Go 1.21

### Dependencies Added

- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI framework
- `github.com/charmbracelet/bubbles` v0.21.0 - TUI components
- `github.com/charmbracelet/lipgloss` v1.1.0 - Terminal styling

### Removed

- `--ffmpeg` CLI flag (uses embedded binaries)
- Custom ffmpeg path from config
- "Beep boop" robot dialog
- Subtitle text from banner
- "Start now?" confirmation prompt
- Old basic directory picker

### Infrastructure

- **Platform-Specific Embedding**: Build tags for each OS/arch combo
- **Download Script**: `scripts/download-ffmpeg.sh` for CI/CD
- **Updated Workflows**: GitHub Actions download binaries before building
- **Documentation**: Added `EMBEDDING.md` technical guide

### Files Added/Modified

**New:**
- `banner.go` - ASCII art logo and colored output
- `dirpicker_bubbletea.go` - Modern directory picker
- `embedded.go` - Core binary extraction logic
- `embedded_*.go` - Platform-specific binary embedding (5 files)
- `scripts/download-ffmpeg.sh` - Automated binary downloads
- `binaries/README.md` - Binary documentation
- `EMBEDDING.md` - Technical embedding guide

**Modified:**
- `main.go` - Redesigned interactive mode, removed ffmpeg flag
- `converter.go` - Updated to receive ffmpegPath parameter
- `.github/workflows/` - Download binaries in CI/CD
- `README.md` - Updated for v4.0 features

## [3.0.0] - 2024-11-07

### 🚀 Major Rewrite

Complete rewrite from Node.js to Go, making podhnologic a true self-contained application.

### Added

- **Self-contained binary**: No Node.js installation required
- **Interactive CLI mode**: Friendly prompts guide users through setup
- **Configuration persistence**: Settings saved to `~/.podhnologic/config.json`
- **Auto-download ffmpeg**: Automatically downloads and manages ffmpeg in `~/.podhnologic/bin/`
- **Cross-platform support**: Native binaries for:
  - Linux (amd64, arm64)
  - macOS (Intel, Apple Silicon)
  - Windows (amd64)
- **Resumability**: Unchanged from v2.x - skips existing output files
- **Parallel processing**: Unchanged from v2.x - uses all CPU cores
- **Command-line flags**: Full backward compatibility with v2.x CLI arguments
- **Dry-run mode**: Preview conversions without executing them
- **Progress tracking**: Visual progress bars for ffmpeg downloads

### Changed

- **Language**: Migrated from JavaScript (Node.js) to Go
- **Distribution**: Single binary instead of npm package
- **Installation**: Direct binary download or build from source (no npm)
- **Configuration**: JSON file in `~/.podhnologic/` instead of command-line only

### Technical Improvements

- **Faster startup**: No npm/Node.js initialization overhead
- **Smaller footprint**: ~13MB binary vs Node.js + dependencies
- **Better error handling**: More detailed error messages and validation
- **Type safety**: Statically typed Go vs dynamic JavaScript
- **Memory efficiency**: Go's efficient memory management

### Migration from v2.x

Users can seamlessly migrate:
1. Install the new Go binary
2. Run once in interactive mode to configure
3. Continue using existing input/output directories
4. All previously converted files will be automatically skipped

### Backward Compatibility

- All v2.x command-line flags are supported
- Same conversion logic and quality settings
- Same ffmpeg encoders and parameters
- Resume support works with v2.x converted files

---

## [2.0.4] - 2024 (Node.js version)

### Changed

- Updated README with clearer documentation
- Improved link to LICENSE.md

## [2.0.3] - 2024 (Node.js version)

### Changed

- Superficial adjustments to codebase
- Clarified optional parameters in documentation

## [1.x - 2.0.2] - 2023-2024 (Node.js version)

Earlier versions built with Node.js. See git history for details.

---

## Version Numbering

- **v3.x**: Go-based self-contained binary
- **v2.x**: Node.js-based npm package (legacy)
- **v1.x**: Initial Node.js versions (legacy)
