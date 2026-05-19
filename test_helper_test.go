package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHelper provides utilities for tests that only need filesystem fixtures.
type TestHelper struct {
	t         *testing.T
	tempDir   string
	inputDir  string
	outputDir string
}

func (h *TestHelper) Setup() {
	if err := os.MkdirAll(h.inputDir, 0755); err != nil {
		h.t.Fatalf("Failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(h.outputDir, 0755); err != nil {
		h.t.Fatalf("Failed to create output dir: %v", err)
	}
}

func (h *TestHelper) WriteInputFile(filename string, content []byte) string {
	path := filepath.Join(h.inputDir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		h.t.Fatalf("Failed to create input subdir: %v", err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		h.t.Fatalf("Failed to write input file: %v", err)
	}
	return path
}
