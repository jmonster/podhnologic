package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// embeddedBinariesFS is defined in platform-specific files with build tags
// See: embedded_darwin_amd64.go, embedded_linux_amd64.go, etc.
// Each platform-specific file embeds only the binaries for that platform

// extractEmbeddedFFmpeg extracts the embedded ffmpeg and ffprobe binaries
// to the specified directory if they don't already exist
func extractEmbeddedFFmpeg(destDir string) error {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	// Determine binary names based on platform
	ffmpegName := "ffmpeg"
	ffprobeName := "ffprobe"
	if runtime.GOOS == "windows" {
		ffmpegName = "ffmpeg.exe"
		ffprobeName = "ffprobe.exe"
	}

	// Embedded paths
	embeddedFFmpeg := fmt.Sprintf("binaries/%s/%s", platform, ffmpegName)
	embeddedFFprobe := fmt.Sprintf("binaries/%s/%s", platform, ffprobeName)

	// Destination paths
	destFFmpeg := filepath.Join(destDir, ffmpegName)
	destFFprobe := filepath.Join(destDir, ffprobeName)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Extract ffmpeg if it doesn't exist
	if _, err := os.Stat(destFFmpeg); os.IsNotExist(err) {
		if err := extractFile(embeddedFFmpeg, destFFmpeg); err != nil {
			return fmt.Errorf("failed to extract ffmpeg: %w", err)
		}
		fmt.Printf("  ✓ Extracted ffmpeg\n")
	}

	// Extract ffprobe if it doesn't exist
	if _, err := os.Stat(destFFprobe); os.IsNotExist(err) {
		if err := extractFile(embeddedFFprobe, destFFprobe); err != nil {
			return fmt.Errorf("failed to extract ffprobe: %w", err)
		}
		fmt.Printf("  ✓ Extracted ffprobe\n")
	}

	return nil
}

// extractFile reads a file from the embedded filesystem and writes it to disk
func extractFile(embeddedPath, destPath string) error {
	// Read embedded file
	data, err := embeddedBinariesFS.ReadFile(embeddedPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", embeddedPath, err)
	}

	// Write to destination
	if err := os.WriteFile(destPath, data, 0755); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// hasEmbeddedBinaries checks if binaries are embedded in this build
func hasEmbeddedBinaries() bool {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	binaryName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binaryName = "ffmpeg.exe"
	}

	// Try to read directory first
	entries, err := embeddedBinariesFS.ReadDir(fmt.Sprintf("binaries/%s", platform))
	if err != nil {
		return false
	}

	// Check if we have the binary
	for _, entry := range entries {
		if entry.Name() == binaryName {
			return true
		}
	}

	return false
}
