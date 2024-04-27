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

// WithAddress provides the address listened by the gRPC server.
// It should be either tcp address like `:8080` or unix socket address like `unix:nilgo.sock`.
//
// By default, it listens on `localhost:8080`  or `:${PORT}` if the environment variable exists.
func WithAddress(addresses ...string) Option {
	return func(options *options) {
		options.addresses = append(options.addresses, addresses...)
	}
}

// WithConfigService registers the pb.ConfigServiceServer implement to the gRPC server.
//
// It uses the global konf.Config if the configs are not provided.
func WithConfigService(configs ...*konf.Config) Option {
	return func(options *options) {
		if options.configs == nil {
			options.configs = []*konf.Config{}
		}
		options.configs = append(options.configs, configs...)
	}
}

type (
	serverOptionFunc struct {
		grpc.EmptyServerOption
		fn func(*serverOptions)
	}
	serverOptions struct {
		handler  slog.Handler
		grpcOpts []grpc.ServerOption
	}

	// Option configures the runner for the gRPC server.
	Option  func(*options)
	options struct {
		addresses []string
		configs   []*konf.Config
	}
)
