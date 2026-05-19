package main

import (
	"bytes"
	"context"
	"fmt"
)

func ffmpegRunner() LinkedFFmpegRunner {
	return NewLinkedFFmpegRunner(LinkedFFmpegModeHidden)
}

func runFFmpeg(args []string) ([]byte, error) {
	result, err := ffmpegRunner().FFmpeg(context.Background(), args...)
	if err != nil {
		return combinedFFmpegOutput(result), err
	}
	return combinedFFmpegOutput(result), nil
}

func runFFprobe(args []string) (LinkedFFmpegResult, error) {
	return ffmpegRunner().FFprobe(context.Background(), args...)
}

func combinedFFmpegOutput(result LinkedFFmpegResult) []byte {
	if len(result.Stdout) == 0 {
		return append([]byte(nil), result.Stderr...)
	}
	if len(result.Stderr) == 0 {
		return append([]byte(nil), result.Stdout...)
	}

	var output bytes.Buffer
	_, _ = output.Write(result.Stdout)
	if result.Stdout[len(result.Stdout)-1] != '\n' {
		_ = output.WriteByte('\n')
	}
	_, _ = output.Write(result.Stderr)
	return output.Bytes()
}

func linkedFFmpegBuildError(err error) error {
	return fmt.Errorf("%w; build with -tags 'linkedffmpeg_cgo linkedffmpeg_hidden' after running scripts/ffmpeg/build-native.sh", err)
}
