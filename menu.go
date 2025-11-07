package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
	"github.com/muesli/termenv"
)

type menuItem struct {
	label    string
	shortcut string
	value    func(*Config) string
	action   string
}

type menuModel struct {
	config        *Config
	configDir     string
	items         []menuItem
	cursor        int
	width         int
	height        int
	quitting      bool
	shouldStart   bool
	errorMessage  string
	successMessage string
}

var (
	selectedItemStyle   lipgloss.Style
	normalItemStyle     lipgloss.Style
	shortcutStyle       lipgloss.Style
	valueStyle          lipgloss.Style
	emptyValueStyle     lipgloss.Style
	menuTitleStyle      lipgloss.Style
	errorStyle          lipgloss.Style
	successStyle        lipgloss.Style
	menuHelpStyle       lipgloss.Style
	menuStylesInited    bool
)

func initMenuStyles() {
	if menuStylesInited {
		return
	}
	menuStylesInited = true

	// Detect terminal background
	output := termenv.DefaultOutput()
	isLight := !output.HasDarkBackground()

	if isLight {
		// Light mode: darker, high-contrast colors
		selectedItemStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("240")).   // Dark gray background
			Foreground(lipgloss.Color("255")).   // White text
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2)

		shortcutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).  // Orange
			Bold(true)

		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("22"))   // Dark green

		emptyValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Dark gray
			Italic(true)

		menuTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("198")).  // Hot pink
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

		errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("160")).  // Dark red
			Bold(true)

		successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("22")).   // Dark green
			Bold(true)

		menuHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))   // Dark gray

	} else {
		// Dark mode: bright, vibrant colors
		selectedItemStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("213")).   // Hot pink background
			Foreground(lipgloss.Color("16")).    // Black text
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2)

		shortcutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).  // Bright yellow
			Bold(true)

		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))    // Bright green

		emptyValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")). // Medium gray
			Italic(true)

		menuTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).  // Hot pink
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

		errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).  // Bright red
			Bold(true)

		successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).   // Neon green
			Bold(true)

		menuHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))   // Medium gray
	}

	normalItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2)
}

func NewMenuModel(config *Config, configDir string) menuModel {
	items := []menuItem{
		{
			label:    "Input Directory",
			shortcut: "I",
			value: func(c *Config) string {
				if c.InputDir == "" {
					return "(not set)"
				}
				return shortenPath(c.InputDir)
			},
			action: "input",
		},
		{
			label:    "Output Directory",
			shortcut: "O",
			value: func(c *Config) string {
				if c.OutputDir == "" {
					return "(not set)"
				}
				return shortenPath(c.OutputDir)
			},
			action: "output",
		},
		{
			label:    "Codec",
			shortcut: "C",
			value: func(c *Config) string {
				if c.Codec == "" {
					return "(not set)"
				}
				return c.Codec
			},
			action: "codec",
		},
		{
			label:    "iPod Mode",
			shortcut: "P",
			value: func(c *Config) string {
				if c.IPod {
					return "enabled"
				}
				return "disabled"
			},
			action: "ipod",
		},
		{
			label:    "Lyrics",
			shortcut: "L",
			value: func(c *Config) string {
				if c.NoLyrics {
					return "strip lyrics"
				}
				return "keep lyrics"
			},
			action: "lyrics",
		},
	}

	return menuModel{
		config:    config,
		configDir: configDir,
		items:     items,
		cursor:    0,
	}
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear messages on any key press
		m.errorMessage = ""
		m.successMessage = ""

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.handleAction()

		// Shortcut keys
		case "i", "I":
			m.cursor = 0
			return m.handleAction()

		case "o", "O":
			m.cursor = 1
			return m.handleAction()

		case "c", "C":
			m.cursor = 2
			return m.handleAction()

		case "p", "P":
			m.cursor = 3
			return m.handleAction()

		case "l", "L":
			m.cursor = 4
			return m.handleAction()

		case "s", "S":
			return m.startConversion()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m menuModel) handleAction() (tea.Model, tea.Cmd) {
	action := m.items[m.cursor].action

	switch action {
	case "input":
		dir, err := RunBubbleTeaDirectoryPicker("📥 Select Input Directory (audio files to convert)", m.config.InputDir)
		if err == nil && dir != "" {
			m.config.InputDir = dir
			saveConfig(m.configDir, *m.config)
		}
		// Force a full redraw after returning from sub-program
		return m, tea.ClearScreen

	case "output":
		dir, err := RunBubbleTeaDirectoryPicker("📤 Select Output Directory (converted files)", m.config.OutputDir)
		if err == nil && dir != "" {
			m.config.OutputDir = dir
			saveConfig(m.configDir, *m.config)
		}
		// Force a full redraw after returning from sub-program
		return m, tea.ClearScreen

	case "codec":
		codec, err := selectCodec(m.config.Codec)
		if err == nil && codec != "" {
			m.config.Codec = codec
			saveConfig(m.configDir, *m.config)
		}
		// Force a full redraw after returning from sub-program
		return m, tea.ClearScreen

	case "ipod":
		m.config.IPod = !m.config.IPod
		if m.config.IPod && m.config.Codec == "" {
			m.config.Codec = "aac"
		}
		saveConfig(m.configDir, *m.config)

	case "lyrics":
		m.config.NoLyrics = !m.config.NoLyrics
		saveConfig(m.configDir, *m.config)
	}

	return m, nil
}

func (m menuModel) startConversion() (tea.Model, tea.Cmd) {
	// Validate configuration
	if m.config.InputDir == "" || m.config.OutputDir == "" {
		m.errorMessage = "⚠ Please set both input and output directories"
		return m, nil
	}
	if m.config.Codec == "" && !m.config.IPod {
		m.errorMessage = "⚠ Please set a codec or enable iPod mode"
		return m, nil
	}

	m.shouldStart = true
	return m, tea.Quit
}

func (m menuModel) View() string {
	// Initialize styles based on terminal background
	initMenuStyles()

	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Banner
	b.WriteString("\n")
	b.WriteString(renderBanner())
	b.WriteString("\n")

	// Title
	b.WriteString(menuTitleStyle.Render("  Configuration"))
	b.WriteString("\n\n")

	// Menu items
	for i, item := range m.items {
		value := item.value(m.config)

		if m.cursor == i {
			// Selected item - render entire line with contrasting text on background
			line := fmt.Sprintf("[%s]  %-18s %s",
				item.shortcut,
				item.label+":",
				value,
			)
			b.WriteString(selectedItemStyle.Render(line))
		} else {
			// Normal item - render with colored components
			valueStyled := valueStyle.Render(value)
			if value == "(not set)" {
				valueStyled = emptyValueStyle.Render(value)
			}

			line := fmt.Sprintf("%s  %-18s %s",
				shortcutStyle.Render("["+item.shortcut+"]"),
				item.label+":",
				valueStyled,
			)
			b.WriteString(normalItemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Actions
	b.WriteString("\n")
	b.WriteString(normalItemStyle.Render(
		shortcutStyle.Render("[S]") + " Start Conversion   " +
			shortcutStyle.Render("[Q]") + " Quit",
	))
	b.WriteString("\n\n")

	// Messages
	if m.errorMessage != "" {
		b.WriteString("  ")
		b.WriteString(errorStyle.Render(m.errorMessage))
		b.WriteString("\n")
	}
	if m.successMessage != "" {
		b.WriteString("  ")
		b.WriteString(successStyle.Render(m.successMessage))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString(normalItemStyle.Render(
		menuHelpStyle.Render("↑↓ or j/k to navigate • Enter/Shortcut to select • S to start"),
	))
	b.WriteString("\n")

	return b.String()
}

func renderBanner() string {
	var b strings.Builder

	// Get colors - they're initialized already in banner.go
	b.WriteString(fmt.Sprintf("%s%s", colorBold, colorHotPink))
	b.WriteString("  ██████╗  ██████╗ ██████╗ ██╗  ██╗███╗   ██╗ ██████╗ ██╗      ██████╗  ██████╗ ██╗ ██████╗\n")
	b.WriteString(fmt.Sprintf("%s", colorTurquoise))
	b.WriteString("  ██╔══██╗██╔═══██╗██╔══██╗██║  ██║████╗  ██║██╔═══██╗██║     ██╔═══██╗██╔════╝ ██║██╔════╝\n")
	b.WriteString(fmt.Sprintf("%s", colorYellow))
	b.WriteString("  ██████╔╝██║   ██║██║  ██║███████║██╔██╗ ██║██║   ██║██║     ██║   ██║██║  ███╗██║██║     \n")
	b.WriteString(fmt.Sprintf("%s", colorOrange))
	b.WriteString("  ██╔═══╝ ██║   ██║██║  ██║██╔══██║██║╚██╗██║██║   ██║██║     ██║   ██║██║   ██║██║██║     \n")
	b.WriteString(fmt.Sprintf("%s", colorHotPink))
	b.WriteString("  ██║     ╚██████╔╝██████╔╝██║  ██║██║ ╚████║╚██████╔╝███████╗╚██████╔╝╚██████╔╝██║╚██████╗\n")
	b.WriteString(fmt.Sprintf("%s", colorTurquoise))
	b.WriteString("  ╚═╝      ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝ ╚═════╝\n")
	b.WriteString(fmt.Sprintf("%s", colorReset))

	// Music wave
	b.WriteString(fmt.Sprintf("%s          %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s\n",
		colorReset,
		colorHotPink, colorReset, colorTurquoise, colorReset, colorYellow, colorReset,
		colorOrange, colorReset, colorHotPink, colorReset, colorTurquoise, colorReset,
		colorYellow, colorReset, colorOrange, colorReset, colorHotPink, colorReset,
		colorTurquoise, colorReset, colorYellow, colorReset, colorOrange, colorReset,
		colorHotPink, colorReset, colorTurquoise, colorReset))

	return b.String()
}

func selectCodec(current string) (string, error) {
	codecs := []string{"flac", "alac", "aac", "wav", "mp3", "opus"}

	// Find current index
	currentIndex := 0
	for i, codec := range codecs {
		if codec == current {
			currentIndex = i
			break
		}
	}

	prompt := promptui.Select{
		Label: "Select Target Codec",
		Items: codecs,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
		CursorPos: currentIndex,
		Size:      6,
	}

	_, result, err := prompt.Run()
	return result, err
}
