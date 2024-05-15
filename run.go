// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package nilgo provides a simple way to bootstrap an application.
package nilgo

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/nil-go/nilgo/log"
	"github.com/nil-go/nilgo/run"
)

// Run runs application with the given arguments synchronously.
//
// The runner passed in are running parallel without explicit order,
// which means it should not have temporal dependency between each other.
//
// The running can be interrupted if any runner returns non-nil error, or it receives an OS signal.
// It waits all runners return unless it's forcefully killed by OS.
//
// For now, it supports passing one of following types as args:
//   - log.Option
//   - run.Option
//   - func(context.Context) error
func Run(args ...any) error { //nolint:cyclop,funlen
	var option options
	for _, arg := range args {
		switch opt := arg.(type) {
		case log.Option:
			option.logOpts = append(option.logOpts, opt)
		case trace.TracerProvider:
			option.traceProvider = opt
			if provider, ok := opt.(interface {
				Shutdown(ctx context.Context) error
			}); ok {
				option.runOpts = append(option.runOpts, run.WithPostRun(provider.Shutdown))
			}
		case slog.Handler:
			option.logHandler = opt
		case metric.MeterProvider:
			option.meterProvider = opt
			if provider, ok := opt.(interface {
				Shutdown(ctx context.Context) error
			}); ok {
				option.runOpts = append(option.runOpts, run.WithPostRun(provider.Shutdown))
			}
		case run.Option:
			option.runOpts = append(option.runOpts, opt)
		case func(context.Context) error:
			option.runners = append(option.runners, opt)
		default:
			return fmt.Errorf("unknown argument type: %T", opt) //nolint:err113
		}
	}

	if option.logHandler != nil {
		if option.traceProvider != nil {
			// Append log sampler at the beginning so it can be overridden by user.
			option.logOpts = append([]log.Option{log.WithSampler(traceSampler())}, option.logOpts...)
		}
		slog.SetDefault(log.New(option.logHandler, option.logOpts...))
		slog.Info("Logger has been initialized.")
	}

	if option.traceProvider != nil {
		otel.SetTracerProvider(option.traceProvider)
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			),
		)
		slog.Info("Trace provider has been initialized.")
	}

	if option.meterProvider != nil {
		otel.SetMeterProvider(option.meterProvider)
		slog.Info("Meter provider has been initialized.")
	}

	runner := run.New(option.runOpts...)
	if err := runner.Run(context.Background(), option.runners...); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}

func traceSampler() func(ctx context.Context) bool {
	return func(ctx context.Context) bool {
		sc := trace.SpanContextFromContext(ctx)

		return !sc.IsValid() || sc.IsSampled()
	}
}

type options struct {
	logOpts []log.Option
	runOpts []run.Option
	runners []func(context.Context) error

	// Below are for internal usage. May switch to Provider in the future.
	logHandler    slog.Handler
	traceProvider trace.TracerProvider
	meterProvider metric.MeterProvider
}
