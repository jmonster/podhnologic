package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const linkedFFmpegBridgeEnv = "PODHNLOGIC_LINKED_FFMPEG_BRIDGE"

var ErrLinkedFFmpegUnavailable = errors.New("linked ffmpeg bridge unavailable")

type LinkedFFmpegTool string

const (
	LinkedFFmpegToolFFmpeg  LinkedFFmpegTool = "ffmpeg"
	LinkedFFmpegToolFFprobe LinkedFFmpegTool = "ffprobe"
)

func NormalizeLinkedFFmpegTool(tool string) (LinkedFFmpegTool, error) {
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case string(LinkedFFmpegToolFFmpeg):
		return LinkedFFmpegToolFFmpeg, nil
	case string(LinkedFFmpegToolFFprobe):
		return LinkedFFmpegToolFFprobe, nil
	default:
		return "", fmt.Errorf("unsupported linked ffmpeg tool %q", tool)
	}
}

type LinkedFFmpegRequest struct {
	Tool LinkedFFmpegTool `json:"tool"`
	Args []string         `json:"args"`
}

func (r LinkedFFmpegRequest) Validate() error {
	switch r.Tool {
	case LinkedFFmpegToolFFmpeg, LinkedFFmpegToolFFprobe:
		return nil
	default:
		return fmt.Errorf("unsupported linked ffmpeg tool %q", r.Tool)
	}
}

func (r LinkedFFmpegRequest) MarshalJSON() ([]byte, error) {
	type alias LinkedFFmpegRequest
	return json.Marshal(alias(r))
}

func ParseLinkedFFmpegRequest(data []byte) (LinkedFFmpegRequest, error) {
	var req LinkedFFmpegRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return LinkedFFmpegRequest{}, fmt.Errorf("decode linked ffmpeg request: %w", err)
	}
	return req, req.Validate()
}

type LinkedFFmpegResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

type LinkedFFmpegMode string

const (
	LinkedFFmpegModeDirect LinkedFFmpegMode = "direct"
	LinkedFFmpegModeHidden LinkedFFmpegMode = "hidden"
)

type LinkedFFmpegRunner struct {
	Mode LinkedFFmpegMode
}

func NewLinkedFFmpegRunner(mode LinkedFFmpegMode) LinkedFFmpegRunner {
	if mode == "" {
		mode = LinkedFFmpegModeDirect
	}
	return LinkedFFmpegRunner{Mode: mode}
}

func DefaultLinkedFFmpegRunner() LinkedFFmpegRunner {
	return NewLinkedFFmpegRunner(LinkedFFmpegModeDirect)
}

func (r LinkedFFmpegRunner) FFmpeg(ctx context.Context, args ...string) (LinkedFFmpegResult, error) {
	return r.run(ctx, LinkedFFmpegRequest{Tool: LinkedFFmpegToolFFmpeg, Args: append([]string(nil), args...)})
}

func (r LinkedFFmpegRunner) FFprobe(ctx context.Context, args ...string) (LinkedFFmpegResult, error) {
	return r.run(ctx, LinkedFFmpegRequest{Tool: LinkedFFmpegToolFFprobe, Args: append([]string(nil), args...)})
}

func (r LinkedFFmpegRunner) run(ctx context.Context, req LinkedFFmpegRequest) (LinkedFFmpegResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := req.Validate(); err != nil {
		return LinkedFFmpegResult{}, err
	}

	switch r.Mode {
	case LinkedFFmpegModeHidden:
		return runLinkedFFmpegHidden(ctx, req)
	case LinkedFFmpegModeDirect, "":
		fallthrough
	default:
		return runLinkedFFmpegNative(ctx, req)
	}
}

func RunLinkedFFmpegDirect(ctx context.Context, args ...string) (LinkedFFmpegResult, error) {
	return DefaultLinkedFFmpegRunner().FFmpeg(ctx, args...)
}

func RunLinkedFFprobeDirect(ctx context.Context, args ...string) (LinkedFFmpegResult, error) {
	return DefaultLinkedFFmpegRunner().FFprobe(ctx, args...)
}

func RunLinkedFFmpegHidden(ctx context.Context, args ...string) (LinkedFFmpegResult, error) {
	return NewLinkedFFmpegRunner(LinkedFFmpegModeHidden).FFmpeg(ctx, args...)
}

func RunLinkedFFprobeHidden(ctx context.Context, args ...string) (LinkedFFmpegResult, error) {
	return NewLinkedFFmpegRunner(LinkedFFmpegModeHidden).FFprobe(ctx, args...)
}
