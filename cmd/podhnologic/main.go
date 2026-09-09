package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Version is a variable so release builds can inject it with -ldflags -X.
var Version = "4.2.0"

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
		fmt.Printf("podhnologic v%s\n", strings.TrimPrefix(Version, "v"))
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
		codecSet := false
		ipodSet := false
		noLyricsSet := false
		flag.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "codec":
				codecSet = true
			case "ipod":
				ipodSet = true
			case "no-lyrics":
				noLyricsSet = true
			}
		})
		var overrideErr error
		config, overrideErr = applyFlagOverrides(config, *codecFlag, codecSet, *ipodFlag, ipodSet, *noLyricsFlag, noLyricsSet)
		if overrideErr != nil {
			log.Fatal(overrideErr)
		}

		// Validate required fields
		if config.InputDir == "" || config.OutputDir == "" {
			log.Fatal("--input and --output are required")
		}

		// Save the config for future use
		if err := saveConfig(configDir, config); err != nil {
			log.Printf("Warning: failed to save config: %v", err)
		}
	}

	// Keep the codec passed to conversion normalized in every mode, including
	// configurations loaded for interactive use.
	config.Codec = normalizeCodec(config.Codec)

	// Run the conversion
	if err := runConversion(config, *dryRunFlag); err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}
}

func applyFlagOverrides(config Config, codec string, codecSet bool, ipod bool, ipodSet bool, noLyrics bool, noLyricsSet bool) (Config, error) {
	if codecSet {
		config.Codec = codec
	}
	if ipodSet {
		config.IPod = ipod
	}
	if noLyricsSet {
		config.NoLyrics = noLyrics
	}

	config.Codec = normalizeCodec(config.Codec)
	if config.IPod && !codecSet && !isIPodCodec(config.Codec) {
		config.Codec = "aac"
	}
	if config.Codec == "" && config.IPod {
		config.Codec = "aac"
	}
	if err := validateConfig(config); err != nil {
		return config, err
	}
	return config, nil
}

var supportedCodecs = map[string]struct{}{
	"flac": {}, "alac": {}, "aac": {}, "wav": {}, "mp3": {}, "opus": {},
}

func normalizeCodec(codec string) string {
	return strings.ToLower(strings.TrimSpace(codec))
}

func isIPodCodec(codec string) bool {
	return codec == "aac" || codec == "alac"
}

// validateConfig checks codec values shared by command-line, terminal, and
// conversion flows.
func validateConfig(config Config) error {
	codec := normalizeCodec(config.Codec)
	if codec == "" {
		if config.IPod {
			return nil // iPod mode supplies the aac default in the caller.
		}
		return fmt.Errorf("a codec is required")
	}
	if _, ok := supportedCodecs[codec]; !ok {
		return fmt.Errorf("unsupported codec %q (use flac, alac, aac, wav, mp3, or opus)", config.Codec)
	}
	if config.IPod && !isIPodCodec(codec) {
		return fmt.Errorf("codec %q is not compatible with iPod mode (use aac or alac)", config.Codec)
	}
	return nil
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
	// Create and run the interactive menu
	menu := NewMenuModel(config, configDir)

	p := tea.NewProgram(menu, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running menu: %w", err)
	}

	// Check if user wants to start conversion
	if m, ok := finalModel.(menuModel); ok {
		if m.shouldStart {
			fmt.Println()
			printSuccess("Starting conversion...")
			return nil
		}
	}

	// User quit without starting
	return fmt.Errorf("cancelled by user")
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

func shortenPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if strings.HasPrefix(path, homeDir) {
		return "~" + strings.TrimPrefix(path, homeDir)
	}

	return path
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
