# Changelog

All notable changes to this project will be documented in this file.

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
