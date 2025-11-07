package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the directory picker
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")). // Cyan
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	selectedPathStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")). // Cyan
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type dirPickerModel struct {
	filepicker   filepicker.Model
	selectedPath string
	err          error
	label        string
	quitting     bool
}

type clearErrorMsg struct{}

func clearErrorAfter() tea.Cmd {
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m dirPickerModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m dirPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			// Select current directory
			m.selectedPath = m.filepicker.CurrentDirectory
			m.quitting = true
			return m, tea.Quit
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	return m, cmd
}

func (m dirPickerModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render(m.label))
	s.WriteString("\n\n")

	// Current directory
	s.WriteString(selectedPathStyle.Render("📂 " + shortenPath(m.filepicker.CurrentDirectory)))
	s.WriteString("\n\n")

	// File picker view
	s.WriteString(m.filepicker.View())
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
		s.WriteString("\n")
	}

	// Help text
	help := helpStyle.Render("↑↓: navigate • ←→/enter: enter/exit dir • enter: select current • q: quit")
	s.WriteString("\n" + help)

	return s.String()
}

// RunBubbleTeaDirectoryPicker runs the Bubble Tea directory picker
func RunBubbleTeaDirectoryPicker(label, defaultPath string) (string, error) {
	fp := filepicker.New()

	// Configure to show directories only
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false
	fp.AutoHeight = false
	fp.Height = 15

	// Set starting directory
	startPath := defaultPath
	if startPath == "" {
		var err error
		startPath, err = os.UserHomeDir()
		if err != nil {
			startPath = "."
		}
	}

	// Expand path
	startPath = expandPath(startPath)
	absPath, err := filepath.Abs(startPath)
	if err == nil {
		startPath = absPath
	}

	// Verify the path exists
	if _, err := os.Stat(startPath); os.IsNotExist(err) {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			startPath = homeDir
		} else {
			startPath = "."
		}
	}

	fp.CurrentDirectory = startPath

	// Custom styling
	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("212")) // Pink
	fp.Styles.Symlink = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(lipgloss.Color("99")) // Purple
	fp.Styles.File = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	fp.Styles.DisabledFile = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	m := dirPickerModel{
		filepicker: fp,
		label:      label,
	}

	tm, err := tea.NewProgram(&m, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return "", err
	}

	mm := tm.(dirPickerModel)
	if mm.selectedPath == "" {
		return "", fmt.Errorf("no directory selected")
	}

	return mm.selectedPath, nil
}
