package main

import (
	"encoding/json"
	"fmt"
)

func probeMetadata(filePath string) (*Metadata, error) {
	result, err := runFFprobe([]string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, string(result.Stderr))
	}

	var metadata Metadata
	if err := json.Unmarshal(result.Stdout, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe metadata: %w", err)
	}

	return &metadata, nil
}
