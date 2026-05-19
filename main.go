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

const (
	Version = "4.1.0"
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
