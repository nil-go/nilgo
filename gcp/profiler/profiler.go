// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package profiler enables the [Cloud Profiler] for the application.
//
// [Cloud Profiler]: https://cloud.google.com/profiler
package profiler

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/profiler"
	"google.golang.org/api/option"
)

// Run starts the [Cloud Profiler] with given options.
// It requires the following IAM roles:
//   - roles/cloudprofiler.agent
//
// [Cloud Profiler]: https://cloud.google.com/profiler
func Run(opts ...option.ClientOption) func(context.Context) error {
	return func(ctx context.Context) error {
		config := profiler.Config{}
		for _, opt := range opts {
			if f, ok := opt.(*optionFunc); ok {
				f.fn(&config)
			}
		}
		// Get service and version from Google Cloud Run environment variables.
		if config.Service == "" {
			config.Service = os.Getenv("K_SERVICE")
		}
		if config.ServiceVersion == "" {
			config.ServiceVersion = os.Getenv("K_REVISION")
		}
		if config.ProjectID == "" {
			config.ProjectID, _ = metadata.ProjectID()
		}

		if err := profiler.Start(config, opts...); err != nil {
			return fmt.Errorf("start cloud profiling: %w", err)
		}
		slog.LogAttrs(ctx, slog.LevelInfo, "Cloud profiling has been initialized.")

		return nil
	}
}
