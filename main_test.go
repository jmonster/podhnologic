package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExpandPath tests the path expansion functionality
func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expand tilde",
			input:    "~/Documents",
			expected: filepath.Join(homeDir, "Documents"),
		},
		{
			name:     "expand lone tilde",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "no expansion needed",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestShortenPath tests the path shortening functionality
func TestShortenPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "shorten home path",
			input:    filepath.Join(homeDir, "Documents"),
			expected: "~/Documents",
		},
		{
			name:     "shorten exact home",
			input:    homeDir,
			expected: "~",
		},
		{
			name:     "non-home path unchanged",
			input:    "/var/log",
			expected: "/var/log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shortenPath(tt.input)
			if result != tt.expected {
				t.Errorf("shortenPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestTrimQuotes tests quote trimming
func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double quotes",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "single quotes",
			input:    "'hello'",
			expected: "hello",
		},
		{
			name:     "no quotes",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "mismatched quotes",
			input:    `"hello'`,
			expected: `"hello'`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "just quotes",
			input:    `""`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("trimQuotes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFindIndex tests the index finding utility
func TestFindIndex(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}

	tests := []struct {
		name     string
		target   string
		expected int
	}{
		{
			name:     "found at beginning",
			target:   "apple",
			expected: 0,
		},
		{
			name:     "found in middle",
			target:   "banana",
			expected: 1,
		},
		{
			name:     "found at end",
			target:   "cherry",
			expected: 2,
		},
		{
			name:     "not found returns 0",
			target:   "orange",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findIndex(items, tt.target)
			if result != tt.expected {
				t.Errorf("findIndex(items, %q) = %d, want %d", tt.target, result, tt.expected)
			}
		})
	}
}

// TestConfigSaveAndLoad tests config persistence
func TestConfigSaveAndLoad(t *testing.T) {
	// Create a temporary directory for config
	tempDir := t.TempDir()

	testConfig := Config{
		InputDir:  "/test/input",
		OutputDir: "/test/output",
		Codec:     "flac",
		IPod:      true,
		NoLyrics:  false,
	}

	// Save config
	err := saveConfig(tempDir, testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tempDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loadedConfig := loadConfig(tempDir)

	// Verify loaded config matches
	if loadedConfig.InputDir != testConfig.InputDir {
		t.Errorf("InputDir = %q, want %q", loadedConfig.InputDir, testConfig.InputDir)
	}
	if loadedConfig.OutputDir != testConfig.OutputDir {
		t.Errorf("OutputDir = %q, want %q", loadedConfig.OutputDir, testConfig.OutputDir)
	}
	if loadedConfig.Codec != testConfig.Codec {
		t.Errorf("Codec = %q, want %q", loadedConfig.Codec, testConfig.Codec)
	}
	if loadedConfig.IPod != testConfig.IPod {
		t.Errorf("IPod = %v, want %v", loadedConfig.IPod, testConfig.IPod)
	}
	if loadedConfig.NoLyrics != testConfig.NoLyrics {
		t.Errorf("NoLyrics = %v, want %v", loadedConfig.NoLyrics, testConfig.NoLyrics)
	}

	// Test that the config is valid JSON
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var jsonConfig Config
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		t.Fatalf("Config is not valid JSON: %v", err)
	}
}

// TestLoadConfigNonExistent tests loading a config that doesn't exist
func TestLoadConfigNonExistent(t *testing.T) {
	tempDir := t.TempDir()

	// Load from empty directory
	config := loadConfig(tempDir)

	// Should return empty config
	if config.InputDir != "" || config.OutputDir != "" || config.Codec != "" {
		t.Error("Loading non-existent config should return empty Config")
	}
}

// TestLoadConfigInvalid tests loading an invalid config file
func TestLoadConfigInvalid(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Write invalid JSON
	if err := os.WriteFile(configPath, []byte("invalid json{{{"), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Should return empty config on error
	config := loadConfig(tempDir)
	if config.InputDir != "" || config.OutputDir != "" || config.Codec != "" {
		t.Error("Loading invalid config should return empty Config")
	}
}

// TestGetOutputExtension tests output extension mapping
func TestGetOutputExtension(t *testing.T) {
	tests := []struct {
		codec    string
		expected string
	}{
		{"alac", ".m4a"},
		{"aac", ".m4a"},
		{"flac", ".flac"},
		{"mp3", ".mp3"},
		{"opus", ".opus"},
		{"wav", ".wav"},
		{"unknown", ".m4a"}, // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.codec, func(t *testing.T) {
			result := getOutputExtension(tt.codec)
			if result != tt.expected {
				t.Errorf("getOutputExtension(%q) = %q, want %q", tt.codec, result, tt.expected)
			}
		})
	}
}

// TestIsAudioFile tests audio file detection
func TestIsAudioFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"song.mp3", true},
		{"song.MP3", true}, // case insensitive
		{"song.wav", true},
		{"song.flac", true},
		{"song.aac", true},
		{"song.opus", true},
		{"song.m4a", true},
		{"song.ogg", true},
		{"document.txt", false},
		{"image.jpg", false},
		{"video.mp4", false},
		{"noextension", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isAudioFile(tt.path)
			if result != tt.expected {
				t.Errorf("isAudioFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestGetFFprobePath tests ffprobe path derivation
func TestGetFFprobePath(t *testing.T) {
	tests := []struct {
		name        string
		ffmpegPath  string
		expectEnds  string
	}{
		{
			name:       "unix path",
			ffmpegPath: "/usr/local/bin/ffmpeg",
			expectEnds: "ffprobe",
		},
	}

	// Platform-specific test
	if strings.Contains(os.Getenv("OS"), "Windows") {
		tests = append(tests, struct {
			name        string
			ffmpegPath  string
			expectEnds  string
		}{
			name:       "windows path",
			ffmpegPath: "C:\\Program Files\\ffmpeg\\ffmpeg.exe",
			expectEnds: "ffprobe.exe",
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFFprobePath(tt.ffmpegPath)
			if !strings.HasSuffix(result, tt.expectEnds) {
				t.Errorf("getFFprobePath(%q) = %q, should end with %q", tt.ffmpegPath, result, tt.expectEnds)
			}
			// Verify it's in the same directory
			if filepath.Dir(result) != filepath.Dir(tt.ffmpegPath) {
				t.Errorf("getFFprobePath should be in same directory as ffmpeg")
			}
		})
	}
}
