// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package profiler

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/profiler"
	"google.golang.org/api/option"
)

func Run(opts ...option.ClientOption) func(context.Context) error {
	return func(ctx context.Context) error {
		config := profiler.Config{}
		for _, opt := range opts {
			if f, ok := opt.(*optionFunc); ok {
				f.fn(&config)
			}
		}

		if err := profiler.Start(config, opts...); err != nil {
			return fmt.Errorf("start cloud profiling: %w", err)
		}
		slog.LogAttrs(ctx, slog.LevelInfo, "Cloud profiling has been initialized.")

		return nil
	}
}
