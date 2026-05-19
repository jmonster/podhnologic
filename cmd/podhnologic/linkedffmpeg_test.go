package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestLinkedFFmpegRequestRoundTrip(t *testing.T) {
	req := LinkedFFmpegRequest{
		Tool: LinkedFFmpegToolFFprobe,
		Args: []string{"-v", "quiet", "-print_format", "json", "-show_streams", "input.m4a"},
	}

	data, err := req.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	decoded, err := ParseLinkedFFmpegRequest(data)
	if err != nil {
		t.Fatalf("ParseLinkedFFmpegRequest failed: %v", err)
	}

	if decoded.Tool != req.Tool {
		t.Fatalf("decoded tool = %q, want %q", decoded.Tool, req.Tool)
	}

	if !reflect.DeepEqual(decoded.Args, req.Args) {
		t.Fatalf("decoded args = %#v, want %#v", decoded.Args, req.Args)
	}
}

func TestLinkedFFmpegDefaultRunnerUnavailable(t *testing.T) {
	result, err := RunLinkedFFmpegDirect(context.Background(), "-version")
	if err == nil {
		return
	}
	if !errors.Is(err, ErrLinkedFFmpegUnavailable) {
		t.Fatalf("RunLinkedFFmpegDirect error = %v, want ErrLinkedFFmpegUnavailable", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("default runner exit code = %d, want 0", result.ExitCode)
	}
}
