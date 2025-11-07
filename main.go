package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/manifoldco/promptui"
)

const (
	Version = "3.0.0"
)

// Config represents the user's saved configuration
type Config struct {
	InputDir   string `json:"input_dir"`
	OutputDir  string `json:"output_dir"`
	Codec      string `json:"codec"`
	IPod       bool   `json:"ipod"`
	NoLyrics   bool   `json:"no_lyrics"`
	FFmpegPath string `json:"ffmpeg_path,omitempty"`
}

var (
	// Command-line flags
	inputFlag   = flag.String("input", "", "Input directory containing audio files")
	outputFlag  = flag.String("output", "", "Output directory for converted files")
	codecFlag   = flag.String("codec", "", "Target codec: flac, alac, aac, wav, mp3, opus")
	ipodFlag    = flag.Bool("ipod", false, "Enable iPod optimizations")
	noLyricsFlag = flag.Bool("no-lyrics", false, "Strip lyrics metadata")
	ffmpegFlag  = flag.String("ffmpeg", "", "Path to ffmpeg binary")
	dryRunFlag  = flag.Bool("dry-run", false, "Show what would be done without converting")
	interactiveFlag = flag.Bool("interactive", false, "Force interactive mode")
	versionFlag = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("podhnologic v%s\n", Version)
		os.Exit(0)
	}

	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		log.Fatalf("Failed to get config directory: %v", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	// Load existing config
	config := loadConfig(configDir)

	// Determine if we should run in interactive mode
	interactive := *interactiveFlag || (flag.NFlag() == 0 && len(os.Args) == 1)

	if interactive {
		// Interactive mode
		if err := runInteractive(&config, configDir); err != nil {
			log.Fatalf("Interactive mode failed: %v", err)
		}
	} else {
		// Command-line mode: override config with flags
		if *inputFlag != "" {
			config.InputDir = *inputFlag
		}
		if *outputFlag != "" {
			config.OutputDir = *outputFlag
		}
		if *codecFlag != "" {
			config.Codec = *codecFlag
		}
		if *ipodFlag {
			config.IPod = true
		}
		if *noLyricsFlag {
			config.NoLyrics = true
		}
		if *ffmpegFlag != "" {
			config.FFmpegPath = *ffmpegFlag
		}

		// Validate required fields
		if config.InputDir == "" || config.OutputDir == "" {
			log.Fatal("--input and --output are required")
		}
		if config.Codec == "" && !config.IPod {
			log.Fatal("--codec or --ipod is required")
		}

		// Save the config for future use
		saveConfig(configDir, config)
	}

	// Ensure we have ffmpeg
	ffmpegPath, err := ensureFFmpeg(configDir, config.FFmpegPath)
	if err != nil {
		log.Fatalf("Failed to ensure ffmpeg is available: %v", err)
	}
	config.FFmpegPath = ffmpegPath

	// Set default codec for iPod mode
	if config.Codec == "" && config.IPod {
		config.Codec = "aac"
	}

	// Run the conversion
	if err := runConversion(config, *dryRunFlag); err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".podhnologic"), nil
}

func loadConfig(configDir string) Config {
	configPath := filepath.Join(configDir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config doesn't exist yet, return empty config
		return Config{}
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Warning: Failed to parse config file: %v", err)
		return Config{}
	}

	return config
}

func saveConfig(configDir string, config Config) error {
	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func runInteractive(config *Config, configDir string) error {
	fmt.Println("=== podhnologic - Interactive Setup ===\n")

	// Input directory
	inputPrompt := promptui.Prompt{
		Label:   "Input directory",
		Default: config.InputDir,
	}
	inputDir, err := inputPrompt.Run()
	if err != nil {
		return err
	}
	config.InputDir = strings.TrimSpace(inputDir)

	// Output directory
	outputPrompt := promptui.Prompt{
		Label:   "Output directory",
		Default: config.OutputDir,
	}
	outputDir, err := outputPrompt.Run()
	if err != nil {
		return err
	}
	config.OutputDir = strings.TrimSpace(outputDir)

	// Codec selection
	codecPrompt := promptui.Select{
		Label: "Select target codec",
		Items: []string{"aac", "alac", "flac", "mp3", "opus", "wav"},
		CursorPos: findIndex([]string{"aac", "alac", "flac", "mp3", "opus", "wav"}, config.Codec),
	}
	_, codec, err := codecPrompt.Run()
	if err != nil {
		return err
	}
	config.Codec = codec

	// iPod optimization
	ipodPrompt := promptui.Select{
		Label: "Enable iPod optimizations?",
		Items: []string{"Yes", "No"},
	}
	if config.IPod {
		ipodPrompt.CursorPos = 0
	} else {
		ipodPrompt.CursorPos = 1
	}
	_, ipodChoice, err := ipodPrompt.Run()
	if err != nil {
		return err
	}
	config.IPod = (ipodChoice == "Yes")

	// No lyrics option
	noLyricsPrompt := promptui.Select{
		Label: "Strip lyrics from metadata?",
		Items: []string{"No", "Yes"},
	}
	if config.NoLyrics {
		noLyricsPrompt.CursorPos = 1
	} else {
		noLyricsPrompt.CursorPos = 0
	}
	_, noLyricsChoice, err := noLyricsPrompt.Run()
	if err != nil {
		return err
	}
	config.NoLyrics = (noLyricsChoice == "Yes")

	// Custom ffmpeg path (optional)
	customFFmpegPrompt := promptui.Select{
		Label: "Use custom ffmpeg path?",
		Items: []string{"No (auto-detect/download)", "Yes (specify path)"},
	}
	_, customFFmpeg, err := customFFmpegPrompt.Run()
	if err != nil {
		return err
	}

	if customFFmpeg == "Yes (specify path)" {
		ffmpegPathPrompt := promptui.Prompt{
			Label:   "FFmpeg path",
			Default: config.FFmpegPath,
		}
		ffmpegPath, err := ffmpegPathPrompt.Run()
		if err != nil {
			return err
		}
		config.FFmpegPath = strings.TrimSpace(ffmpegPath)
	} else {
		config.FFmpegPath = ""
	}

	// Save configuration
	if err := saveConfig(configDir, *config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("\n✓ Configuration saved to", filepath.Join(configDir, "config.json"))

	// Ask if they want to start conversion now
	startPrompt := promptui.Select{
		Label: "Start conversion now?",
		Items: []string{"Yes", "No"},
	}
	_, startChoice, err := startPrompt.Run()
	if err != nil {
		return err
	}

	if startChoice == "No" {
		fmt.Println("Configuration saved. Run 'podhnologic' again to start conversion.")
		os.Exit(0)
	}

	return nil
}

func findIndex(items []string, target string) int {
	for i, item := range items {
		if item == target {
			return i
		}
	}
	return 0
}

func getFFmpegDownloadURL() (string, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Map of platform/arch to download URLs
	// Using static builds from https://github.com/BtbN/FFmpeg-Builds (Linux/Windows)
	// and https://evermeet.cx/ffmpeg/ (macOS)

	switch os {
	case "linux":
		if arch == "amd64" {
			return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz", nil
		} else if arch == "arm64" {
			return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linuxarm64-gpl.tar.xz", nil
		}
	case "darwin":
		// For macOS, we'll use homebrew or prompt user to install
		// For now, return an error suggesting manual installation
		return "", fmt.Errorf("macOS: please install ffmpeg via 'brew install ffmpeg'")
	case "windows":
		if arch == "amd64" {
			return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip", nil
		}
	}

	return "", fmt.Errorf("unsupported platform: %s/%s", os, arch)
}

func ensureFFmpeg(configDir, customPath string) (string, error) {
	// If custom path is specified, verify it works
	if customPath != "" {
		if err := testFFmpeg(customPath); err != nil {
			return "", fmt.Errorf("custom ffmpeg path is invalid: %w", err)
		}
		return customPath, nil
	}

	// Try to find ffmpeg in PATH
	if path, err := findInPath("ffmpeg"); err == nil {
		if testFFmpeg(path) == nil {
			fmt.Printf("Found ffmpeg in PATH: %s\n", path)
			return path, nil
		}
	}

	// Check if we already downloaded it
	binDir := filepath.Join(configDir, "bin")
	localFFmpeg := filepath.Join(binDir, "ffmpeg")
	if runtime.GOOS == "windows" {
		localFFmpeg += ".exe"
	}

	if _, err := os.Stat(localFFmpeg); err == nil {
		if testFFmpeg(localFFmpeg) == nil {
			return localFFmpeg, nil
		}
	}

	// Need to download ffmpeg
	fmt.Println("FFmpeg not found. Attempting to download...")

	downloadURL, err := getFFmpegDownloadURL()
	if err != nil {
		return "", err
	}

	if err := downloadAndExtractFFmpeg(downloadURL, binDir); err != nil {
		return "", fmt.Errorf("failed to download ffmpeg: %w", err)
	}

	return localFFmpeg, nil
}

func testFFmpeg(path string) error {
	// Test if ffmpeg works by running -version
	cmd := execCommand(path, "-version")
	return cmd.Run()
}

func findInPath(binary string) (string, error) {
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	pathEnv := os.Getenv("PATH")
	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	for _, dir := range strings.Split(pathEnv, separator) {
		fullPath := filepath.Join(dir, binary)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("%s not found in PATH", binary)
}
