// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package grpc

import (
	"slices"
	"strings"

	"github.com/nil-go/konf"
)

// WithAddress provides the address listened by the gRPC server.
// It should be either tcp address like `:8080` or unix socket address like `unix:nilgo.sock`.
//
// By default, it listens on `localhost:8080`  or `:${PORT}` if the environment variable exists.
func WithAddress(addresses ...string) Option {
	return func(options *options) {
		options.addresses = slices.Grow(options.addresses, len(addresses))
		for _, address := range addresses {
			network := "tcp"
			if strings.HasPrefix(address, "unix:") {
				network = "unix"
				address = strings.TrimPrefix(address[5:], "//")
			}
			options.addresses = append(options.addresses, socket{network: network, address: address})
		}
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
	// Option configures the runner for the gRPC server.
	Option  func(*options)
	options struct {
		addresses []socket
		configs   []*konf.Config
	}
	socket struct {
		network string
		address string
	}
)
