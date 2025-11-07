package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var audioExtensions = []string{".mp3", ".wav", ".flac", ".aac", ".opus", ".m4a", ".ogg"}

// Metadata represents audio file metadata
type Metadata struct {
	Format struct {
		Tags map[string]string `json:"tags"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
	} `json:"streams"`
}

func runConversion(config Config, dryRun bool) error {
	if dryRun {
		fmt.Println("=== DRY RUN MODE - No files will be converted ===")
	}

	// Verify input directory exists
	if _, err := os.Stat(config.InputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", config.InputDir)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Collect all audio files
	files, err := collectAudioFiles(config.InputDir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("No audio files found in input directory")
		return nil
	}

	fmt.Printf("Found %d audio files\n", len(files))
	fmt.Printf("Using %d threads\n\n", runtime.NumCPU())

	// Process files in parallel
	return processFilesParallel(files, config, dryRun)
}

func collectAudioFiles(rootDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isAudioFile(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func isAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, audioExt := range audioExtensions {
		if ext == audioExt {
			return true
		}
	}
	return false
}

func processFilesParallel(files []string, config Config, dryRun bool) error {
	numWorkers := runtime.NumCPU()
	fileChan := make(chan string, len(files))
	errorChan := make(chan error, len(files))

	// Fill the channel with files
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				if err := processFile(file, config, dryRun); err != nil {
					errorChan <- err
				}
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
	close(errorChan)

	// Check for errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		fmt.Printf("\n%d files failed to process\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	fmt.Println("\n✓ All tasks completed")

	return nil
}

func processFile(inputPath string, config Config, dryRun bool) error {
	// Get relative path from input dir
	relPath, err := filepath.Rel(config.InputDir, inputPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Determine output extension
	outputExt := getOutputExtension(config.Codec)

	// Build output path
	outputPath := filepath.Join(config.OutputDir, relPath)
	outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + outputExt

	// Check if output already exists (resumability)
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("✓ Skipping (exists): %s\n", relPath)
		return nil
	}

	// Extract metadata
	metadata, err := extractMetadata(inputPath, config.FFmpegPath)
	if err != nil {
		return fmt.Errorf("failed to extract metadata from %s: %w", inputPath, err)
	}

	// Build ffmpeg command
	args := buildFFmpegArgs(inputPath, outputPath, config, metadata)

	if dryRun {
		fmt.Printf("[DRY RUN] %s -> %s\n", inputPath, outputPath)
		fmt.Printf("  Command: %s %s\n\n", config.FFmpegPath, strings.Join(args, " "))
		return nil
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Run ffmpeg
	fmt.Printf("Converting: %s\n", relPath)

	cmd := execCommand(config.FFmpegPath, args...)

	// Capture stderr for error reporting
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Read stderr (but don't print unless there's an error)
	stderrData, _ := io.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("conversion failed for %s: %w\nFFmpeg output: %s", inputPath, err, string(stderrData))
	}

	fmt.Printf("✓ Completed: %s\n", relPath)

	return nil
}

func extractMetadata(filePath, ffmpegPath string) (*Metadata, error) {
	ffprobePath := getFFprobePath(ffmpegPath)

	cmd := execCommand(ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var metadata Metadata
	if err := json.Unmarshal(output, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func buildFFmpegArgs(inputPath, outputPath string, config Config, metadata *Metadata) []string {
	args := []string{
		"-i", inputPath,
		"-map", "0",
		"-map_metadata", "-1",
	}

	// Normalize tags to lowercase for case-insensitive lookup
	normalizedTags := make(map[string]string)
	for key, value := range metadata.Format.Tags {
		normalizedTags[strings.ToLower(key)] = value
	}

	// Add desired metadata
	desiredKeys := []string{"title", "artist", "album", "date", "track", "genre", "disc"}
	if !config.NoLyrics {
		desiredKeys = append(desiredKeys, "lyrics")
	}

	for _, key := range desiredKeys {
		if value, ok := normalizedTags[key]; ok {
			args = append(args, "-metadata", fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add codec-specific parameters
	args = append(args, getCodecParams(config)...)

	// Add output path
	args = append(args, outputPath)

	return args
}

func getCodecParams(config Config) []string {
	var params []string

	// Determine best AAC encoder
	aacCodec := "aac"
	if hasEncoder(config.FFmpegPath, "aac_at") {
		aacCodec = "aac_at"
	} else if hasEncoder(config.FFmpegPath, "libfdk_aac") {
		aacCodec = "libfdk_aac"
	}

	switch config.Codec {
	case "alac":
		params = []string{"-c:a", "alac", "-c:v", "copy"}
		if config.IPod {
			params = append(params, "-sample_fmt", "s16p", "-ar", "44100", "-movflags", "+faststart", "-disposition:a", "0")
		}

	case "aac":
		params = []string{"-c:a", aacCodec, "-b:a", "256k", "-c:v", "copy"}
		if config.IPod {
			params = append(params, "-ar", "44100", "-movflags", "+faststart", "-disposition:a", "0")
		}

	case "flac":
		params = []string{"-c:a", "flac", "-c:v", "copy"}

	case "mp3":
		params = []string{"-c:a", "libmp3lame", "-q:a", "0"}

	case "opus":
		params = []string{"-c:a", "libopus", "-b:a", "128k", "-vn"}

	case "wav":
		params = []string{"-c:a", "pcm_s16le", "-vn"}
	}

	return params
}

func getOutputExtension(codec string) string {
	extensions := map[string]string{
		"alac": ".m4a",
		"aac":  ".m4a",
		"flac": ".flac",
		"mp3":  ".mp3",
		"opus": ".opus",
		"wav":  ".wav",
	}

	if ext, ok := extensions[codec]; ok {
		return ext
	}

	return ".m4a" // default
}

var encoderCache = make(map[string]bool)
var encoderCacheMutex sync.Mutex

func hasEncoder(ffmpegPath, encoder string) bool {
	encoderCacheMutex.Lock()
	defer encoderCacheMutex.Unlock()

	if cached, ok := encoderCache[encoder]; ok {
		return cached
	}

	cmd := execCommand(ffmpegPath, "-h", fmt.Sprintf("encoder=%s", encoder))
	output, err := cmd.CombinedOutput()

	result := err == nil && strings.Contains(string(output), fmt.Sprintf("Encoder %s", encoder))
	encoderCache[encoder] = result

	return result
}

// execCommand is a helper to create exec.Command
// This is separate so we can mock it in tests if needed
func execCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
