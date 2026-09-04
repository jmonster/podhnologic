package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type menuItem struct {
	label    string
	shortcut string
	value    func(*Config) string
	action   string
}

type menuModel struct {
	config         *Config
	configDir      string
	items          []menuItem
	cursor         int
	width          int
	height         int
	quitting       bool
	shouldStart    bool
	errorMessage   string
	successMessage string
	childAction    string
	dirPicker      *dirPickerModel
	codecPicker    *codecPickerModel
}

var (
	selectedItemStyle lipgloss.Style
	normalItemStyle   lipgloss.Style
	shortcutStyle     lipgloss.Style
	valueStyle        lipgloss.Style
	emptyValueStyle   lipgloss.Style
	menuTitleStyle    lipgloss.Style
	errorStyle        lipgloss.Style
	successStyle      lipgloss.Style
	menuHelpStyle     lipgloss.Style
	menuStylesInited  bool
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
		// Light mode: darker, high-contrast Apple green colors
		selectedItemStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(appleLightPhosphorDark)).
			Foreground(lipgloss.Color(appleWhite)).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2)

		shortcutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleRainbowOrange)).
			Bold(true)

		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleRainbowBlue))

		emptyValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleGrayDark)).
			Italic(true)

		menuTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleLightPhosphorDark)).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

		errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleRainbowRed)).
			Bold(true)

		successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleLightPhosphorMid)).
			Bold(true)

		menuHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleGrayDark))

	} else {
		// Dark mode: bright Apple II phosphor colors
		selectedItemStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(applePhosphorMid)).
			Foreground(lipgloss.Color(appleBlack)).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2)

		shortcutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleRainbowYellow)).
			Bold(true)

		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleRainbowBlue))

		emptyValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleGrayDim)).
			Italic(true)

		menuTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(applePhosphorBright)).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

		errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleRainbowRed)).
			Bold(true)

		successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(applePhosphorBright)).
			Bold(true)

		menuHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(appleGrayDim))
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
	if m.childAction != "" {
		return m.updateChild(msg)
	}
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
		m.childAction = action
		m.dirPicker = newDirPickerModel("📥 Select Input Directory (audio files to convert)", m.config.InputDir)
		return m, m.dirPicker.Init()

	case "output":
		m.childAction = action
		m.dirPicker = newDirPickerModel("📤 Select Output Directory (converted files)", m.config.OutputDir)
		return m, m.dirPicker.Init()

	case "codec":
		m.childAction = action
		m.codecPicker = newCodecPicker(m.config.Codec)
		return m, m.codecPicker.Init()

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

func (m menuModel) updateChild(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.childAction {
	case "input", "output":
		m.dirPicker, cmd = updateDirPicker(m.dirPicker, msg)
		if m.dirPicker.quitting {
			if m.dirPicker.selectedPath != "" {
				if m.childAction == "input" {
					m.config.InputDir = m.dirPicker.selectedPath
				} else {
					m.config.OutputDir = m.dirPicker.selectedPath
				}
				saveConfig(m.configDir, *m.config)
			}
			m.childAction, m.dirPicker = "", nil
		}
	case "codec":
		m.codecPicker, cmd = updateCodecPicker(m.codecPicker, msg)
		if m.codecPicker.quitting {
			if m.codecPicker.selected != "" {
				m.config.Codec = m.codecPicker.selected
				saveConfig(m.configDir, *m.config)
			}
			m.childAction, m.codecPicker = "", nil
		}
	}
	return m, cmd
}

type codecPickerModel struct {
	codecs   []string
	cursor   int
	selected string
	quitting bool
}

func newCodecPicker(current string) *codecPickerModel {
	items := []string{"flac", "alac", "aac", "wav", "mp3", "opus"}
	cursor := 0
	for i, codec := range items {
		if codec == current {
			cursor = i
			break
		}
	}
	return &codecPickerModel{codecs: items, cursor: cursor}
}

func (m codecPickerModel) Init() tea.Cmd { return nil }

func (m codecPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.codecs)-1 {
			m.cursor++
		}
	case "enter", " ":
		m.selected, m.quitting = m.codecs[m.cursor], true
	}
	return m, nil
}

func (m codecPickerModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select Target Codec"))
	b.WriteString("\n\n")
	for i, codec := range m.codecs {
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render("▶ " + codec))
		} else {
			b.WriteString("  " + codec)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n" + helpStyle.Render("↑↓ navigate • Enter select • Esc cancel"))
	return b.String()
}

func updateCodecPicker(m *codecPickerModel, msg tea.Msg) (*codecPickerModel, tea.Cmd) {
	model, cmd := m.Update(msg)
	value := model.(codecPickerModel)
	return &value, cmd
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
	if err := validateConfig(*m.config); err != nil {
		m.errorMessage = "⚠ " + err.Error()
		return m, nil
	}

	m.shouldStart = true
	return m, tea.Quit
}

func (m menuModel) View() string {
	// Child views reuse the menu's terminal-aware styles.
	initMenuStyles()
	if m.childAction != "" {
		if m.dirPicker != nil {
			return m.dirPicker.View()
		}
		if m.codecPicker != nil {
			return m.codecPicker.View()
		}
	}
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
	b.WriteString(fmt.Sprintf("%s%s", colorBold, colorPhosphorGlow))
	b.WriteString("  ██████╗  ██████╗ ██████╗ ██╗  ██╗███╗   ██╗ ██████╗ ██╗      ██████╗  ██████╗ ██╗ ██████╗\n")
	b.WriteString(fmt.Sprintf("%s", colorPhosphorMid))
	b.WriteString("  ██╔══██╗██╔═══██╗██╔══██╗██║  ██║████╗  ██║██╔═══██╗██║     ██╔═══██╗██╔════╝ ██║██╔════╝\n")
	b.WriteString(fmt.Sprintf("%s", colorPhosphorBright))
	b.WriteString("  ██████╔╝██║   ██║██║  ██║███████║██╔██╗ ██║██║   ██║██║     ██║   ██║██║  ███╗██║██║     \n")
	b.WriteString(fmt.Sprintf("%s", colorPhosphorDim))
	b.WriteString("  ██╔═══╝ ██║   ██║██║  ██║██╔══██║██║╚██╗██║██║   ██║██║     ██║   ██║██║   ██║██║██║     \n")
	b.WriteString(fmt.Sprintf("%s", colorPhosphorBright))
	b.WriteString("  ██║     ╚██████╔╝██████╔╝██║  ██║██║ ╚████║╚██████╔╝███████╗╚██████╔╝╚██████╔╝██║╚██████╗\n")
	b.WriteString(fmt.Sprintf("%s", colorPhosphorMid))
	b.WriteString("  ╚═╝      ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝ ╚═╝ ╚═════╝\n")
	b.WriteString(fmt.Sprintf("%s", colorReset))

	// Music wave
	b.WriteString(fmt.Sprintf("%s          %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s  %s♬%s  %s♪%s  %s♫%s\n",
		colorReset,
		colorPhosphorDim, colorReset, colorPhosphorMid, colorReset, colorPhosphorBright, colorReset,
		colorPhosphorGlow, colorReset, colorPhosphorDim, colorReset, colorPhosphorMid, colorReset,
		colorPhosphorBright, colorReset, colorPhosphorGlow, colorReset, colorPhosphorDim, colorReset,
		colorPhosphorMid, colorReset, colorPhosphorBright, colorReset, colorPhosphorGlow, colorReset,
		colorPhosphorDim, colorReset, colorPhosphorMid, colorReset))

	return b.String()
}
