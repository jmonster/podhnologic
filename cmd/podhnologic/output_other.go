//go:build !darwin && !linux && !windows

package main

import "os"

func publishOutput(tempPath, outputPath string) error {
	if err := os.Link(tempPath, outputPath); err != nil {
		return err
	}
	return os.Remove(tempPath)
}
