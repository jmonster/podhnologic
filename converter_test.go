package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// NewTestHelper creates a new test helper with temporary directories
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir := t.TempDir()

	return &TestHelper{
		t:         t,
		tempDir:   tempDir,
		inputDir:  filepath.Join(tempDir, "input"),
		outputDir: filepath.Join(tempDir, "output"),
	}
}

// VerifyFileExists checks if a file exists
func (h *TestHelper) VerifyFileExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		h.t.Errorf("Expected file does not exist: %s", path)
	}
}

// VerifyFileNotExists checks if a file does not exist
func (h *TestHelper) VerifyFileNotExists(path string) {
	if _, err := os.Stat(path); err == nil {
		h.t.Errorf("File should not exist: %s", path)
	}
}

// VerifyMetadataHasKey checks if metadata contains a specific key
func (h *TestHelper) VerifyMetadataHasKey(metadata *Metadata, key string) {
	keyLower := strings.ToLower(key)
	for k := range metadata.Format.Tags {
		if strings.ToLower(k) == keyLower {
			return
		}
	}
	h.t.Errorf("Metadata missing expected key: %s", key)
}

// VerifyMetadataLacksKey checks if metadata does not contain a specific key
func (h *TestHelper) VerifyMetadataLacksKey(metadata *Metadata, key string) {
	keyLower := strings.ToLower(key)
	for k := range metadata.Format.Tags {
		if strings.ToLower(k) == keyLower {
			h.t.Errorf("Metadata should not have key: %s", key)
			return
		}
	}
}

// TestCollectAudioFiles tests audio file collection
func TestCollectAudioFiles(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	// Create some test files
	testFiles := []string{
		"song1.mp3",
		"song2.flac",
		"subdir/song3.wav",
		"document.txt", // should be ignored
		"image.jpg",    // should be ignored
	}

	for _, file := range testFiles {
		filePath := filepath.Join(helper.inputDir, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Collect audio files
	files, err := collectAudioFiles(helper.inputDir)
	if err != nil {
		t.Fatalf("collectAudioFiles failed: %v", err)
	}

	// Should find exactly 3 audio files
	if len(files) != 3 {
		t.Errorf("Expected 3 audio files, got %d", len(files))
	}

	// Verify all found files are audio files
	for _, file := range files {
		if !isAudioFile(file) {
			t.Errorf("Non-audio file collected: %s", file)
		}
	}
}

// TestProcessFileDryRun tests dry run mode
func TestProcessFileDryRun(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	inputFile := helper.WriteInputFile("test.mp3", []byte("dry run fixture"))

	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "flac",
		IPod:      false,
		NoLyrics:  false,
	}

	// Process the file in dry-run mode
	err := processFile(inputFile, config, true)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify output file does NOT exist
	expectedOutput := filepath.Join(helper.outputDir, "test.flac")
	helper.VerifyFileNotExists(expectedOutput)
}

// TestProcessFileSkipsExisting tests that existing files are skipped
func TestProcessFileSkipsExisting(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	inputFile := helper.WriteInputFile("test.mp3", []byte("skip fixture"))

	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "wav",
		IPod:      false,
		NoLyrics:  false,
	}

	// Create the output file manually
	expectedOutput := filepath.Join(helper.outputDir, "test.wav")
	if err := os.WriteFile(expectedOutput, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Process the file - should skip
	err := processFile(inputFile, config, false)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify the file still has original content (wasn't overwritten)
	content, err := os.ReadFile(expectedOutput)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if string(content) != "existing" {
		t.Error("Existing file was overwritten when it should have been skipped")
	}
}

// TestBuildFFmpegArgs tests FFmpeg argument construction
func TestBuildFFmpegArgs(t *testing.T) {
	metadata := &Metadata{
		Format: struct {
			Tags map[string]string `json:"tags"`
		}{
			Tags: map[string]string{
				"title":  "Test Title",
				"artist": "Test Artist",
				"lyrics": "Test Lyrics",
			},
		},
	}

	tests := []struct {
		name           string
		config         Config
		expectContains []string
		expectLacks    []string
	}{
		{
			name: "preserve lyrics",
			config: Config{
				Codec:    "flac",
				NoLyrics: false,
			},
			expectContains: []string{"-metadata", "lyrics=Test Lyrics"},
			expectLacks:    []string{},
		},
		{
			name: "strip lyrics",
			config: Config{
				Codec:    "flac",
				NoLyrics: true,
			},
			expectContains: []string{"-metadata", "title=Test Title"},
			expectLacks:    []string{"lyrics=Test Lyrics"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildFFmpegArgs("/input.mp3", "/output.flac", tt.config, metadata)
			argsStr := strings.Join(args, " ")

			for _, expected := range tt.expectContains {
				if !strings.Contains(argsStr, expected) {
					t.Errorf("Expected args to contain %q, got: %s", expected, argsStr)
				}
			}

			for _, unexpected := range tt.expectLacks {
				if strings.Contains(argsStr, unexpected) {
					t.Errorf("Expected args NOT to contain %q, got: %s", unexpected, argsStr)
				}
			}
		})
	}
}

func TestBuildFFmpegArgsPreservesConversionContract(t *testing.T) {
	metadata := &Metadata{
		Format: struct {
			Tags map[string]string `json:"tags"`
		}{
			Tags: map[string]string{
				"TITLE":       "Test Title",
				"artist":      "Test Artist",
				"album":       "Test Album",
				"date":        "2024",
				"track":       "7",
				"genre":       "Test Genre",
				"disc":        "2",
				"lyrics":      "Test Lyrics",
				"comment":     "Drop Me",
				"description": "Drop Me Too",
			},
		},
	}

	tests := []struct {
		name     string
		config   Config
		expected []string
	}{
		{
			name: "flac keeps mapped streams and whitelisted metadata",
			config: Config{
				Codec:    "flac",
				NoLyrics: false,
			},
			expected: []string{
				"-i", "/input.mp3",
				"-map", "0",
				"-map_metadata", "-1",
				"-metadata", "title=Test Title",
				"-metadata", "artist=Test Artist",
				"-metadata", "album=Test Album",
				"-metadata", "date=2024",
				"-metadata", "track=7",
				"-metadata", "genre=Test Genre",
				"-metadata", "disc=2",
				"-metadata", "lyrics=Test Lyrics",
				"-c:a", "flac",
				"-c:v", "copy",
				"/output.flac",
			},
		},
		{
			name: "ipod alac keeps compatibility flags",
			config: Config{
				Codec: "alac",
				IPod:  true,
			},
			expected: []string{
				"-i", "/input.mp3",
				"-map", "0",
				"-map_metadata", "-1",
				"-metadata", "title=Test Title",
				"-metadata", "artist=Test Artist",
				"-metadata", "album=Test Album",
				"-metadata", "date=2024",
				"-metadata", "track=7",
				"-metadata", "genre=Test Genre",
				"-metadata", "disc=2",
				"-metadata", "lyrics=Test Lyrics",
				"-c:a", "alac",
				"-c:v", "copy",
				"-sample_fmt", "s16p",
				"-ar", "44100",
				"-movflags", "+faststart",
				"-disposition:a", "0",
				"/output.m4a",
			},
		},
		{
			name: "wav strips attached streams and lyrics when requested",
			config: Config{
				Codec:    "wav",
				NoLyrics: true,
			},
			expected: []string{
				"-i", "/input.mp3",
				"-map", "0",
				"-map_metadata", "-1",
				"-metadata", "title=Test Title",
				"-metadata", "artist=Test Artist",
				"-metadata", "album=Test Album",
				"-metadata", "date=2024",
				"-metadata", "track=7",
				"-metadata", "genre=Test Genre",
				"-metadata", "disc=2",
				"-c:a", "pcm_s16le",
				"-vn",
				"/output.wav",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := tt.expected[len(tt.expected)-1]
			args := buildFFmpegArgs("/input.mp3", outputPath, tt.config, metadata)
			if strings.Join(args, "\x00") != strings.Join(tt.expected, "\x00") {
				t.Fatalf("unexpected ffmpeg args\n got: %q\nwant: %q", args, tt.expected)
			}

			argsStr := strings.Join(args, " ")
			for _, dropped := range []string{"comment=Drop Me", "description=Drop Me Too"} {
				if strings.Contains(argsStr, dropped) {
					t.Fatalf("ffmpeg args preserved non-whitelisted metadata %q: %v", dropped, args)
				}
			}
			if tt.config.NoLyrics && strings.Contains(argsStr, "lyrics=Test Lyrics") {
				t.Fatalf("ffmpeg args preserved lyrics while NoLyrics=true: %v", args)
			}
		})
	}
}

func TestProcessFilesParallelReturnsConversionErrors(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	inputFile := filepath.Join(inputDir, "broken.mp3")
	if err := os.WriteFile(inputFile, []byte("not audio"), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	config := Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Codec:     "flac",
	}

	err := processFilesParallel([]string{inputFile}, config, false)
	if err == nil {
		t.Fatal("processFilesParallel returned nil after conversion failure")
	}
	if !strings.Contains(err.Error(), "failed to extract metadata") && !strings.Contains(err.Error(), "linked ffmpeg bridge unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGetCodecParamsSimple tests codec parameter generation
func TestGetCodecParamsSimple(t *testing.T) {
	tests := []struct {
		codec          string
		ipod           bool
		expectContains []string
	}{
		{
			codec:          "flac",
			ipod:           false,
			expectContains: []string{"-c:a", "flac"},
		},
		{
			codec:          "wav",
			ipod:           false,
			expectContains: []string{"-c:a", "pcm_s16le"},
		},
		{
			codec:          "mp3",
			ipod:           false,
			expectContains: []string{"-c:a", "libmp3lame"},
		},
		{
			codec:          "opus",
			ipod:           false,
			expectContains: []string{"-c:a", "libopus"},
		},
		{
			codec:          "aac",
			ipod:           true,
			expectContains: []string{"-c:a", "-ar", "44100"},
		},
		{
			codec:          "alac",
			ipod:           true,
			expectContains: []string{"-c:a", "alac", "44100"},
		},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("%s_ipod=%v", tt.codec, tt.ipod)
		t.Run(testName, func(t *testing.T) {
			config := Config{
				Codec: tt.codec,
				IPod:  tt.ipod,
			}

			params := getCodecParamsSimple(config)
			paramsStr := strings.Join(params, " ")

			for _, expected := range tt.expectContains {
				if !strings.Contains(paramsStr, expected) {
					t.Errorf("Expected params to contain %q, got: %v", expected, params)
				}
			}
		})
	}
}

// TestRunConversionEmptyInput tests behavior with empty input directory
func TestRunConversionEmptyInput(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "flac",
	}

	// Run conversion on empty directory
	err := runConversion(config, false)
	if err != nil {
		t.Errorf("runConversion should not error on empty directory: %v", err)
	}
}

// TestRunConversionNonExistentInput tests behavior with non-existent input
func TestRunConversionNonExistentInput(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	config := Config{
		InputDir:  filepath.Join(helper.tempDir, "nonexistent"),
		OutputDir: helper.outputDir,
		Codec:     "flac",
	}

	// Run conversion on non-existent directory
	err := runConversion(config, false)
	if err == nil {
		t.Error("runConversion should error on non-existent input directory")
	}
}
