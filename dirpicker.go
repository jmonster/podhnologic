package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"
)

// DirectoryPicker provides an interactive directory selection interface
type DirectoryPicker struct {
	Label       string
	DefaultPath string
	StartPath   string
}

// Run displays the directory picker and returns the selected directory
func (dp *DirectoryPicker) Run() (string, error) {
	// Start from default path if provided, otherwise current directory
	startPath := dp.StartPath
	if startPath == "" {
		if dp.DefaultPath != "" {
			startPath = dp.DefaultPath
		} else {
			var err error
			startPath, err = os.Getwd()
			if err != nil {
				startPath = "."
			}
		}
	}

	// Expand ~ to home directory
	startPath = expandPath(startPath)

	// If path doesn't exist, start from home directory
	if _, err := os.Stat(startPath); os.IsNotExist(err) {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			startPath = homeDir
		}
	}

	// Make absolute
	absPath, err := filepath.Abs(startPath)
	if err == nil {
		startPath = absPath
	}

	return dp.navigate(startPath)
}

func (dp *DirectoryPicker) navigate(currentPath string) (string, error) {
	for {
		// Get directories in current path
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return "", fmt.Errorf("failed to read directory: %w", err)
		}

		// Filter for directories only
		var dirs []string
		for _, entry := range entries {
			if entry.IsDir() {
				name := entry.Name()
				// Skip hidden directories in the list (but allow navigating to them by typing)
				if !strings.HasPrefix(name, ".") {
					dirs = append(dirs, name)
				}
			}
		}

		sort.Strings(dirs)

		// Build menu items
		items := []string{}

		// Add current directory as first option
		defaultIndicator := ""
		if dp.DefaultPath != "" && currentPath == filepath.Clean(expandPath(dp.DefaultPath)) {
			defaultIndicator = " [default]"
		}
		items = append(items, fmt.Sprintf("✓ Select this directory%s", defaultIndicator))

		// Add type custom path option
		items = append(items, "✎ Type custom path...")

		// Add parent directory if not at root
		if currentPath != filepath.Dir(currentPath) {
			items = append(items, ".. (parent directory)")
		}

		// Add subdirectories
		for _, dir := range dirs {
			items = append(items, "📁 "+dir)
		}

		// Create prompt
		prompt := promptui.Select{
			Label: fmt.Sprintf("%s [%s]", dp.Label, shortenPath(currentPath)),
			Items: items,
			Size:  15,
			Templates: &promptui.SelectTemplates{
				Help: "{{ \"Use arrow keys to navigate, Enter to select, Ctrl+C to cancel\" | faint }}",
			},
		}

		idx, result, err := prompt.Run()
		if err != nil {
			return "", err
		}

		// Handle selection
		if idx == 0 {
			// Select current directory
			return currentPath, nil
		} else if idx == 1 {
			// Type custom path
			return dp.promptCustomPath(currentPath)
		} else if idx == 2 && currentPath != filepath.Dir(currentPath) {
			// Navigate to parent
			currentPath = filepath.Dir(currentPath)
		} else {
			// Navigate to subdirectory
			// Remove emoji prefix if present
			dirName := strings.TrimPrefix(result, "📁 ")
			newPath := filepath.Join(currentPath, dirName)

			// Verify it's a valid directory
			if info, err := os.Stat(newPath); err == nil && info.IsDir() {
				currentPath = newPath
			}
		}
	}
}

func (dp *DirectoryPicker) promptCustomPath(currentPath string) (string, error) {
	placeholder := ""
	if dp.DefaultPath != "" {
		placeholder = dp.DefaultPath
	}

	prompt := promptui.Prompt{
		Label:   "Enter path (or press Enter for default)",
		Default: placeholder,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	result = strings.TrimSpace(result)

	// If empty, use default
	if result == "" && dp.DefaultPath != "" {
		return expandPath(dp.DefaultPath), nil
	}

	// Expand and validate path
	path := expandPath(trimQuotes(result))

	// Make absolute if relative
	if !filepath.IsAbs(path) {
		path = filepath.Join(currentPath, path)
	}

	return path, nil
}

// shortenPath shortens a path for display by replacing home directory with ~
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
