// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package gcp

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/profiler"
	"google.golang.org/api/option"
)

// WithProfiler enables Google [Cloud Profiler].
// It requires the following IAM roles:
//   - roles/cloudprofiler.agent
//
// [Cloud Profiler]: https://cloud.google.com/profiler
func WithProfiler(opts ...option.ClientOption) Option {
	return func(options *options) {
		if options.profilerOpts == nil {
			options.profilerOpts = []option.ClientOption{}
		}
		options.profilerOpts = append(options.profilerOpts, opts...)
	}
}

// WithMutextProfiling enables mutex profiling.
func WithMutextProfiling() Option {
	return func(options *options) {
		options.mutextProfiling = true
	}
}

type profilerOptions struct {
	profilerOpts    []option.ClientOption
	mutextProfiling bool
}

func profile(options *options) func(context.Context) error {
	if options.profilerOpts == nil {
		return nil
	}

	return func(ctx context.Context) error {
		slog.LogAttrs(ctx, slog.LevelInfo, "Start cloud profiler.", slog.String("project", options.project))
		if err := profiler.Start(profiler.Config{
			ProjectID:      options.project,
			Service:        options.service,
			ServiceVersion: options.version,
			MutexProfiling: options.mutextProfiling,
		}, options.profilerOpts...); err != nil {
			return fmt.Errorf("start cloud profiler: %w", err)
		}

		return nil
	}
}
