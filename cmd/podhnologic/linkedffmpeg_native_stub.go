//go:build !linkedffmpeg_cgo

package main

import (
	"context"
	"fmt"
)

func runLinkedFFmpegNative(ctx context.Context, req LinkedFFmpegRequest) (LinkedFFmpegResult, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return LinkedFFmpegResult{}, err
		}
	}

	return LinkedFFmpegResult{}, fmt.Errorf("%w: build with -tags linkedffmpeg_cgo", ErrLinkedFFmpegUnavailable)
}
