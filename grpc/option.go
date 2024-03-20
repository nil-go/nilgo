// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

//nolint:ireturn
package grpc

import (
	"log/slog"

	"github.com/nil-go/konf"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// WithTelemetry enables trace and metrics instruments on the gRPC server.
func WithTelemetry(opts ...otelgrpc.Option) grpc.ServerOption {
	return serverOptionFunc{
		fn: func(options *serverOptions) {
			if options.otelOpts == nil {
				options.otelOpts = []otelgrpc.Option{}
			}

			options.otelOpts = append(options.otelOpts, opts...)
		},
	}
}

// WithLogHandler provides the slog.Handler for gRPC logs.
//
// If the handler is not provided, it uses handler from slog.Default().
func WithLogHandler(handler slog.Handler) grpc.ServerOption {
	return serverOptionFunc{
		fn: func(options *serverOptions) {
			if handler != nil {
				options.handler = handler
			}
		},
	}
}

// WithConfigService registers the pb.ConfigServiceServer implement to the gRPC server.
//
// It uses the global konf.Config if the configs are not provided.
func WithConfigService(configs ...*konf.Config) grpc.ServerOption {
	return serverOptionFunc{
		fn: func(options *serverOptions) {
			if options.configs == nil {
				options.configs = []*konf.Config{}
			}
			options.configs = append(options.configs, configs...)
		},
	}
}

type (
	serverOptionFunc struct {
		grpc.EmptyServerOption
		fn func(*serverOptions)
	}
	serverOptions struct {
		handler  slog.Handler
		configs  []*konf.Config
		otelOpts []otelgrpc.Option
		grpcOpts []grpc.ServerOption
	}
)
