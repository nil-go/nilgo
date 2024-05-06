// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package gcp provides customization for application runs on GCP.
//
// It integrates with following GCP services:
//   - [Cloud Logging]
//   - [Cloud Error Reporting]
//   - [Cloud Trace]
//   - [Cloud Monitoring]
//   - [Cloud Profiler]
//
// [Cloud Logging]: https://cloud.google.com/logging
// [Cloud Error Reporting]: https://cloud.google.com/error-reporting
// [Cloud Trace]: https://cloud.google.com/trace
// [Cloud Monitoring]: https://cloud.google.com/monitoring
// [Cloud Profiler]: https://cloud.google.com/profiler
package gcp

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/nil-go/sloth/gcp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/nil-go/nilgo/gcp/profiler"
)

// Options provides the [nilgo.Run] options for application runs on GCP.
//
// By default, only logging and error reporting are configured.
// Profiler need to enable explicitly
// using corresponding Option(s).
func Options(opts ...Option) ([]any, error) { //nolint:cyclop,funlen
	option := options{}
	for _, opt := range opts {
		opt(&option)
	}
	if option.project == "" {
		option.project, _ = metadata.ProjectID()
	}
	if option.service == "" {
		option.service = os.Getenv("K_SERVICE")
	}
	if option.version == "" {
		option.version = os.Getenv("K_REVISION")
	}

	if option.logOpts != nil {
		option.logOpts = append(option.logOpts, gcp.WithErrorReporting(option.service, option.version))
	}
	if option.project == "" {
		return []any{gcp.New(option.logOpts...)}, nil
	}

	if option.traceOpts != nil {
		option.logOpts = append(option.logOpts, gcp.WithTrace(option.project))
	}
	appOpts := []any{gcp.New(option.logOpts...)}

	res := resource.Default()
	ctx := context.Background()
	if option.traceOpts != nil {
		exporter, err := otlptracegrpc.New(ctx, option.traceOpts...)
		if err != nil {
			return nil, fmt.Errorf("create otlp trace exporter: %w", err)
		}
		appOpts = append(appOpts,
			trace.NewTracerProvider(
				trace.WithBatcher(exporter),
				trace.WithResource(res),
				trace.WithSampler(trace.ParentBased(trace.NeverSample())),
			),
		)
	}
	if option.metricOpts != nil {
		exporter, err := otlpmetricgrpc.New(ctx, option.metricOpts...)
		if err != nil {
			return nil, fmt.Errorf("create otlp metric exporter: %w", err)
		}

		appOpts = append(appOpts,
			metric.NewMeterProvider(
				metric.WithReader(metric.NewPeriodicReader(exporter)),
				metric.WithResource(res),
			),
		)
	}

	if option.profilerOpts != nil {
		option.profilerOpts = append(
			[]profiler.Option{
				profiler.WithProject(option.project),
				profiler.WithService(option.service),
				profiler.WithVersion(option.version),
			},
			option.profilerOpts...,
		)
		appOpts = append(appOpts, profiler.Run(option.profilerOpts...))
	}

	return appOpts, nil
}
