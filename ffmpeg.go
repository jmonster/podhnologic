package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/ulikunitz/xz"
)

func downloadAndExtractFFmpeg(url, binDir string) error {
	// Create bin directory
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	// Download the archive
	fmt.Printf("Downloading ffmpeg from %s...\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "ffmpeg-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Download with progress bar
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)

	if _, err := io.Copy(io.MultiWriter(tmpFile, bar), resp.Body); err != nil {
		return err
	}

	// Close the file before extracting
	tmpFile.Close()

	fmt.Println("\nExtracting ffmpeg...")

	// Extract based on file extension
	if strings.HasSuffix(url, ".tar.xz") {
		return extractTarXz(tmpFile.Name(), binDir)
	} else if strings.HasSuffix(url, ".tar.gz") {
		return extractTarGz(tmpFile.Name(), binDir)
	} else if strings.HasSuffix(url, ".zip") {
		return extractZip(tmpFile.Name(), binDir)
	}

	return fmt.Errorf("unsupported archive format")
}

func extractTarXz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create xz reader
	xzReader, err := xz.NewReader(file)
	if err != nil {
		return err
	}

	// Create tar reader
	tarReader := tar.NewReader(xzReader)

	return extractTarFiles(tarReader, destDir)
}

func extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	return extractTarFiles(tarReader, destDir)
}

func extractTarFiles(tarReader *tar.Reader, destDir string) error {
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// We only want the ffmpeg and ffprobe binaries
		baseName := filepath.Base(header.Name)
		if baseName == "ffmpeg" || baseName == "ffprobe" {
			targetPath := filepath.Join(destDir, baseName)

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			fmt.Printf("Extracted: %s\n", baseName)
		}
	}

	return nil
}

func extractZip(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		// We only want the ffmpeg.exe and ffprobe.exe binaries
		baseName := filepath.Base(file.Name)
		if baseName == "ffmpeg.exe" || baseName == "ffprobe.exe" {
			targetPath := filepath.Join(destDir, baseName)

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}

			rc, err := file.Open()
			if err != nil {
				outFile.Close()
				return err
			}

			if _, err := io.Copy(outFile, rc); err != nil {
				outFile.Close()
				rc.Close()
				return err
			}

			outFile.Close()
			rc.Close()

			fmt.Printf("Extracted: %s\n", baseName)
		}
	}

	return nil
}

func getFFprobePath(ffmpegPath string) string {
	dir := filepath.Dir(ffmpegPath)
	baseName := "ffprobe"
	if runtime.GOOS == "windows" {
		baseName += ".exe"
	}
	return filepath.Join(dir, baseName)
}
