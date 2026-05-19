//go:build linkedffmpeg_hidden

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

func init() {
	if os.Getenv(linkedFFmpegBridgeEnv) == "" {
		return
	}

	code, err := handleLinkedFFmpegBridge(os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		if code == 0 {
			code = 1
		}
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(code)
}

func handleLinkedFFmpegBridge(stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	var req LinkedFFmpegRequest
	if err := json.NewDecoder(stdin).Decode(&req); err != nil {
		return 1, fmt.Errorf("decode linked ffmpeg bridge request: %w", err)
	}
	if err := req.Validate(); err != nil {
		return 1, err
	}

	result, err := runLinkedFFmpegNative(context.Background(), req)
	if len(result.Stdout) > 0 {
		if _, writeErr := stdout.Write(result.Stdout); writeErr != nil && err == nil {
			err = writeErr
		}
	}
	if len(result.Stderr) > 0 {
		if _, writeErr := stderr.Write(result.Stderr); writeErr != nil && err == nil {
			err = writeErr
		}
	}
	if err != nil {
		if result.ExitCode == 0 {
			result.ExitCode = 1
		}
		return result.ExitCode, err
	}

	return result.ExitCode, nil
}

func runLinkedFFmpegHidden(ctx context.Context, req LinkedFFmpegRequest) (LinkedFFmpegResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return LinkedFFmpegResult{}, err
	}

	exe, err := os.Executable()
	if err != nil {
		return LinkedFFmpegResult{}, fmt.Errorf("resolve linked ffmpeg executable: %w", err)
	}

	cmd := exec.CommandContext(ctx, exe)
	cmd.Env = append(os.Environ(), linkedFFmpegBridgeEnv+"=1")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return LinkedFFmpegResult{}, fmt.Errorf("open linked ffmpeg stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return LinkedFFmpegResult{}, fmt.Errorf("open linked ffmpeg stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		return LinkedFFmpegResult{}, fmt.Errorf("open linked ffmpeg stderr pipe: %w", err)
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stdoutBuf, stdout)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stderrBuf, stderr)
	}()

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return LinkedFFmpegResult{}, fmt.Errorf("start linked ffmpeg bridge: %w", err)
	}

	if payload, err := json.Marshal(req); err != nil {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return LinkedFFmpegResult{}, fmt.Errorf("marshal linked ffmpeg bridge request: %w", err)
	} else {
		if _, writeErr := io.Copy(stdin, bytes.NewReader(payload)); writeErr != nil {
			_ = stdin.Close()
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return LinkedFFmpegResult{}, fmt.Errorf("write linked ffmpeg bridge request: %w", writeErr)
		}
	}

	_ = stdin.Close()

	waitErr := cmd.Wait()
	wg.Wait()

	result := LinkedFFmpegResult{
		Stdout: append([]byte(nil), stdoutBuf.Bytes()...),
		Stderr: append([]byte(nil), stderrBuf.Bytes()...),
	}

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		return result, fmt.Errorf("linked ffmpeg hidden bridge failed: %w", waitErr)
	}

	return result, nil
}
