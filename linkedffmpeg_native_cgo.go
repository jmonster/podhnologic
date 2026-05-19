//go:build linkedffmpeg_cgo && cgo

package main

/*
#include <stdint.h>
#include <stdlib.h>

extern int podhnologic_linked_ffmpeg_main(const char *tool, int argc, const char **argv, int stdout_fd, int stderr_fd);
*/
import "C"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"unsafe"
)

func runLinkedFFmpegNative(ctx context.Context, req LinkedFFmpegRequest) (LinkedFFmpegResult, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return LinkedFFmpegResult{}, err
		}
	}

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return LinkedFFmpegResult{}, fmt.Errorf("create linked ffmpeg stdout pipe: %w", err)
	}
	defer stdoutR.Close()
	defer stdoutW.Close()

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		return LinkedFFmpegResult{}, fmt.Errorf("create linked ffmpeg stderr pipe: %w", err)
	}
	defer stderrR.Close()
	defer stderrW.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stdoutBuf, stdoutR)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stderrBuf, stderrR)
	}()

	cTool := C.CString(string(req.Tool))
	defer C.free(unsafe.Pointer(cTool))

	cArgs := make([]*C.char, len(req.Args))
	for i, arg := range req.Args {
		cArgs[i] = C.CString(arg)
		defer C.free(unsafe.Pointer(cArgs[i]))
	}

	var argPtr **C.char
	if len(cArgs) > 0 {
		argPtr = (**C.char)(unsafe.Pointer(&cArgs[0]))
	}

	exitCode := int(C.podhnologic_linked_ffmpeg_main(
		cTool,
		C.int(len(cArgs)),
		argPtr,
		C.int(stdoutW.Fd()),
		C.int(stderrW.Fd()),
	))

	_ = stdoutW.Close()
	_ = stderrW.Close()
	wg.Wait()

	result := LinkedFFmpegResult{
		Stdout:   append([]byte(nil), stdoutBuf.Bytes()...),
		Stderr:   append([]byte(nil), stderrBuf.Bytes()...),
		ExitCode: exitCode,
	}

	if exitCode != 0 {
		return result, fmt.Errorf("linked ffmpeg %s exited with code %d", req.Tool, exitCode)
	}

	return result, nil
}
