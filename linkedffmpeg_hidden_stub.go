//go:build !linkedffmpeg_hidden

package main

import (
	"context"
	"fmt"
)

func runLinkedFFmpegHidden(ctx context.Context, req LinkedFFmpegRequest) (LinkedFFmpegResult, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return LinkedFFmpegResult{}, err
		}
	}

	return LinkedFFmpegResult{}, fmt.Errorf("%w: build with -tags linkedffmpeg_hidden", ErrLinkedFFmpegUnavailable)
}
