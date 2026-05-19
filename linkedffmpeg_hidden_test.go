//go:build linkedffmpeg_hidden

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLinkedFFmpegBridgeHandlerUnavailableWithoutCgo(t *testing.T) {
	req := LinkedFFmpegRequest{
		Tool: LinkedFFmpegToolFFprobe,
		Args: []string{"-version"},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code, err := handleLinkedFFmpegBridge(bytes.NewReader(payload), &stdout, &stderr)
	if err == nil {
		if code != 0 {
			t.Fatalf("handleLinkedFFmpegBridge code = %d, want 0", code)
		}
		return
	}
	if !errors.Is(err, ErrLinkedFFmpegUnavailable) {
		t.Fatalf("handleLinkedFFmpegBridge error = %v, want ErrLinkedFFmpegUnavailable", err)
	}
	if code == 0 {
		t.Fatal("handleLinkedFFmpegBridge code = 0, want non-zero")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout len = %d, want 0", stdout.Len())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr len = %d, want 0", stderr.Len())
	}
}

func TestLinkedFFmpegConvertsM4AWithAttachedPNG(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("system ffmpeg is required to synthesize the attached-art fixture")
	}

	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("os.MkdirAll input failed: %v", err)
	}

	coverPNG, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatalf("decode cover fixture failed: %v", err)
	}
	coverPath := filepath.Join(tempDir, "cover.png")
	if err := os.WriteFile(coverPath, coverPNG, 0644); err != nil {
		t.Fatalf("write cover fixture failed: %v", err)
	}

	inputPath := filepath.Join(inputDir, "cover-art.m4a")
	cmd := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-f", "lavfi",
		"-i", "sine=frequency=440:duration=0.25",
		"-i", coverPath,
		"-map", "0:a",
		"-map", "1:v",
		"-c:a", "aac",
		"-b:a", "96k",
		"-c:v", "copy",
		"-disposition:v:0", "attached_pic",
		inputPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture ffmpeg failed: %v\n%s", err, output)
	}

	config := Config{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Codec:     "flac",
	}
	if err := processFile(inputPath, config, false); err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	outputPath := filepath.Join(outputDir, "cover-art.flac")
	result, err := runFFprobe([]string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		outputPath,
	})
	if err != nil {
		t.Fatalf("linked ffprobe failed: %v\n%s", err, result.Stderr)
	}

	var probed struct {
		Streams []struct {
			CodecName string `json:"codec_name"`
			CodecType string `json:"codec_type"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(result.Stdout, &probed); err != nil {
		t.Fatalf("decode ffprobe output failed: %v\n%s", err, result.Stdout)
	}

	var hasFLACAudio bool
	var hasPNGArt bool
	for _, stream := range probed.Streams {
		if stream.CodecType == "audio" && stream.CodecName == "flac" {
			hasFLACAudio = true
		}
		if stream.CodecType == "video" && stream.CodecName == "png" {
			hasPNGArt = true
		}
	}
	if !hasFLACAudio || !hasPNGArt {
		t.Fatalf("output streams = %#v, want flac audio and png attached art", probed.Streams)
	}
}

func TestLinkedFFmpegEncodersFromFLAC(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("os.MkdirAll input failed: %v", err)
	}

	wavPath := filepath.Join(tempDir, "source.wav")
	writePCM16WAV(t, wavPath, 48000, 48000)

	sourceFLAC := filepath.Join(inputDir, "encoder-source.flac")
	if output, err := runFFmpeg([]string{
		"-y",
		"-i", wavPath,
		"-metadata", "title=Encoder Proof",
		"-metadata", "artist=Podhnologic",
		"-c:a", "flac",
		sourceFLAC,
	}); err != nil {
		t.Fatalf("linked ffmpeg failed to create flac source: %v\n%s", err, output)
	}
	assertAudioStream(t, sourceFLAC, "flac", "48000", "")

	tests := []struct {
		name               string
		codec              string
		ipod               bool
		wantExt            string
		wantCodec          string
		wantSampleRate     string
		wantSampleFmt      string
		wantMoovBeforeMdat bool
	}{
		{
			name:               "alac ipod m4a",
			codec:              "alac",
			ipod:               true,
			wantExt:            ".m4a",
			wantCodec:          "alac",
			wantSampleRate:     "44100",
			wantSampleFmt:      "s16p",
			wantMoovBeforeMdat: true,
		},
		{
			name:               "aac ipod m4a mp4 container",
			codec:              "aac",
			ipod:               true,
			wantExt:            ".m4a",
			wantCodec:          "aac",
			wantSampleRate:     "44100",
			wantMoovBeforeMdat: true,
		},
		{
			name:      "flac",
			codec:     "flac",
			wantExt:   ".flac",
			wantCodec: "flac",
		},
		{
			name:      "mp3",
			codec:     "mp3",
			wantExt:   ".mp3",
			wantCodec: "mp3",
		},
		{
			name:      "opus",
			codec:     "opus",
			wantExt:   ".opus",
			wantCodec: "opus",
		},
		{
			name:           "wav",
			codec:          "wav",
			wantExt:        ".wav",
			wantCodec:      "pcm_s16le",
			wantSampleRate: "48000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := filepath.Join(tempDir, "output", tt.name)
			config := Config{
				InputDir:  inputDir,
				OutputDir: outputDir,
				Codec:     tt.codec,
				IPod:      tt.ipod,
			}
			if err := processFile(sourceFLAC, config, false); err != nil {
				t.Fatalf("processFile failed: %v", err)
			}

			outputPath := filepath.Join(outputDir, "encoder-source"+tt.wantExt)
			assertAudioStream(t, outputPath, tt.wantCodec, tt.wantSampleRate, tt.wantSampleFmt)
			if tt.wantMoovBeforeMdat {
				assertMoovBeforeMdat(t, outputPath)
			}
		})
	}
}

type probedStream struct {
	CodecName  string `json:"codec_name"`
	CodecType  string `json:"codec_type"`
	SampleFmt  string `json:"sample_fmt"`
	SampleRate string `json:"sample_rate"`
}

func assertAudioStream(t *testing.T, path, wantCodec, wantSampleRate, wantSampleFmt string) probedStream {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("output stat failed for %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatalf("output file is empty: %s", path)
	}

	result, err := runFFprobe([]string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		path,
	})
	if err != nil {
		t.Fatalf("linked ffprobe failed for %s: %v\n%s", path, err, result.Stderr)
	}

	var probed struct {
		Streams []probedStream `json:"streams"`
	}
	if err := json.Unmarshal(result.Stdout, &probed); err != nil {
		t.Fatalf("decode ffprobe output failed for %s: %v\n%s", path, err, result.Stdout)
	}

	for _, stream := range probed.Streams {
		if stream.CodecType != "audio" {
			continue
		}
		if stream.CodecName != wantCodec {
			t.Fatalf("%s audio codec = %q, want %q", path, stream.CodecName, wantCodec)
		}
		if wantSampleRate != "" && stream.SampleRate != wantSampleRate {
			t.Fatalf("%s sample_rate = %q, want %q", path, stream.SampleRate, wantSampleRate)
		}
		if wantSampleFmt != "" && stream.SampleFmt != wantSampleFmt {
			t.Fatalf("%s sample_fmt = %q, want %q", path, stream.SampleFmt, wantSampleFmt)
		}
		return stream
	}

	t.Fatalf("%s has no audio stream; streams = %#v", path, probed.Streams)
	return probedStream{}
}

func assertMoovBeforeMdat(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output failed for %s: %v", path, err)
	}
	moov := bytes.Index(data, []byte("moov"))
	mdat := bytes.Index(data, []byte("mdat"))
	if moov < 0 || mdat < 0 {
		t.Fatalf("%s missing moov or mdat atom; moov=%d mdat=%d", path, moov, mdat)
	}
	if moov > mdat {
		t.Fatalf("%s is not faststart; moov=%d mdat=%d", path, moov, mdat)
	}
}

func writePCM16WAV(t *testing.T, path string, sampleRate, samples int) {
	t.Helper()

	const channels = 1
	const bitsPerSample = 16
	dataBytes := samples * channels * bitsPerSample / 8

	var buf bytes.Buffer
	buf.WriteString("RIFF")
	writeLittleEndian(t, &buf, uint32(36+dataBytes))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	writeLittleEndian(t, &buf, uint32(16))
	writeLittleEndian(t, &buf, uint16(1))
	writeLittleEndian(t, &buf, uint16(channels))
	writeLittleEndian(t, &buf, uint32(sampleRate))
	writeLittleEndian(t, &buf, uint32(sampleRate*channels*bitsPerSample/8))
	writeLittleEndian(t, &buf, uint16(channels*bitsPerSample/8))
	writeLittleEndian(t, &buf, uint16(bitsPerSample))
	buf.WriteString("data")
	writeLittleEndian(t, &buf, uint32(dataBytes))

	for i := 0; i < samples; i++ {
		phase := 2 * math.Pi * 440 * float64(i) / float64(sampleRate)
		sample := int16(math.Sin(phase) * 12000)
		writeLittleEndian(t, &buf, sample)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write wav fixture failed: %v", err)
	}
}

func writeLittleEndian(t *testing.T, buf *bytes.Buffer, value any) {
	t.Helper()
	if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
		t.Fatalf("binary.Write failed for %T: %v", value, err)
	}
}
