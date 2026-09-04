package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var audioExtensions = []string{
	".aa", ".aac", ".aax", ".ac3", ".aif", ".aiff", ".ape", ".au", ".caf",
	".dsf", ".dts", ".flac", ".m4a", ".m4b", ".mka", ".mp3", ".mpc", ".oga",
	".ogg", ".oma", ".opus", ".shn", ".tak", ".tta", ".voc", ".w64", ".wav",
	".webm", ".wma", ".wv", ".xwma",
}

// Metadata represents audio file metadata
type Metadata struct {
	Format struct {
		Tags map[string]string `json:"tags"`
	} `json:"format"`
	Streams []struct {
		CodecType string            `json:"codec_type"`
		Tags      map[string]string `json:"tags"`
	} `json:"streams"`
}

func runConversion(config Config, dryRun bool) error {
	config.Codec = normalizeCodec(config.Codec)
	if config.Codec == "" && config.IPod {
		config.Codec = "aac"
	}
	if err := validateConfig(config); err != nil {
		return err
	}
	if dryRun {
		fmt.Println("=== DRY RUN MODE - No files will be converted ===")
	}

	// Verify input directory exists
	info, err := os.Stat(config.InputDir)
	if err != nil {
		return fmt.Errorf("cannot read input directory %s: %w", config.InputDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("input is not a directory: %s", config.InputDir)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	// Resolve aliases before excluding the destination from the source walk.
	config.InputDir, err = filepath.EvalSymlinks(config.InputDir)
	if err != nil {
		return err
	}
	config.InputDir, err = filepath.Abs(config.InputDir)
	if err != nil {
		return err
	}
	config.OutputDir, err = filepath.EvalSymlinks(config.OutputDir)
	if err != nil {
		return err
	}
	config.OutputDir, err = filepath.Abs(config.OutputDir)
	if err != nil {
		return err
	}
	outputInfo, err := os.Stat(config.OutputDir)
	if err != nil {
		return err
	}
	if os.SameFile(info, outputInfo) {
		return fmt.Errorf("input and output must be different directories")
	}

	// Collect all audio files
	files, err := collectAudioFiles(config.InputDir, config.OutputDir)
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

func collectAudioFiles(rootDir string, excludedDirs ...string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			for _, excluded := range excludedDirs {
				if path == excluded {
					return filepath.SkipDir
				}
			}
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
	// Fail the plan before starting any writes if two sources share a name.
	destinations := make(map[string]string, len(files))
	for _, file := range files {
		output, err := conversionOutputPath(file, config)
		if err != nil {
			return err
		}
		if previous, exists := destinations[output]; exists {
			return fmt.Errorf("output collision: %s and %s both convert to %s; rename one source or use separate output directories", previous, file, output)
		}
		destinations[output] = file
	}
	numWorkers := min(runtime.NumCPU(), len(files))
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
	var errs []error
	for err := range errorChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		fmt.Printf("\n%d files failed to process\n", len(errs))
		for _, err := range errs {
			fmt.Printf("  - %v\n", err)
		}
		return errors.Join(errs...)
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

	outputPath, err := conversionOutputPath(inputPath, config)
	if err != nil {
		return err
	}

	// Check if output already exists (resumability)
	if info, err := os.Lstat(outputPath); err == nil {
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("output exists but is not a completed audio file: %s; move or remove it before retrying", outputPath)
		}
		fmt.Printf("✓ Skipping (exists): %s\n", relPath)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot inspect output %s: %w", outputPath, err)
	}

	metadata := &Metadata{}
	if !dryRun {
		metadata, err = extractMetadata(inputPath)
		if err != nil {
			return fmt.Errorf("failed to extract metadata from %s: %w", inputPath, err)
		}
	}

	if dryRun {
		args := buildFFmpegArgs(inputPath, outputPath, config, metadata)
		fmt.Printf("[DRY RUN] %s -> %s\n", inputPath, outputPath)
		fmt.Printf("  FFmpeg args: %s\n\n", strings.Join(args, " "))
		return nil
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Keep partial work out of the final path, on the same filesystem. The
	// .partial suffix also keeps interrupted work out of future audio scans.
	temp, err := os.CreateTemp(filepath.Dir(outputPath), ".podhnologic-*.partial")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Close(); err != nil {
		return err
	}
	args := buildFFmpegArgs(inputPath, tempPath, config, metadata)
	format := strings.TrimPrefix(getOutputExtension(config.Codec), ".")
	if format == "m4a" {
		format = "ipod"
	}
	args = append([]string{"-nostdin", "-y"}, args[:len(args)-1]...)
	args = append(args, "-f", format, tempPath)

	// Only the private temporary file may be overwritten by FFmpeg.
	fmt.Printf("Converting: %s\n", relPath)

	output, err := runFFmpeg(args)
	if err != nil {
		return fmt.Errorf("conversion failed for %s: %w\nFFmpeg output: %s", inputPath, err, string(output))
	}
	if err := publishOutput(tempPath, outputPath); err != nil {
		return fmt.Errorf("publish output %s without overwriting an existing file: %w", outputPath, err)
	}

	fmt.Printf("✓ Completed: %s\n", relPath)

	return nil
}

func conversionOutputPath(inputPath string, config Config) (string, error) {
	relPath, err := filepath.Rel(config.InputDir, inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}
	output := filepath.Join(config.OutputDir, relPath)
	return strings.TrimSuffix(output, filepath.Ext(output)) + getOutputExtension(config.Codec), nil
}

func extractMetadata(filePath string) (*Metadata, error) {
	return probeMetadata(filePath)
}

func buildFFmpegArgs(inputPath, outputPath string, config Config, metadata *Metadata) []string {
	args := []string{
		"-i", inputPath,
		"-map", "0",
		"-map_metadata", "-1",
	}

	// Normalize tags to lowercase for case-insensitive lookup
	normalizedTags := make(map[string]string)
	for _, stream := range metadata.Streams {
		if stream.CodecType == "audio" {
			for key, value := range stream.Tags {
				normalizedTags[strings.ToLower(key)] = value
			}
			break
		}
	}
	// Preserve the existing preference for container tags when both are set.
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

	args = append(args, getCodecParamsSimple(config)...)

	// Add output path
	args = append(args, outputPath)

	return args
}

func getCodecParamsSimple(config Config) []string {
	var params []string

	// Use aac_at for macOS (best quality), fallback to aac for other platforms
	aacCodec := "aac"
	if runtime.GOOS == "darwin" {
		aacCodec = "aac_at"
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
		params = []string{"-c:a", "libmp3lame", "-q:a", "0", "-c:v", "copy"}

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
