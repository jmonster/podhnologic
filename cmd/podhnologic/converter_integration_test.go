//go:build linkedffmpeg_cgo && linkedffmpeg_hidden && cgo

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLargeLyricsSurviveAACConversion(t *testing.T) {
	h := NewTestHelper(t)
	h.Setup()
	wav := filepath.Join(h.tempDir, "source.wav")
	writePCM16WAV(t, wav, 44100, 4410)
	// Exceeds Linux's per-argument limit and macOS's total argument limit.
	// Include Unicode and newlines to exercise the JSON pipe.
	lyrics := strings.Repeat("A lyric with é and a newline\n", 65536)
	source := filepath.Join(h.inputDir, "large-lyrics.flac")
	if out, err := runFFmpeg([]string{"-i", wav, "-metadata", "lyrics=" + lyrics, "-c:a", "flac", source}); err != nil {
		t.Fatalf("create large-lyrics FLAC: %v\n%s", err, out)
	}
	for _, noLyrics := range []bool{false, true} {
		name := "preserved"
		if noLyrics {
			name = "removed"
		}
		t.Run(name, func(t *testing.T) {
			config := Config{InputDir: h.inputDir, OutputDir: filepath.Join(h.outputDir, name), Codec: "aac", IPod: true, NoLyrics: noLyrics}
			if err := processFile(source, config, false); err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(config.OutputDir, "large-lyrics.m4a")
			assertAudioStream(t, output, "aac", "44100", "")
			metadata, err := extractMetadata(output)
			if err != nil {
				t.Fatal(err)
			}
			got, exists := metadata.Format.Tags["lyrics"]
			if noLyrics {
				if exists {
					t.Fatal("--no-lyrics retained the oversized tag")
				}
			} else if got != lyrics {
				t.Fatalf("lyrics changed: got %d bytes, want %d", len(got), len(lyrics))
			}
		})
	}
}

func TestNativeAACNMRWithIPodSettings(t *testing.T) {
	temp := t.TempDir()
	wav := filepath.Join(temp, "source.wav")
	writePCM16WAV(t, wav, 48000, 48000)
	output := filepath.Join(temp, "ipod.m4a")
	// Exercise the Linux/Windows encoder even when tests run on macOS.
	args := append([]string{"-i", wav}, getCodecParamsForPlatform(Config{Codec: "aac", IPod: true}, "linux")...)
	args = append(args, output)
	if out, err := runFFmpeg(args); err != nil {
		t.Fatalf("native NMR AAC with PNS disabled: %v\n%s", err, out)
	}
	stream := assertAudioStream(t, output, "aac", "44100", "")
	// --enable-small builds report AV_PROFILE_AAC_LOW numerically (1).
	if stream.Profile != "LC" && stream.Profile != "1" {
		t.Fatalf("iPod AAC profile = %q, want LC", stream.Profile)
	}
	assertMoovBeforeMdat(t, output)
	decoded := filepath.Join(temp, "decoded.wav")
	if out, err := runFFmpeg([]string{"-i", output, "-c:a", "pcm_s16le", decoded}); err != nil {
		t.Fatalf("decode NMR AAC: %v\n%s", err, out)
	}
	assertAudioStream(t, decoded, "pcm_s16le", "44100", "s16")
}

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
