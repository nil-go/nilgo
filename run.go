// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package nilgo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nil-go/konf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/nil-go/nilgo/config"
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
// For now, it only can pass one of following types for args:
//   - config.Option
//   - log.Option
//   - run.Option
//   - func(context.Context) error
func Run(args ...any) error { //nolint:cyclop,funlen
	var (
		configOpts    []config.Option
		logOpts       []log.Option
		runOpts       []run.Option
		runners       []func(context.Context) error
		traceProvider trace.TracerProvider
		meterProvider metric.MeterProvider
	)
	for _, arg := range args {
		switch opt := arg.(type) {
		case config.Option:
			configOpts = append(configOpts, opt)
		case slog.Handler:
			logOpts = append(logOpts, log.WithHandler(opt))
		case log.Option:
			logOpts = append(logOpts, opt)
		case trace.TracerProvider:
			traceProvider = opt
			logOpts = append([]log.Option{
				log.WithSampler(func(ctx context.Context) bool {
					sc := trace.SpanContextFromContext(ctx)

					return !sc.IsValid() || sc.IsSampled()
				}),
			}, logOpts...)
		case metric.MeterProvider:
			meterProvider = opt
		case run.Option:
			runOpts = append(runOpts, opt)
		case func(context.Context) error:
			runners = append(runners, opt)
		default:
			return fmt.Errorf("unknown argument type: %T", opt) //nolint:err113
		}
	}

	logger := log.New(logOpts...)
	slog.SetDefault(logger)
	slog.Info("Logger has been initialized.")

	cfg, err := config.New(configOpts...)
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}
	konf.SetDefault(cfg)
	slog.Info("Config has been initialized.")

	if traceProvider != nil {
		otel.SetTracerProvider(traceProvider)
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			),
		)
		if provider, ok := traceProvider.(interface {
			Shutdown(ctx context.Context) error
		}); ok {
			runOpts = append(runOpts, run.WithPostRun(provider.Shutdown))
		}
		slog.Info("Trace provider has been initialized.")
	}

	if meterProvider != nil {
		otel.SetMeterProvider(meterProvider)
		if provider, ok := meterProvider.(interface {
			Shutdown(ctx context.Context) error
		}); ok {
			runOpts = append(runOpts, run.WithPostRun(provider.Shutdown))
		}
		slog.Info("Meter provider has been initialized.")
	}

	runner := run.New(append(runOpts, run.WithPreRun(cfg.Watch))...)
	if err := runner.Run(context.Background(), runners...); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}
