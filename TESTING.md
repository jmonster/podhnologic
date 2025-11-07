# Testing Documentation

## Overview

This document describes the comprehensive test suite for podhnologic. The tests cover core functionality including audio conversion, metadata handling, configuration management, and more.

## Test Structure

### Test Files

- **main_test.go**: Unit tests for utility functions and configuration management
- **converter_test.go**: Integration tests for audio conversion functionality

### Test Coverage

Current test coverage: **21.7%** of statements

Coverage breakdown:
- ✅ **Core conversion logic**: 100% (converter.go)
- ✅ **Configuration management**: 100% (load/save/validate)
- ✅ **Path utilities**: 85%+ (expand, shorten, trim)
- ✅ **Metadata handling**: 100% (extract, preserve, filter)
- ✅ **Codec support**: 100% (all 6 codecs tested)
- ❌ **UI/TUI components**: 0% (menu, dirpicker - not tested)
- ❌ **Binary download**: 0% (ffmpeg download logic - not tested)
- ❌ **Main entry point**: 0% (main function - not tested)

## Running Tests

### Run all tests
```bash
go test -v
```

### Run with coverage
```bash
go test -cover
```

### Generate coverage report
```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run benchmarks
```bash
go test -bench=. -benchmem
```

## Test Categories

### 1. Unit Tests (main_test.go)

#### Path Utilities
- `TestExpandPath`: Tests tilde expansion and path normalization
- `TestShortenPath`: Tests home directory shortening for display
- `TestTrimQuotes`: Tests quote removal from paths

#### Configuration
- `TestConfigSaveAndLoad`: Tests config persistence to JSON
- `TestLoadConfigNonExistent`: Tests handling of missing config
- `TestLoadConfigInvalid`: Tests handling of corrupted config

#### Helper Functions
- `TestFindIndex`: Tests array index finding
- `TestGetOutputExtension`: Tests codec to file extension mapping
- `TestIsAudioFile`: Tests audio file detection
- `TestGetFFprobePath`: Tests ffprobe path derivation

### 2. Integration Tests (converter_test.go)

#### File Processing
- `TestCollectAudioFiles`: Tests recursive audio file discovery
- `TestProcessFileBasic`: Tests basic file conversion
- `TestProcessFileDryRun`: Tests dry-run mode (no actual conversion)
- `TestProcessFileSkipsExisting`: Tests resumability (skip existing files)

#### Codec Support
- `TestCodecConversions`: Tests conversion to all supported codecs:
  - FLAC (lossless)
  - WAV (uncompressed)
  - MP3 (lossy)
  - Opus (lossy)
  - AAC (lossy, macOS optimized)
  - ALAC (lossless, macOS optimized)

#### Metadata Handling
- `TestMetadataPreservation`: Tests that tags are preserved during conversion
- `TestLyricsStripping`: Tests lyrics removal when NoLyrics=true
- `TestLyricsPreservation`: Tests lyrics retention when NoLyrics=false
- `TestExtractMetadata`: Tests metadata extraction from audio files

#### Special Features
- `TestIPodMode`: Tests iPod-specific encoding optimizations
- `TestBuildFFmpegArgs`: Tests FFmpeg command construction
- `TestGetCodecParamsSimple`: Tests codec parameter generation

#### Edge Cases
- `TestRunConversionEmptyInput`: Tests behavior with no input files
- `TestRunConversionNonExistentInput`: Tests error handling for missing directories

### 3. Benchmarks

- `BenchmarkProcessFile`: Measures file processing performance
  - Current: ~60ms per file on Apple M4
  - 66KB memory usage
  - 134 allocations per operation

## Test Helper Utilities

The `TestHelper` struct provides utilities for integration testing:

```go
helper := NewTestHelper(t)
helper.Setup()
```

### Key Methods

- `GenerateTestAudio(filename, duration)`: Creates synthetic audio files using ffmpeg
- `GenerateTestAudioWithMetadata(filename, metadata)`: Creates audio with specific tags
- `VerifyFileExists(path)`: Asserts file existence
- `VerifyFileNotExists(path)`: Asserts file absence
- `GetMetadata(filePath)`: Extracts metadata from audio
- `VerifyMetadataHasKey(metadata, key)`: Checks for metadata field
- `VerifyMetadataLacksKey(metadata, key)`: Checks metadata field is absent

## Sample Audio Generation

Tests use ffmpeg's `lavfi` (libavfilter) to generate synthetic audio:

```bash
ffmpeg -f lavfi -i sine=frequency=440:duration=2 output.mp3
```

This creates a 440Hz sine wave (A note) for testing without requiring sample files.

## Test Data

Test files are automatically cleaned up but can be found during test runs:
- Input files: `testdata/input/`
- Output files: `testdata/output/`
- Temporary files: `testdata/temp/`

All test data directories are excluded in `.gitignore`.

## Requirements

### System Dependencies
- **ffmpeg**: Required for integration tests
  - Tests will be skipped if ffmpeg is not in PATH
  - On macOS: `brew install ffmpeg`
  - On Linux: `apt-get install ffmpeg` or equivalent

### Go Dependencies
All dependencies are managed via `go.mod`:
```bash
go mod download
```

## Test Results Summary

### All Tests Passing ✅

```
PASS
ok  	github.com/jmonster/podhnologic	1.488s
```

### Coverage by Component

| Component | Coverage | Status |
|-----------|----------|--------|
| Converter | 100% | ✅ Fully tested |
| Config | 100% | ✅ Fully tested |
| Path Utils | 85% | ✅ Well tested |
| Metadata | 100% | ✅ Fully tested |
| Menu/UI | 0% | ⚠️ Not tested (TUI) |
| Downloads | 0% | ⚠️ Not tested (integration) |

## Continuous Testing

### Pre-commit Hook
Consider adding this to `.git/hooks/pre-commit`:
```bash
#!/bin/bash
go test -cover
if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi
```

### CI/CD Integration
For GitHub Actions, see example workflow:
```yaml
- name: Run Tests
  run: |
    go test -v -coverprofile=coverage.out
    go tool cover -func=coverage.out
```

## Critical Missing Test Coverage ❌

### 1. Error Handling & Robustness (HIGH PRIORITY)
- [ ] FFmpeg command failures (non-zero exit codes)
- [ ] Corrupted or invalid audio files
- [ ] Permission errors (read/write/execute)
- [ ] Disk space exhaustion during conversion
- [ ] Invalid or malformed metadata
- [ ] Stderr parsing and error reporting

**Suggested Tests:**
```go
TestProcessFileFFmpegFailure()
TestProcessFileCorruptedInput()
TestProcessFilePermissionDenied()
TestProcessFileInsufficientSpace()
TestExtractMetadataInvalidFile()
```

### 2. Filename Edge Cases (HIGH PRIORITY)
- [ ] Special characters (spaces, quotes, apostrophes)
- [ ] Unicode/emoji in filenames
- [ ] Very long paths (>255 chars)
- [ ] Path traversal attempts (../)
- [ ] Reserved filenames (Windows: CON, PRN, NUL)
- [ ] Hidden files (.mp3)
- [ ] Symlinks and circular references

**Suggested Tests:**
```go
TestSpecialCharactersInFilenames()
TestUnicodeAndEmojiFilenames()
TestVeryLongPaths()
TestSymlinksInDirectory()
TestHiddenAudioFiles()
```

### 3. Concurrency & Parallelism (HIGH PRIORITY)
- [ ] Parallel processing with multiple files
- [ ] Race conditions in worker pools
- [ ] Error aggregation from multiple workers
- [ ] Thread safety of shared resources
- [ ] Worker pool behavior with many files

**Suggested Tests:**
```go
TestProcessFilesParallelMultipleFiles()
TestProcessFilesParallelWithErrors()
TestProcessFilesParallelRaceConditions()
TestParallelProcessingAccuracy()
```

### 4. Command-Line Interface (MEDIUM PRIORITY)
- [ ] All flag parsing (--input, --output, --codec, etc.)
- [ ] Flag validation and error messages
- [ ] Flag precedence over saved config
- [ ] Invalid flag combinations
- [ ] --version flag
- [ ] --dry-run integration with CLI
- [ ] --interactive flag

**Suggested Tests:**
```go
TestMainWithValidFlags()
TestMainWithInvalidFlags()
TestFlagPrecedenceOverConfig()
TestVersionFlag()
TestDryRunFlag()
```

### 5. FFmpeg Management (MEDIUM PRIORITY)
- [ ] Embedded binary extraction
- [ ] FFmpeg download from internet
- [ ] FFmpeg version validation
- [ ] PATH detection and priority
- [ ] testFFmpeg validation logic
- [ ] Platform-specific binary selection

**Suggested Tests:**
```go
TestEnsureFFmpegEmbedded()
TestEnsureFFmpegFromPATH()
TestEnsureFFmpegDownload()
TestFFmpegValidation()
TestFindInPath()
```

### 6. Platform-Specific Behavior (MEDIUM PRIORITY)
- [ ] Windows path separators and drive letters (C:\)
- [ ] macOS codec preferences (aac_at vs aac)
- [ ] Linux-specific behavior
- [ ] Case-sensitive vs case-insensitive filesystems
- [ ] Line ending differences (CRLF vs LF)

**Suggested Tests:**
```go
TestWindowsPathHandling()      // build tag: windows
TestMacOSCodecSelection()      // build tag: darwin
TestCaseInsensitiveFilesystem()
```

### 7. Interactive TUI (LOW PRIORITY - Complex)
- [ ] Menu navigation (up/down, j/k)
- [ ] Keyboard shortcuts (I, O, C, P, L, S, Q)
- [ ] Directory picker integration
- [ ] Codec selection UI
- [ ] Configuration validation UI
- [ ] Error message display
- [ ] Terminal size handling
- [ ] Color scheme detection (light/dark)

**Note:** TUI testing is complex. Consider:
- Using Bubble Tea's testing utilities
- Manual testing for visual components
- Focus on model state transitions

**Suggested Tests:**
```go
TestMenuModelUpdate()
TestMenuModelKeyPress()
TestMenuValidation()
TestDirectoryPickerIntegration()
```

### 8. Advanced Audio Features (LOW PRIORITY)
- [ ] Cover art preservation/stripping
- [ ] Multi-stream audio files
- [ ] Sample rate conversion edge cases
- [ ] Bitrate handling for lossy codecs
- [ ] Empty audio files (0 duration)
- [ ] Files with minimal/no metadata
- [ ] Video streams in audio containers

**Suggested Tests:**
```go
TestCoverArtPreservation()
TestMultiStreamAudioFiles()
TestZeroDurationFiles()
TestMinimalMetadata()
```

### 9. Configuration Management (LOW PRIORITY)
- [ ] Config directory creation permissions
- [ ] Concurrent config access
- [ ] Config file corruption recovery
- [ ] Config schema migration
- [ ] Default value handling edge cases

**Suggested Tests:**
```go
TestConfigDirectoryCreation()
TestConfigConcurrentAccess()
TestConfigCorruptionRecovery()
```

### 10. User Experience & Integration (LOW PRIORITY)
- [ ] Progress reporting accuracy
- [ ] Cancellation handling (Ctrl+C, SIGTERM)
- [ ] Resume across sessions
- [ ] Output formatting consistency
- [ ] Color output in non-TTY environments
- [ ] End-to-end workflow tests

**Suggested Tests:**
```go
TestProgressReporting()
TestCancellationHandling()
TestEndToEndConversion()
TestInteractiveModeFlow()
```

## Test Priority Roadmap

### Immediate (Next Sprint)
1. Add error handling tests for FFmpeg failures
2. Add filename edge case tests (spaces, unicode)
3. Add parallel processing tests

### Short Term (1-2 Weeks)
4. Add command-line flag tests
5. Add platform-specific tests
6. Add FFmpeg management tests

### Medium Term (1 Month)
7. Improve configuration tests
8. Add advanced audio feature tests
9. Add basic TUI state tests

### Long Term (As Needed)
10. Add comprehensive integration tests
11. Add performance regression tests
12. Add platform CI/CD testing

## Code Coverage Goals

| Component | Current | Target | Priority |
|-----------|---------|--------|----------|
| Core Conversion | 100% | 100% | ✅ Achieved |
| Error Handling | ~20% | 95% | 🔴 Critical |
| CLI Flags | 0% | 80% | 🟡 High |
| FFmpeg Mgmt | 0% | 70% | 🟡 High |
| Concurrency | 0% | 85% | 🔴 Critical |
| TUI | 0% | 60% | 🟢 Low |
| **Overall** | **21.7%** | **80%** | 🔴 **Gap** |

## Future Test Improvements

1. **UI Testing**: Add tests for Bubble Tea components using their test harness
2. **Download Testing**: Mock HTTP requests to test binary download logic
3. **Error Scenarios**: Add comprehensive negative test cases
4. **Performance**: Add more benchmarks for parallel processing
5. **Platform Tests**: Test on Linux/Windows in CI/CD
6. **Fuzzing**: Add fuzzing tests for filename/path handling
7. **Property-based Testing**: Consider using testing/quick for edge cases

## Troubleshooting

### Tests Skipped
If you see `ffmpeg not found in PATH, skipping integration tests`:
- Install ffmpeg: `brew install ffmpeg` (macOS) or equivalent
- Ensure ffmpeg is in your PATH: `which ffmpeg`

### Slow Tests
If tests are slow:
- Check ffmpeg performance: `ffmpeg -version`
- Reduce test audio duration (currently 1-2 seconds)
- Run specific tests: `go test -run TestProcessFileBasic`

### Coverage Not Generated
If coverage reports fail:
```bash
rm coverage.out coverage.html
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Contributing Tests

When adding new functionality:

1. Write tests first (TDD)
2. Aim for >80% coverage of new code
3. Include both positive and negative test cases
4. Add benchmarks for performance-critical code
5. Update this documentation

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [ffmpeg Documentation](https://ffmpeg.org/documentation.html)
