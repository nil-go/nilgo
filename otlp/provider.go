// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package otlp provides OpenTelemetry OTLP provider for trace and metric
package otlp

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// TraceProvider creates a new trace provider with OTLP exporter.
func TraceProvider(opts ...otlptracegrpc.Option) (*trace.TracerProvider, error) {
	opts = append([]otlptracegrpc.Option{otlptracegrpc.WithInsecure()}, opts...)
	exporter, err := otlptracegrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.Default()),
		trace.WithSampler(trace.ParentBased(trace.NeverSample())),
	), nil
}

// MeterProvider creates a new meter provider with OTLP exporter.
func MeterProvider(opts ...otlpmetricgrpc.Option) (*metric.MeterProvider, error) {
	opts = append([]otlpmetricgrpc.Option{otlpmetricgrpc.WithInsecure()}, opts...)
	exporter, err := otlpmetricgrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp metric exporter: %w", err)
	}

	return metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(resource.Default()),
	), nil
}
