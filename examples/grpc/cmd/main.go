// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package main

import (
	"context"

	"cloud.google.com/go/compute/metadata"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"github.com/nil-go/nilgo"
	"github.com/nil-go/nilgo/dev"
	"github.com/nil-go/nilgo/gcp/log"
	"github.com/nil-go/nilgo/gcp/profiler"
	ngrpc "github.com/nil-go/nilgo/grpc"
	"github.com/nil-go/nilgo/otlp"
)

func main() {
	var (
		opts []nilgo.Option
		runs []func(context.Context) error
	)
	switch {
	case metadata.OnGCE():
		opts = append(opts, nilgo.WithLogHandler(log.Handler()))
		traceProvider, err := otlp.TraceProvider()
		if err != nil {
			panic(err)
		}
		opts = append(opts, nilgo.WithTraceProvider(traceProvider))
		meterProvider, err := otlp.MeterProvider()
		if err != nil {
			panic(err)
		}
		opts = append(opts, nilgo.WithMeterProvider(meterProvider))
		runs = append(runs, profiler.Run(profiler.WithMutexProfiling()))
	default:
		runs = append(runs, dev.Pprof)
	}

	runs = append(runs,
		ngrpc.Run(
			ngrpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler())),
			ngrpc.WithConfigService(),
		),
	)

	if err := nilgo.New(opts...).Run(context.Background(), runs...); err != nil {
		panic(err)
	}
}
