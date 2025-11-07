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
	Version = "4.0.0"
)

// Config represents the user's saved configuration
type Config struct {
	InputDir  string `json:"input_dir"`
	OutputDir string `json:"output_dir"`
	Codec     string `json:"codec"`
	IPod      bool   `json:"ipod"`
	NoLyrics  bool   `json:"no_lyrics"`
}

var (
	// Command-line flags
	inputFlag       = flag.String("input", "", "Input directory containing audio files")
	outputFlag      = flag.String("output", "", "Output directory for converted files")
	codecFlag       = flag.String("codec", "", "Target codec: flac, alac, aac, wav, mp3, opus")
	ipodFlag        = flag.Bool("ipod", false, "Enable iPod optimizations")
	noLyricsFlag    = flag.Bool("no-lyrics", false, "Strip lyrics metadata")
	dryRunFlag      = flag.Bool("dry-run", false, "Show what would be done without converting")
	interactiveFlag = flag.Bool("interactive", false, "Force interactive mode")
	versionFlag     = flag.Bool("version", false, "Show version information")
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
			config.InputDir = expandPath(*inputFlag)
		}
		if *outputFlag != "" {
			config.OutputDir = expandPath(*outputFlag)
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

	// Ensure we have ffmpeg (uses embedded binaries)
	ffmpegPath, err := ensureFFmpeg(configDir)
	if err != nil {
		log.Fatalf("Failed to ensure ffmpeg is available: %v", err)
	}

	// Set default codec for iPod mode
	if config.Codec == "" && config.IPod {
		config.Codec = "aac"
	}

	// Run the conversion
	if err := runConversion(config, ffmpegPath, *dryRunFlag); err != nil {
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
	printBanner()

	for {
		// Display current configuration
		fmt.Println()
		fmt.Printf("%s%sCurrent Configuration:%s\n\n", colorBold, colorCyan, colorReset)

		// Input directory
		inputDisplay := config.InputDir
		if inputDisplay == "" {
			inputDisplay = colorRed + "(not set)" + colorReset
		} else {
			inputDisplay = colorGreen + shortenPath(inputDisplay) + colorReset
		}
		fmt.Printf("  %s[I]%s Input Directory:  %s\n", colorYellow+colorBold, colorReset, inputDisplay)

		// Output directory
		outputDisplay := config.OutputDir
		if outputDisplay == "" {
			outputDisplay = colorRed + "(not set)" + colorReset
		} else {
			outputDisplay = colorGreen + shortenPath(outputDisplay) + colorReset
		}
		fmt.Printf("  %s[O]%s Output Directory: %s\n", colorYellow+colorBold, colorReset, outputDisplay)

		// Codec
		codecDisplay := config.Codec
		if codecDisplay == "" {
			codecDisplay = colorRed + "(not set)" + colorReset
		} else {
			codecDisplay = colorGreen + codecDisplay + colorReset
		}
		fmt.Printf("  %s[C]%s Codec:            %s\n", colorYellow+colorBold, colorReset, codecDisplay)

		// iPod mode
		ipodDisplay := "disabled"
		if config.IPod {
			ipodDisplay = colorGreen + "enabled" + colorReset
		} else {
			ipodDisplay = colorWhite + "disabled" + colorReset
		}
		fmt.Printf("  %s[P]%s iPod Mode:        %s\n", colorYellow+colorBold, colorReset, ipodDisplay)

		// Strip lyrics
		lyricsDisplay := "keep lyrics"
		if config.NoLyrics {
			lyricsDisplay = colorYellow + "strip lyrics" + colorReset
		} else {
			lyricsDisplay = colorWhite + "keep lyrics" + colorReset
		}
		fmt.Printf("  %s[L]%s Lyrics:           %s\n", colorYellow+colorBold, colorReset, lyricsDisplay)

		fmt.Println()
		fmt.Printf("%s%s[Enter]%s Start Conversion  %s[Q]%s Quit\n",
			colorGreen+colorBold, colorReset, colorReset,
			colorRed+colorBold, colorReset)
		fmt.Println()
		fmt.Print("Select option: ")

		// Read single key
		var input string
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "i":
			dir, err := RunBubbleTeaDirectoryPicker("📥 Select Input Directory (audio files to convert)", config.InputDir)
			if err != nil {
				printError(fmt.Sprintf("Error: %v", err))
				continue
			}
			config.InputDir = dir
			saveConfig(configDir, *config)

		case "o":
			dir, err := RunBubbleTeaDirectoryPicker("📤 Select Output Directory (where converted files will be saved)", config.OutputDir)
			if err != nil {
				printError(fmt.Sprintf("Error: %v", err))
				continue
			}
			config.OutputDir = dir
			saveConfig(configDir, *config)

		case "c":
			codecPrompt := promptui.Select{
				Label:     "Select target codec",
				Items:     []string{"aac", "alac", "flac", "mp3", "opus", "wav"},
				CursorPos: findIndex([]string{"aac", "alac", "flac", "mp3", "opus", "wav"}, config.Codec),
			}
			_, codec, err := codecPrompt.Run()
			if err != nil {
				continue
			}
			config.Codec = codec
			saveConfig(configDir, *config)

		case "p":
			config.IPod = !config.IPod
			saveConfig(configDir, *config)

		case "l":
			config.NoLyrics = !config.NoLyrics
			saveConfig(configDir, *config)

		case "", "s", "start":
			// Validate configuration
			if config.InputDir == "" || config.OutputDir == "" {
				printError("Please set both input and output directories before starting")
				continue
			}
			if config.Codec == "" && !config.IPod {
				printError("Please set a codec or enable iPod mode before starting")
				continue
			}
			fmt.Println()
			printSuccess("Starting conversion...")
			return nil

		case "q", "quit", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)

		default:
			printWarning("Invalid option. Please try again.")
		}
	}
}

func findIndex(items []string, target string) int {
	for i, item := range items {
		if item == target {
			return i
		}
	}
	return 0
}

func trimQuotes(s string) string {
	// Remove surrounding single or double quotes
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func expandPath(path string) string {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") || path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~" {
			return homeDir
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
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

func ensureFFmpeg(configDir string) (string, error) {
	binDir := filepath.Join(configDir, "bin")
	localFFmpeg := filepath.Join(binDir, "ffmpeg")
	if runtime.GOOS == "windows" {
		localFFmpeg += ".exe"
	}

	// Priority 1: Check if we have embedded binaries and extract them
	if hasEmbeddedBinaries() {
		// Check if already extracted
		if _, err := os.Stat(localFFmpeg); err == nil {
			if testFFmpeg(localFFmpeg) == nil {
				return localFFmpeg, nil
			}
		}

		// Extract embedded binaries
		printInfo("Extracting embedded ffmpeg binaries...")
		if err := extractEmbeddedFFmpeg(binDir); err != nil {
			fmt.Printf("Warning: Failed to extract embedded binaries: %v\n", err)
			// Continue to other methods
		} else {
			if testFFmpeg(localFFmpeg) == nil {
				return localFFmpeg, nil
			}
		}
	}

	// Priority 2: Try to find ffmpeg in PATH
	if path, err := findInPath("ffmpeg"); err == nil {
		if testFFmpeg(path) == nil {
			fmt.Printf("Found ffmpeg in PATH: %s\n", path)
			return path, nil
		}
	}

	// Priority 3: Check if we already downloaded it
	if _, err := os.Stat(localFFmpeg); err == nil {
		if testFFmpeg(localFFmpeg) == nil {
			return localFFmpeg, nil
		}
	}

	// Priority 4: Download ffmpeg
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
