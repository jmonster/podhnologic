package main

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMenuDirectoryPickerStaysInParentProgram(t *testing.T) {
	dir := t.TempDir()
	config := &Config{}
	m := NewMenuModel(config, dir)

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = model.(menuModel)
	if m.childAction != "input" || m.dirPicker == nil {
		t.Fatalf("input action did not activate embedded picker: %#v", m)
	}
	// Init is normally scheduled by Bubble Tea; seed the current directory
	// here so the focused model test can drive completion synchronously.
	m.dirPicker.filepicker.CurrentDirectory = dir

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = model.(menuModel)
	if m.childAction != "" || m.dirPicker != nil {
		t.Fatalf("picker did not return control to parent: %#v", m)
	}
	if config.InputDir != dir {
		t.Fatalf("InputDir = %q, want %q", config.InputDir, dir)
	}
	if _, err := os.Stat(config.InputDir); err != nil {
		t.Fatalf("selected directory is unavailable: %v", err)
	}
}

func TestMenuCodecPickerSelectionStaysInParentProgram(t *testing.T) {
	config := &Config{Codec: "aac"}
	m := NewMenuModel(config, t.TempDir())
	m.cursor = 2

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(menuModel)
	if m.childAction != "codec" || m.codecPicker == nil {
		t.Fatalf("codec action did not activate embedded picker: %#v", m)
	}

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(menuModel)
	if m.childAction != "" || m.codecPicker != nil {
		t.Fatalf("codec picker did not return control to parent: %#v", m)
	}
	if config.Codec != "aac" {
		t.Fatalf("Codec = %q, want aac", config.Codec)
	}
}

func TestMenuStartConversionRejectsIncompatibleIPodCodec(t *testing.T) {
	m := NewMenuModel(&Config{InputDir: t.TempDir(), OutputDir: t.TempDir(), Codec: "flac", IPod: true}, t.TempDir())
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	got := model.(menuModel)
	if cmd != nil || got.shouldStart {
		t.Fatal("incompatible iPod codec should keep the menu open")
	}
	if got.errorMessage == "" {
		t.Fatal("expected an actionable validation error")
	}
}
