// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package log provides a slog.Handler integrated with [Cloud Logging].
//
// [Cloud Logging]: https://cloud.google.com/logging
package log

import (
	"context"
	"log/slog"
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/nil-go/sloth/gcp"
	"github.com/nil-go/sloth/otel"
	"github.com/nil-go/sloth/rate"
	"github.com/nil-go/sloth/sampling"
	"go.opentelemetry.io/otel/trace"
)

// Logger returns a slog.Logger which integrates with [Cloud Logging] and [Cloud Error Reporting].
//
// [Cloud Logging]: https://cloud.google.com/logging
// [Cloud Error Reporting]: https://cloud.google.com/error-reporting
func Logger(opts ...Option) *slog.Logger {
	option := options{}
	for _, opt := range opts {
		opt(&option)
	}
	// Get service and version from Google Cloud Run environment variables.
	if option.service == "" {
		option.service = os.Getenv("K_SERVICE")
	}
	if option.version == "" {
		option.version = os.Getenv("K_REVISION")
	}
	if option.project == "" {
		option.project, _ = metadata.ProjectIDWithContext(context.Background())
	}

	logOpts := append(
		[]gcp.Option{
			gcp.WithErrorReporting(option.service, option.version),
			gcp.WithTrace(option.project),
		},
		option.logOpts...,
	)

	handler := gcp.New(logOpts...)
	handler = rate.New(handler)
	handler = sampling.New(handler,
		func(ctx context.Context) bool {
			sc := trace.SpanContextFromContext(ctx)

			return !sc.IsValid() || sc.IsSampled()
		},
	)
	handler = otel.New(handler)

	return slog.New(handler)
}
