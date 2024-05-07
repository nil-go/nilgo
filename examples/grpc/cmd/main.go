// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package main

import (
	"embed"

	"cloud.google.com/go/compute/metadata"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"github.com/nil-go/nilgo"
	"github.com/nil-go/nilgo/config"
	"github.com/nil-go/nilgo/dev"
	"github.com/nil-go/nilgo/gcp"
	"github.com/nil-go/nilgo/gcp/profiler"
	ngrpc "github.com/nil-go/nilgo/grpc"
)

//go:embed config
var configFS embed.FS

func main() {
	var args []any
	switch {
	case metadata.OnGCE():
		opts, err := gcp.Args(
			gcp.WithLog(),
			gcp.WithTrace(),
			gcp.WithMetric(),
			gcp.WithProfiler(profiler.WithMutexProfiling()),
		)
		if err != nil {
			panic(err)
		}
		args = append(args, opts...)
	default:
		args = append(args, dev.Pprof)
	}
	args = append(args,
		config.WithFS(configFS),
		ngrpc.Run(
			ngrpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler())),
			ngrpc.WithConfigService(),
		),
	)

	if err := nilgo.Run(args...); err != nil {
		panic(err)
	}
}
