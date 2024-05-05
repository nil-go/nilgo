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
func Run(args ...any) error {
	// Setup configuration first so others can use it.
	var configOpts []config.Option
	for _, arg := range args {
		if opt, ok := arg.(config.Option); ok {
			configOpts = append(configOpts, opt)
		}
	}
	cfg, err := config.New(configOpts...)
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}
	konf.SetDefault(cfg)

	option := options{
		runOpts: []run.Option{
			run.WithPreRun(cfg.Watch),
		},
	}
	if err := option.apply(args); err != nil {
		return err
	}

	if option.traceProvider != nil {
		// Append log sampler at the beginning so it can be overridden by user.
		option.logOpts = append([]log.Option{log.WithSampler(traceSampler())}, option.logOpts...)
	}
	slog.SetDefault(log.New(option.logOpts...))
	slog.Info("Logger has been initialized.")

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
	logOpts       []log.Option
	runOpts       []run.Option
	runners       []func(context.Context) error
	traceProvider trace.TracerProvider
	meterProvider metric.MeterProvider
}

func (o *options) apply(args []any) error { //nolint:cyclop
	for _, arg := range args {
		switch opt := arg.(type) {
		case func() []any:
			if err := o.apply(opt()); err != nil {
				return err
			}
		case config.Option:
			// Already handled.
		case slog.Handler:
			o.logOpts = append(o.logOpts, log.WithHandler(opt))
		case log.Option:
			o.logOpts = append(o.logOpts, opt)
		case trace.TracerProvider:
			o.traceProvider = opt
			if provider, ok := opt.(interface {
				Shutdown(ctx context.Context) error
			}); ok {
				o.runOpts = append(o.runOpts, run.WithPostRun(provider.Shutdown))
			}
		case metric.MeterProvider:
			o.meterProvider = opt
			if provider, ok := opt.(interface {
				Shutdown(ctx context.Context) error
			}); ok {
				o.runOpts = append(o.runOpts, run.WithPostRun(provider.Shutdown))
			}
		case run.Option:
			o.runOpts = append(o.runOpts, opt)
		case func(context.Context) error:
			o.runners = append(o.runners, opt)
		default:
			return fmt.Errorf("unknown argument type: %T", opt) //nolint:err113
		}
	}

	return nil
}
