package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConversionRejectsOutputCollisionsBeforeProcessing(t *testing.T) {
	h := NewTestHelper(t)
	h.Setup()
	h.WriteInputFile("same.wav", []byte("first source"))
	h.WriteInputFile("same.flac", []byte("second source"))
	err := runConversion(Config{InputDir: h.inputDir, OutputDir: h.outputDir, Codec: "mp3"}, false)
	if err == nil || !strings.Contains(err.Error(), "output collision") {
		t.Fatalf("expected collision before probing inputs, got %v", err)
	}
	entries, err := os.ReadDir(h.outputDir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("collision wrote output: %v, %v", entries, err)
	}
}

func TestCollectAudioFilesExcludesDestination(t *testing.T) {
	h := NewTestHelper(t)
	h.Setup()
	want := h.WriteInputFile("album/song.wav", []byte("source"))
	h.WriteInputFile("converted/album/song.flac", []byte("prior output"))
	h.WriteInputFile("converted-extra/other.flac", []byte("another source"))
	files, err := collectAudioFiles(h.inputDir, filepath.Join(h.inputDir, "converted"))
	if err != nil || len(files) != 2 || files[0] != want {
		t.Fatalf("destination exclusion removed wrong files: %v, %v", files, err)
	}
}

func TestConversionRejectsSameInputAndOutput(t *testing.T) {
	h := NewTestHelper(t)
	h.Setup()
	err := runConversion(Config{InputDir: h.inputDir, OutputDir: h.inputDir, Codec: "flac"}, false)
	if err == nil || !strings.Contains(err.Error(), "different directories") {
		t.Fatalf("expected same-directory rejection, got %v", err)
	}
}

func TestPublishOutputDoesNotReplaceExistingFile(t *testing.T) {
	dir := t.TempDir()
	temp := filepath.Join(dir, "work.partial")
	dest := filepath.Join(dir, "song.flac")
	if err := os.WriteFile(temp, []byte("completed audio"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("existing audio"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publishOutput(temp, dest); err == nil {
		t.Fatal("publication replaced an existing destination")
	}
	data, err := os.ReadFile(dest)
	if err != nil || string(data) != "existing audio" {
		t.Fatalf("existing output changed: %q, %v", data, err)
	}
	if err := os.Remove(dest); err != nil {
		t.Fatal(err)
	}
	if err := publishOutput(temp, dest); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(dest)
	if err != nil || string(data) != "completed audio" {
		t.Fatalf("wrong published audio: %q, %v", data, err)
	}
	if _, err := os.Stat(temp); !os.IsNotExist(err) {
		t.Fatalf("temporary file remains after publication: %v", err)
	}
}

func TestExistingEmptyOutputIsNotReportedComplete(t *testing.T) {
	h := NewTestHelper(t)
	h.Setup()
	source := h.WriteInputFile("song.wav", []byte("source"))
	if err := os.WriteFile(filepath.Join(h.outputDir, "song.flac"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	err := processFile(source, Config{InputDir: h.inputDir, OutputDir: h.outputDir, Codec: "flac"}, false)
	if err == nil || !strings.Contains(err.Error(), "not a completed audio file") {
		t.Fatalf("empty output was accepted: %v", err)
	}
}

func TestMetadataUsesFirstAudioStreamAndContainerOverrides(t *testing.T) {
	var metadata Metadata
	err := json.Unmarshal([]byte(`{"format":{"tags":{"TITLE":"Container title"}},"streams":[
		{"codec_type":"video","tags":{"artist":"Cover artist"}},
		{"codec_type":"audio","tags":{"title":"Stream title","ARTIST":"Music artist","lyrics":"Words","comment":"Drop me"}},
		{"codec_type":"audio","tags":{"artist":"Second stream artist"}}
	]}`), &metadata)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(buildFFmpegArgs("in.opus", "out.flac", Config{Codec: "flac"}, &metadata), " ")
	for _, want := range []string{"title=Container title", "artist=Music artist", "lyrics=Words"} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing %q: %s", want, args)
		}
	}
	for _, unwanted := range []string{"Stream title", "Cover artist", "Second stream artist", "Drop me"} {
		if strings.Contains(args, unwanted) {
			t.Fatalf("unexpected %q: %s", unwanted, args)
		}
	}
	args = strings.Join(buildFFmpegArgs("in.opus", "out.flac", Config{Codec: "flac", NoLyrics: true}, &metadata), " ")
	if strings.Contains(args, "lyrics=") {
		t.Fatalf("stream lyrics survived --no-lyrics: %s", args)
	}
}
