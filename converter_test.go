package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestHelper provides utilities for testing
type TestHelper struct {
	t         *testing.T
	tempDir   string
	inputDir  string
	outputDir string
	ffmpegPath string
}

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

// Setup prepares the test environment
func (h *TestHelper) Setup() {
	// Create input and output directories
	if err := os.MkdirAll(h.inputDir, 0755); err != nil {
		h.t.Fatalf("Failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(h.outputDir, 0755); err != nil {
		h.t.Fatalf("Failed to create output dir: %v", err)
	}

	// Find ffmpeg in PATH or use system ffmpeg
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		h.t.Skip("ffmpeg not found in PATH, skipping integration tests")
	}
	h.ffmpegPath = ffmpegPath
}

// GenerateTestAudio creates a test audio file using ffmpeg
func (h *TestHelper) GenerateTestAudio(filename string, duration int) string {
	outputPath := filepath.Join(h.inputDir, filename)

	durationStr := strconv.Itoa(duration)

	// Generate a simple sine wave audio file
	// -f lavfi: use libavfilter virtual input
	// sine=frequency=440:duration=X: generate 440Hz sine wave for X seconds
	cmd := exec.Command(h.ffmpegPath,
		"-f", "lavfi",
		"-i", "sine=frequency=440:duration="+durationStr,
		"-t", durationStr,
		"-y", // overwrite
		outputPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		h.t.Fatalf("Failed to generate test audio %s: %v\nOutput: %s", filename, err, output)
	}

	return outputPath
}

// GenerateTestAudioWithMetadata creates a test audio file with specific metadata
func (h *TestHelper) GenerateTestAudioWithMetadata(filename string, metadata map[string]string) string {
	outputPath := filepath.Join(h.inputDir, filename)

	// Build ffmpeg command with metadata
	args := []string{
		"-f", "lavfi",
		"-i", "sine=frequency=440:duration=2",
		"-t", "2",
	}

	// Add metadata arguments
	for key, value := range metadata {
		args = append(args, "-metadata", key+"="+value)
	}

	args = append(args, "-y", outputPath)

	cmd := exec.Command(h.ffmpegPath, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		h.t.Fatalf("Failed to generate test audio with metadata: %v\nOutput: %s", err, output)
	}

	return outputPath
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

// GetMetadata extracts metadata from an audio file
func (h *TestHelper) GetMetadata(filePath string) *Metadata {
	metadata, err := extractMetadata(filePath, h.ffmpegPath)
	if err != nil {
		h.t.Fatalf("Failed to extract metadata from %s: %v", filePath, err)
	}
	return metadata
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

// TestProcessFileBasic tests basic file processing
func TestProcessFileBasic(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	// Generate a test audio file
	inputFile := helper.GenerateTestAudio("test.mp3", 1)

	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "wav",
		IPod:      false,
		NoLyrics:  false,
	}

	// Process the file
	err := processFile(inputFile, config, helper.ffmpegPath, false)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify output file exists
	expectedOutput := filepath.Join(helper.outputDir, "test.wav")
	helper.VerifyFileExists(expectedOutput)
}

// TestProcessFileDryRun tests dry run mode
func TestProcessFileDryRun(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	// Generate a test audio file
	inputFile := helper.GenerateTestAudio("test.mp3", 1)

	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "flac",
		IPod:      false,
		NoLyrics:  false,
	}

	// Process the file in dry-run mode
	err := processFile(inputFile, config, helper.ffmpegPath, true)
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

	// Generate a test audio file
	inputFile := helper.GenerateTestAudio("test.mp3", 1)

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
	err := processFile(inputFile, config, helper.ffmpegPath, false)
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

// TestCodecConversions tests conversion to different codecs
func TestCodecConversions(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	codecs := []struct {
		name string
		ext  string
	}{
		{"flac", ".flac"},
		{"wav", ".wav"},
		{"mp3", ".mp3"},
		{"opus", ".opus"},
	}

	// Only test AAC/ALAC on platforms where they're well supported
	if runtime.GOOS == "darwin" {
		codecs = append(codecs,
			struct {
				name string
				ext  string
			}{"aac", ".m4a"},
			struct {
				name string
				ext  string
			}{"alac", ".m4a"},
		)
	}

	for _, codec := range codecs {
		t.Run(codec.name, func(t *testing.T) {
			// Generate input file
			inputFile := helper.GenerateTestAudio("input.mp3", 1)

			config := Config{
				InputDir:  helper.inputDir,
				OutputDir: filepath.Join(helper.outputDir, codec.name),
				Codec:     codec.name,
				IPod:      false,
				NoLyrics:  false,
			}

			// Create output directory
			if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
				t.Fatalf("Failed to create output dir: %v", err)
			}

			// Process the file
			err := processFile(inputFile, config, helper.ffmpegPath, false)
			if err != nil {
				t.Fatalf("processFile failed for %s: %v", codec.name, err)
			}

			// Verify output file exists with correct extension
			expectedOutput := filepath.Join(config.OutputDir, "input"+codec.ext)
			helper.VerifyFileExists(expectedOutput)

			// Verify file is not empty
			info, err := os.Stat(expectedOutput)
			if err != nil {
				t.Fatalf("Failed to stat output file: %v", err)
			}
			if info.Size() == 0 {
				t.Errorf("Output file is empty for codec %s", codec.name)
			}
		})
	}
}

// TestMetadataPreservation tests that metadata is preserved
func TestMetadataPreservation(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	// Generate file with metadata
	testMetadata := map[string]string{
		"title":  "Test Song",
		"artist": "Test Artist",
		"album":  "Test Album",
		"date":   "2024",
		"genre":  "Test Genre",
	}

	inputFile := helper.GenerateTestAudioWithMetadata("song.mp3", testMetadata)

	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "flac",
		IPod:      false,
		NoLyrics:  false,
	}

	// Process the file
	err := processFile(inputFile, config, helper.ffmpegPath, false)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify output exists
	outputFile := filepath.Join(helper.outputDir, "song.flac")
	helper.VerifyFileExists(outputFile)

	// Extract and verify metadata
	metadata := helper.GetMetadata(outputFile)

	for key := range testMetadata {
		helper.VerifyMetadataHasKey(metadata, key)
	}
}

// TestLyricsStripping tests that lyrics are stripped when NoLyrics is true
func TestLyricsStripping(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	// Generate file with lyrics
	testMetadata := map[string]string{
		"title":  "Test Song",
		"artist": "Test Artist",
		"lyrics": "These are test lyrics\nLine 2\nLine 3",
	}

	inputFile := helper.GenerateTestAudioWithMetadata("song_with_lyrics.mp3", testMetadata)

	// Test with lyrics stripping enabled
	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "flac",
		IPod:      false,
		NoLyrics:  true,
	}

	// Process the file
	err := processFile(inputFile, config, helper.ffmpegPath, false)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify output exists
	outputFile := filepath.Join(helper.outputDir, "song_with_lyrics.flac")
	helper.VerifyFileExists(outputFile)

	// Extract and verify metadata
	metadata := helper.GetMetadata(outputFile)

	// Should have title and artist
	helper.VerifyMetadataHasKey(metadata, "title")
	helper.VerifyMetadataHasKey(metadata, "artist")

	// Should NOT have lyrics
	helper.VerifyMetadataLacksKey(metadata, "lyrics")
}

// TestLyricsPreservation tests that lyrics are kept when NoLyrics is false
func TestLyricsPreservation(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	// Generate file with lyrics
	testMetadata := map[string]string{
		"title":  "Test Song",
		"lyrics": "These are test lyrics",
	}

	inputFile := helper.GenerateTestAudioWithMetadata("song_keep_lyrics.mp3", testMetadata)

	// Test with lyrics preservation
	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "flac",
		IPod:      false,
		NoLyrics:  false,
	}

	// Process the file
	err := processFile(inputFile, config, helper.ffmpegPath, false)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify output exists
	outputFile := filepath.Join(helper.outputDir, "song_keep_lyrics.flac")
	helper.VerifyFileExists(outputFile)

	// Extract and verify metadata
	metadata := helper.GetMetadata(outputFile)

	// Should have lyrics
	helper.VerifyMetadataHasKey(metadata, "lyrics")
}

// TestIPodMode tests iPod-specific encoding parameters
func TestIPodMode(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("iPod mode tests require macOS for AAC/ALAC support")
	}

	helper := NewTestHelper(t)
	helper.Setup()

	// Generate input file
	inputFile := helper.GenerateTestAudio("ipod_test.mp3", 1)

	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "aac",
		IPod:      true,
		NoLyrics:  false,
	}

	// Process the file
	err := processFile(inputFile, config, helper.ffmpegPath, false)
	if err != nil {
		t.Fatalf("processFile with iPod mode failed: %v", err)
	}

	// Verify output file exists
	expectedOutput := filepath.Join(helper.outputDir, "ipod_test.m4a")
	helper.VerifyFileExists(expectedOutput)

	// Verify file properties using ffprobe
	metadata := helper.GetMetadata(expectedOutput)

	// Should have an audio stream
	hasAudioStream := false
	for _, stream := range metadata.Streams {
		if stream.CodecType == "audio" {
			hasAudioStream = true
			break
		}
	}
	if !hasAudioStream {
		t.Error("iPod mode output missing audio stream")
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
	err := runConversion(config, helper.ffmpegPath, false)
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
	err := runConversion(config, helper.ffmpegPath, false)
	if err == nil {
		t.Error("runConversion should error on non-existent input directory")
	}
}

// TestExtractMetadata tests metadata extraction
func TestExtractMetadata(t *testing.T) {
	helper := NewTestHelper(t)
	helper.Setup()

	// Generate a test file with metadata
	testMetadata := map[string]string{
		"title":  "Extract Test",
		"artist": "Test Artist",
	}
	inputFile := helper.GenerateTestAudioWithMetadata("extract.mp3", testMetadata)

	// Extract metadata
	metadata, err := extractMetadata(inputFile, helper.ffmpegPath)
	if err != nil {
		t.Fatalf("extractMetadata failed: %v", err)
	}

	// Verify metadata structure
	if metadata == nil {
		t.Fatal("metadata is nil")
	}

	// Check for expected tags (case-insensitive)
	found := false
	for key := range metadata.Format.Tags {
		if strings.ToLower(key) == "title" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'title' in metadata tags")
	}
}

// BenchmarkProcessFile benchmarks file processing
func BenchmarkProcessFile(b *testing.B) {
	helper := &TestHelper{
		t:       &testing.T{},
		tempDir: b.TempDir(),
	}
	helper.inputDir = filepath.Join(helper.tempDir, "input")
	helper.outputDir = filepath.Join(helper.tempDir, "output")
	helper.Setup()

	// Generate a test audio file
	inputFile := helper.GenerateTestAudio("bench.mp3", 1)

	config := Config{
		InputDir:  helper.inputDir,
		OutputDir: helper.outputDir,
		Codec:     "flac",
		IPod:      false,
		NoLyrics:  false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Change output dir each iteration to avoid skip logic
		config.OutputDir = filepath.Join(helper.outputDir, strconv.Itoa(i))
		os.MkdirAll(config.OutputDir, 0755)
		processFile(inputFile, config, helper.ffmpegPath, false)
	}
}
