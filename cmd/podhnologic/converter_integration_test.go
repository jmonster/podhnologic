//go:build linkedffmpeg_cgo && linkedffmpeg_hidden && cgo

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func conversionFixture(t *testing.T, args ...string) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("system ffmpeg is required to create this fixture")
	}
	cmd := exec.Command("ffmpeg", append([]string{"-hide_banner", "-loglevel", "error", "-y"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create fixture: %v\n%s", err, out)
	}
}

func TestFailedConversionLeavesNoFinalOutputAndCanRetry(t *testing.T) {
	h := NewTestHelper(t)
	h.Setup()
	source := filepath.Join(h.inputDir, "song.mka")
	conversionFixture(t, "-f", "lavfi", "-i", "sine=frequency=440:duration=0.1", "-f", "lavfi", "-i", "sine=frequency=880:duration=0.1", "-map", "0:a", "-map", "1:a", "-c:a", "flac", source)
	config := Config{InputDir: h.inputDir, OutputDir: h.outputDir, Codec: "mp3"}
	if err := processFile(source, config, false); err == nil {
		t.Fatal("expected MP3 muxing of two audio streams to fail")
	}
	entries, err := os.ReadDir(h.outputDir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed conversion left artifacts: %v, %v", entries, err)
	}
	conversionFixture(t, "-f", "lavfi", "-i", "sine=duration=0.1", "-c:a", "flac", source)
	if err := runConversion(config, false); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	assertAudioStream(t, filepath.Join(h.outputDir, "song.mp3"), "mp3", "44100", "")
}

func TestNestedOutputAndStreamMetadataRoundTrip(t *testing.T) {
	h := NewTestHelper(t)
	h.Setup()
	source := filepath.Join(h.inputDir, "tagged.opus")
	conversionFixture(t, "-f", "lavfi", "-i", "sine=duration=0.1", "-c:a", "libopus", "-metadata", "title=Audit Title", "-metadata", "artist=Audit Artist", source)
	output := filepath.Join(h.inputDir, "converted")
	config := Config{InputDir: h.inputDir, OutputDir: output, Codec: "flac"}
	for i := 0; i < 2; i++ {
		if err := runConversion(config, false); err != nil {
			t.Fatal(err)
		}
	}
	files, err := collectAudioFiles(output)
	if err != nil || len(files) != 1 || files[0] != filepath.Join(output, "tagged.flac") {
		t.Fatalf("converted output was scanned again: %v, %v", files, err)
	}
	metadata, err := extractMetadata(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Format.Tags["title"] != "Audit Title" || metadata.Format.Tags["artist"] != "Audit Artist" {
		t.Fatalf("stream tags lost: %v", metadata.Format.Tags)
	}
}
