// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package gcp provides customization for application runs on GCP.
package gcp

import (
	"context"
	"log/slog"

	"github.com/nil-go/sloth/gcp"
	"github.com/nil-go/sloth/otel"
	"github.com/nil-go/sloth/rate"
	"github.com/nil-go/sloth/sampling"
	"go.opentelemetry.io/otel/trace"

	"github.com/nil-go/nilgo/gcp/profiler"
)

// LogHandler returns a slog.Handler integrated with [Cloud Logging] and [Cloud Error Reporting].
//
// [Cloud Logging]: https://cloud.google.com/logging
// [Cloud Error Reporting]: https://cloud.google.com/error-reporting
func LogHandler(opts ...Option) slog.Handler {
	option := options{}
	option.apply(opts)
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

	return handler
}

// Profiler returns a function to start [Cloud Profiler].
// It requires the following IAM roles:
//   - roles/cloudprofiler.agent
//
// [Cloud Profiler]: https://cloud.google.com/profiler
func Profiler(opts ...Option) func(context.Context) error {
	option := options{}
	option.apply(opts)
	profilerOpts := append(
		[]profiler.Option{
			profiler.WithProject(option.project),
			profiler.WithService(option.service),
			profiler.WithVersion(option.version),
		},
		option.profilerOpts...,
	)

	return profiler.Run(profilerOpts...)
}
