// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

//nolint:ireturn
package grpc

import (
	"log/slog"

	"github.com/nil-go/konf"
	"google.golang.org/grpc"
)

// LogHandler provides the slog.Handler for gRPC logs.
//
// If the handler is not provided, it uses handler from slog.Default().
func LogHandler(handler slog.Handler) grpc.ServerOption {
	return serverOptionFunc{
		fn: func(options *serverOptions) {
			if handler != nil {
				options.handler = handler
			}
		},
	}
}

// ConfigService registers the pb.ConfigServiceServer implement to the gRPC server.
//
// It uses the global konf.Config if the configs are not provided.
func ConfigService(configs ...*konf.Config) grpc.ServerOption {
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
		grpcOpts []grpc.ServerOption
	}
)
