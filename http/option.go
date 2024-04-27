// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package http

import (
	"log/slog"
	"time"

	"github.com/nil-go/konf"
)

// WithTimeout provides the duration that timeout http request.
//
// By default, it has no timeout, which depends on the reverse proxy
// (e.g. API Gateway, Load Balancer, etc.) to cancel the request with timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(options *options) {
		options.timeout = timeout
	}
}

// WithLogHandler provides the slog.Handler for HTTP server logs.
//
// By default it uses handler from slog.Default().
func WithLogHandler(handler slog.Handler) Option {
	return func(options *options) {
		if handler != nil {
			options.handler = handler
		}
	}
}

// ConfigService registers the endpoint `_config/{path}` for config explanation.
//
// It uses the global konf.Config if the configs are not provided.
func ConfigService(configs ...*konf.Config) Option {
	return func(options *options) {
		if options.configs == nil {
			options.configs = []*konf.Config{}
		}
		options.configs = append(options.configs, configs...)
	}
}

type (
	// Option configures the http server.
	Option  func(*options)
	options struct {
		timeout time.Duration
		handler slog.Handler
		configs []*konf.Config
	}
)
